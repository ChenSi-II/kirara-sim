// Command import-weapons converts the per-weapon JSON files used by this fork
// into the static files expected by gcsim weapon packages.
//
// The upstream weapon pipeline consumes GenshinData's ExcelBinOutput files.
// The Weapons directory has a different schema, so it needs a small adapter
// instead of being passed to pipeline --weap directly.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var weaponClassFolders = map[string]string{
	"WEAPON_SWORD_ONE_HAND": "sword",
	"WEAPON_CLAYMORE":       "claymore",
	"WEAPON_POLE":           "spear",
	"WEAPON_CATALYST":       "catalyst",
	"WEAPON_BOW":            "bow",
}

type weaponProp struct {
	InitValue float64 `json:"InitValue"`
	PropType  string  `json:"PropType"`
	Curve     string  `json:"Type"`
}

type refinement struct {
	Name      string    `json:"Name"`
	Desc      string    `json:"Desc"`
	ParamList []float64 `json:"ParamList"`
}

type sourceWeapon struct {
	ID         int
	Name       string                        `json:"Name"`
	WeaponType string                        `json:"WeaponType"`
	WeaponProp []weaponProp                  `json:"WeaponProp"`
	Rarity     int                           `json:"Rarity"`
	Icon       string                        `json:"Icon"`
	Ascension  map[string]map[string]float64 `json:"Ascension"`
	Refinement map[string]refinement         `json:"Refinement"`
	Skin       bool                          `json:"Skin"`
	TPS        bool                          `json:"Tps"`
	Stats      map[string]json.RawMessage    `json:"StatsModifier"`
	XPUpper    map[string]float64            `json:"XPRequirements"`
	XPLower    map[string]float64            `json:"XpRequirements"`
}

func (w sourceWeapon) equipable() bool {
	_, validClass := weaponClassFolders[w.WeaponType]
	return validClass && !w.Skin && !w.TPS && len(w.WeaponProp) > 0 &&
		len(w.Stats) > 0 && len(w.Ascension) > 0 && len(w.Refinement) > 0
}

type weaponConfig struct {
	PackageName string
	Key         string
	GenshinID   int
	Path        string
}

type weaponSpec struct {
	ID          int    `json:"id"`
	PackageName string `json:"package_name"`
	Key         string `json:"key"`
	KeyConstant string `json:"key_constant"`
}

func main() {
	sourceRoot := flag.String("source", "Weapons", "directory containing <weapon-id>.json files")
	targetRoot := flag.String("target", filepath.Join("internal", "weapons"), "gcsim weapon package root")
	idsFlag := flag.String("ids", "", "comma-separated IDs to write (required with -write)")
	write := flag.Bool("write", false, "write data_gen.textproto and <package>_gen.go for selected IDs")
	scaffold := flag.Bool("scaffold", false, "create missing config.yml and passive stub for selected IDs")
	manifest := flag.String("manifest", filepath.Join("scripts", "import-weapons", "weapons.json"), "JSON manifest used by -scaffold")
	flag.Parse()

	selected, err := parseIDs(*idsFlag)
	check(err)
	if (*write || *scaffold) && len(selected) == 0 {
		check(errors.New("-write and -scaffold require an explicit, non-empty -ids list"))
	}

	sources, skipped, err := loadSources(*sourceRoot)
	check(err)
	if *scaffold {
		specs, err := loadSpecs(*manifest)
		check(err)
		check(scaffoldPackages(*targetRoot, sources, specs, selected))
	}
	configs, err := loadConfigs(*targetRoot)
	check(err)

	missing := make([]int, 0)
	for id := range sources {
		if _, ok := configs[id]; !ok {
			missing = append(missing, id)
		}
	}
	sort.Ints(missing)

	fmt.Printf("equipable=%d skipped=%d configured=%d missing=%d\n", len(sources), skipped, len(configs), len(missing))
	for _, id := range missing {
		w := sources[id]
		fmt.Printf("missing id=%d class=%s rarity=%d name=%q icon=%s\n", id, w.WeaponType, w.Rarity, w.Name, w.Icon)
	}

	if !*write {
		return
	}

	ids := make([]int, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		w, ok := sources[id]
		if !ok {
			check(fmt.Errorf("ID %d is absent or is not an equipable weapon", id))
		}
		cfg, ok := configs[id]
		if !ok {
			check(fmt.Errorf("ID %d has no config.yml under %s", id, *targetRoot))
		}
		wantFolder := weaponClassFolders[w.WeaponType]
		classFolder := filepath.Base(filepath.Dir(filepath.Dir(cfg.Path)))
		if classFolder != wantFolder {
			check(fmt.Errorf("ID %d config is under the wrong class directory: want %s", id, wantFolder))
		}
		check(writeGenerated(w, cfg))
		fmt.Printf("wrote id=%d key=%s package=%s\n", id, cfg.Key, cfg.PackageName)
	}
}

func loadSpecs(path string) (map[int]weaponSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []weaponSpec
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	result := make(map[int]weaponSpec, len(entries))
	for _, spec := range entries {
		if spec.ID <= 0 || spec.PackageName == "" || spec.Key == "" || spec.KeyConstant == "" {
			return nil, fmt.Errorf("invalid weapon manifest entry: %+v", spec)
		}
		if _, exists := result[spec.ID]; exists {
			return nil, fmt.Errorf("duplicate weapon ID %d in %s", spec.ID, path)
		}
		result[spec.ID] = spec
	}
	return result, nil
}

func scaffoldPackages(root string, sources map[int]sourceWeapon, specs map[int]weaponSpec, selected map[int]bool) error {
	ids := make([]int, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		w, ok := sources[id]
		if !ok {
			return fmt.Errorf("ID %d is absent or is not an equipable weapon", id)
		}
		spec, ok := specs[id]
		if !ok {
			return fmt.Errorf("ID %d is missing from the scaffold manifest", id)
		}
		dir := filepath.Join(root, weaponClassFolders[w.WeaponType], spec.PackageName)
		if _, err := os.Stat(dir); err == nil {
			return fmt.Errorf("refusing to scaffold existing directory %s", dir)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		config := fmt.Sprintf("package_name: %s\ngenshin_id: %d\nkey: %s\n", spec.PackageName, id, spec.Key)
		if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(config), 0o644); err != nil {
			return err
		}
		stub := renderPassiveStub(spec.PackageName, spec.KeyConstant)
		if err := os.WriteFile(filepath.Join(dir, spec.PackageName+".go"), stub, 0o644); err != nil {
			return err
		}
		fmt.Printf("scaffolded id=%d key=%s package=%s\n", id, spec.Key, spec.PackageName)
	}
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "import-weapons:", err)
		os.Exit(1)
	}
}

func parseIDs(raw string) (map[int]bool, error) {
	result := make(map[int]bool)
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}
	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid weapon ID %q", part)
		}
		result[id] = true
	}
	return result, nil
}

func loadSources(root string) (map[int]sourceWeapon, int, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, 0, err
	}
	weapons := make(map[int]sourceWeapon)
	skipped := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		if err != nil {
			return nil, 0, fmt.Errorf("parse ID from %s: %w", entry.Name(), err)
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, 0, err
		}
		var w sourceWeapon
		if err := json.Unmarshal(data, &w); err != nil {
			return nil, 0, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		w.ID = id
		if !w.equipable() {
			skipped++
			continue
		}
		if _, exists := weapons[id]; exists {
			return nil, 0, fmt.Errorf("duplicate weapon ID %d", id)
		}
		weapons[id] = w
	}
	return weapons, skipped, nil
}

func loadConfigs(root string) (map[int]weaponConfig, error) {
	configs := make(map[int]weaponConfig)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "config.yml" {
			return nil
		}
		cfg, err := parseConfig(path)
		if err != nil {
			return err
		}
		if previous, exists := configs[cfg.GenshinID]; exists {
			return fmt.Errorf("duplicate genshin_id %d in %s and %s", cfg.GenshinID, previous.Path, path)
		}
		configs[cfg.GenshinID] = cfg
		return nil
	})
	return configs, err
}

func parseConfig(path string) (weaponConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return weaponConfig{}, err
	}
	defer f.Close()

	cfg := weaponConfig{Path: path}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch strings.TrimSpace(parts[0]) {
		case "package_name":
			cfg.PackageName = value
		case "key":
			cfg.Key = value
		case "genshin_id":
			cfg.GenshinID, err = strconv.Atoi(value)
			if err != nil {
				return weaponConfig{}, fmt.Errorf("parse %s genshin_id: %w", path, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return weaponConfig{}, err
	}
	if cfg.PackageName == "" || cfg.Key == "" || cfg.GenshinID == 0 {
		return weaponConfig{}, fmt.Errorf("%s is missing package_name, key, or genshin_id", path)
	}
	return cfg, nil
}

func writeGenerated(w sourceWeapon, cfg weaponConfig) error {
	hash := existingNameHash(filepath.Join(filepath.Dir(cfg.Path), "data_gen.textproto"))
	if err := os.WriteFile(
		filepath.Join(filepath.Dir(cfg.Path), "data_gen.textproto"),
		renderTextProto(w, cfg, hash),
		0o644,
	); err != nil {
		return err
	}
	return os.WriteFile(
		filepath.Join(filepath.Dir(cfg.Path), cfg.PackageName+"_gen.go"),
		renderGeneratedGo(cfg.PackageName),
		0o644,
	)
}

func existingNameHash(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	const marker = "name_text_hash_map:"
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, marker) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, marker))
		hash, _ := strconv.ParseUint(value, 10, 64)
		return hash
	}
	return 0
}

func renderTextProto(w sourceWeapon, cfg weaponConfig, nameHash uint64) []byte {
	var out bytes.Buffer
	fmt.Fprintf(&out, "id: %d\n", w.ID)
	fmt.Fprintf(&out, "key: %q\n", cfg.Key)
	fmt.Fprintf(&out, "rarity: %d\n", w.Rarity)
	fmt.Fprintf(&out, "weapon_class: %s\n", w.WeaponType)
	fmt.Fprintf(&out, "image_name: %q\n", w.Icon)
	out.WriteString("base_stats: {\n")
	for _, prop := range w.WeaponProp {
		out.WriteString("   base_props: {\n")
		fmt.Fprintf(&out, "      prop_type: %s\n", prop.PropType)
		fmt.Fprintf(&out, "      initial_value: %s\n", formatFloat(prop.InitValue))
		fmt.Fprintf(&out, "      curve: %s\n", prop.Curve)
		out.WriteString("   }\n")
	}
	out.WriteString("   promo_data: {\n      max_level: 20\n   }\n")
	ascensionKeys := make([]int, 0, len(w.Ascension))
	for raw := range w.Ascension {
		level, err := strconv.Atoi(raw)
		if err == nil {
			ascensionKeys = append(ascensionKeys, level)
		}
	}
	sort.Ints(ascensionKeys)
	for _, ascension := range ascensionKeys {
		props := w.Ascension[strconv.Itoa(ascension)]
		baseATK, ok := findBaseAttack(props)
		if !ok {
			continue
		}
		fmt.Fprintf(&out, "   promo_data: {\n      max_level: %d\n", []int{0, 40, 50, 60, 70, 80, 90}[ascension])
		out.WriteString("      add_props: {\n         prop_type: FIGHT_PROP_BASE_ATTACK\n")
		fmt.Fprintf(&out, "         value: %s\n", formatFloat(baseATK))
		out.WriteString("      }\n   }\n")
	}
	out.WriteString("}\n")
	if nameHash != 0 {
		fmt.Fprintf(&out, "name_text_hash_map: %d\n", nameHash)
	}
	return out.Bytes()
}

func findBaseAttack(props map[string]float64) (float64, bool) {
	for name, value := range props {
		normalized := strings.ToLower(strings.ReplaceAll(name, "_", ""))
		if normalized == "fightpropbaseattack" {
			return value, true
		}
	}
	return 0, false
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func renderGeneratedGo(packageName string) []byte {
	return []byte(fmt.Sprintf(`// Code generated by "pipeline"; DO NOT EDIT.
package %s

import (
	_ "embed"

	"github.com/genshinsim/gcsim/pkg/model"
	"google.golang.org/protobuf/encoding/prototext"
)

//go:embed data_gen.textproto
var pbData []byte
var base *model.WeaponData

func init() {
	base = &model.WeaponData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
}

func (x *Weapon) Data() *model.WeaponData {
	return base
}
`, packageName))
}

func renderPassiveStub(packageName, keyConstant string) []byte {
	return []byte(fmt.Sprintf(`package %s

import (
	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

func init() {
	core.RegisterWeaponFunc(keys.%s, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	return w, nil
}
`, packageName, keyConstant))
}
