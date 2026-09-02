package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/genshinsim/gcsim/pkg/model"
	"github.com/shizukayuki/excel-hk4e"
)

const communityWeaponSchemaVersion = 1

type communityManifest struct {
	SchemaVersion int    `json:"schema_version"`
	DataTrack     string `json:"data_track"`
}

type communityWeaponRecord struct {
	SchemaVersion int    `json:"schema_version"`
	Kind          string `json:"kind"`
	Identity      struct {
		ID        string `json:"id"`
		GCSimSlug string `json:"gcsim_slug"` //nolint:tagliatelle // fixed by the external community schema
	} `json:"identity"`
	GCSim struct {
		ConfigSeed struct {
			Kind       string `json:"kind"`
			Name       string `json:"name"`
			OverrideID string `json:"override_id"`
		} `json:"config_seed"`
	} `json:"gcsim"` //nolint:tagliatelle // fixed by the external community schema
	ImplementationInputs struct {
		Confirmed struct {
			Identity struct {
				Icon       string `json:"icon"`
				Name       string `json:"name"`
				Rarity     uint32 `json:"rarity"`
				WeaponType string `json:"weapon_type"`
			} `json:"identity"`
			BaseStats *struct {
				StatsModifier map[string]json.RawMessage `json:"stats_modifier"`
				WeaponProp    []communityWeaponProp      `json:"weapon_prop"`
			} `json:"base_stats"`
			Refinements map[string]communityRefinement `json:"refinements"`
		} `json:"confirmed"`
	} `json:"implementation_inputs"`
	Sources map[string]json.RawMessage `json:"sources"`
}

type communityWeaponProp struct {
	InitialValue float64 `json:"init_value"` //nolint:tagliatelle // fixed by the external community schema
	PropType     string  `json:"prop_type"`
	CurveType    string  `json:"type"` //nolint:tagliatelle // fixed by the external community schema
}

type communityRefinement struct {
	Name      string    `json:"name"`
	Desc      string    `json:"desc"`
	ParamList []float64 `json:"param_list"`
}

type communityLocalizedSource struct {
	Payload struct {
		Name string `json:"name"`
	} `json:"payload"`
}

func loadCommunityWeaponRef(root fs.FS, cfg *Config) (*weaponBuildRef, error) {
	return loadCommunityWeaponRefWithPromotion(root, cfg, func(rarity uint32, props []*model.WeaponProp) ([]*model.PromotionData, error) {
		promoteID, err := findCommunityPromotionTemplate(rarity, props)
		if err != nil {
			return nil, err
		}
		return buildPromotionData(promoteID)
	})
}

func loadCommunityWeaponRefWithPromotion(
	root fs.FS,
	cfg *Config,
	promotion func(uint32, []*model.WeaponProp) ([]*model.PromotionData, error),
) (*weaponBuildRef, error) {
	if !fs.ValidPath(cfg.Version) || strings.Contains(cfg.Version, "/") {
		return nil, fmt.Errorf("unsupported community version %q", cfg.Version)
	}

	base := path.Join("pipeline/community/data", cfg.Version)
	manifestData, err := fs.ReadFile(root, path.Join(base, "manifest.json"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("unsupported community version %q: manifest not found", cfg.Version)
		}
		return nil, fmt.Errorf("read community version %q manifest: %w", cfg.Version, err)
	}
	var manifest communityManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("decode community version %q manifest: %w", cfg.Version, err)
	}
	if manifest.SchemaVersion != communityWeaponSchemaVersion {
		return nil, fmt.Errorf(
			"unsupported community version %q manifest schema_version %d",
			cfg.Version,
			manifest.SchemaVersion,
		)
	}
	if manifest.DataTrack != "preview" {
		return nil, fmt.Errorf(
			"community version %q has data_track %q, expected preview",
			cfg.Version,
			manifest.DataTrack,
		)
	}
	if cfg.Override.Id == 0 {
		return nil, errors.New("community weapon source requires override.id")
	}

	filename := path.Join(base, "weapons", strconv.FormatUint(uint64(cfg.Override.Id), 10)+".json")
	data, err := fs.ReadFile(root, filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf(
				"community weapon %d not found in version %q at %s",
				cfg.Override.Id,
				cfg.Version,
				filename,
			)
		}
		return nil, fmt.Errorf("read community weapon %d: %w", cfg.Override.Id, err)
	}

	var record communityWeaponRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("decode community weapon %d: %w", cfg.Override.Id, err)
	}
	if err := validateCommunityWeaponRecord(&record, cfg); err != nil {
		return nil, fmt.Errorf("community weapon %d: %w", cfg.Override.Id, err)
	}
	return normalizeCommunityWeapon(&record, promotion)
}

func validateCommunityWeaponRecord(record *communityWeaponRecord, cfg *Config) error {
	if record.SchemaVersion != communityWeaponSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d", record.SchemaVersion)
	}
	if record.Kind != "weapon" {
		return fmt.Errorf("kind is %q, expected weapon", record.Kind)
	}
	if record.GCSim.ConfigSeed.Kind != "weapon" {
		return fmt.Errorf("gcsim.config_seed.kind is %q, expected weapon", record.GCSim.ConfigSeed.Kind)
	}

	identityID, err := strconv.ParseUint(record.Identity.ID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid identity.id %q: %w", record.Identity.ID, err)
	}
	if uint32(identityID) != cfg.Override.Id {
		return fmt.Errorf("identity.id %d does not match override.id %d", identityID, cfg.Override.Id)
	}
	seedID, err := strconv.ParseUint(record.GCSim.ConfigSeed.OverrideID, 10, 32)
	if err != nil {
		return fmt.Errorf(
			"invalid gcsim.config_seed.override_id %q: %w",
			record.GCSim.ConfigSeed.OverrideID,
			err,
		)
	}
	if uint32(seedID) != cfg.Override.Id {
		return fmt.Errorf(
			"gcsim.config_seed.override_id %d does not match override.id %d",
			seedID,
			cfg.Override.Id,
		)
	}
	if record.GCSim.ConfigSeed.Name != cfg.Name {
		return fmt.Errorf(
			"gcsim.config_seed.name %q does not match config name %q",
			record.GCSim.ConfigSeed.Name,
			cfg.Name,
		)
	}
	if record.Identity.GCSimSlug != cfg.Name {
		return fmt.Errorf(
			"identity.gcsim_slug %q does not match config name %q",
			record.Identity.GCSimSlug,
			cfg.Name,
		)
	}

	identity := record.ImplementationInputs.Confirmed.Identity
	if identity.Name == "" {
		return errors.New("implementation_inputs.confirmed.identity.name is missing")
	}
	if slug := excel.SlugLower(identity.Name); slug != cfg.Name {
		return fmt.Errorf("display name slug %q does not match config name %q", slug, cfg.Name)
	}
	if identity.Rarity == 0 {
		return errors.New("implementation_inputs.confirmed.identity.rarity is missing")
	}
	if ConvertEnum[model.WeaponType](identity.WeaponType, model.WeaponType_value, -1) == -1 {
		return fmt.Errorf("invalid weapon_type %q", identity.WeaponType)
	}

	baseStats := record.ImplementationInputs.Confirmed.BaseStats
	if baseStats == nil {
		return errors.New("implementation_inputs.confirmed.base_stats is missing")
	}
	if len(baseStats.WeaponProp) == 0 {
		return errors.New("implementation_inputs.confirmed.base_stats.weapon_prop is missing")
	}
	if len(baseStats.StatsModifier) == 0 {
		return errors.New("implementation_inputs.confirmed.base_stats.stats_modifier is missing")
	}

	refinements := record.ImplementationInputs.Confirmed.Refinements
	if len(refinements) == 0 {
		return errors.New("implementation_inputs.confirmed.refinements is missing")
	}
	for refine := 1; refine <= 5; refine++ {
		key := strconv.Itoa(refine)
		value, ok := refinements[key]
		if !ok {
			return fmt.Errorf("implementation_inputs.confirmed.refinements R%d is missing", refine)
		}
		if value.Name == "" || value.Desc == "" || len(value.ParamList) == 0 {
			return fmt.Errorf("implementation_inputs.confirmed.refinements R%d is incomplete", refine)
		}
	}
	return nil
}

func normalizeCommunityWeapon(
	record *communityWeaponRecord,
	promotion func(uint32, []*model.WeaponProp) ([]*model.PromotionData, error),
) (*weaponBuildRef, error) {
	identity := record.ImplementationInputs.Confirmed.Identity
	baseStats := record.ImplementationInputs.Confirmed.BaseStats
	modelData := &model.WeaponData{
		Id:          uint32(mustParseCommunityID(record.Identity.ID)),
		Key:         record.Identity.GCSimSlug,
		Rarity:      identity.Rarity,
		WeaponClass: ConvertEnum[model.WeaponType](identity.WeaponType, model.WeaponType_value, -1),
		ImageName:   communityAwakenIcon(identity.Icon),
		BaseStats:   &model.WeaponStatsData{},
	}
	for _, value := range baseStats.WeaponProp {
		prop, err := buildWeaponProp(value.InitialValue, value.PropType, value.CurveType)
		if err != nil {
			return nil, err
		}
		modelData.BaseStats.BaseProps = append(modelData.BaseStats.BaseProps, prop)
	}

	var err error
	modelData.BaseStats.PromoData, err = promotion(identity.Rarity, modelData.BaseStats.BaseProps)
	if err != nil {
		return nil, err
	}

	refinements := record.ImplementationInputs.Confirmed.Refinements
	first := refinements["1"]
	attr := &AttributeSpec{
		Type:  "passive",
		Name:  first.Name,
		Desc:  first.Desc,
		Index: NewTalentIndex(first.ParamList),
	}
	attr.SetValues(5, func(i int) []float64 {
		return refinements[strconv.Itoa(i+1)].ParamList
	})

	localization := communityWeaponLocalization(record, identity.Name)
	return &weaponBuildRef{
		Name:         identity.Name,
		Model:        modelData,
		Attributes:   []*AttributeSpec{attr},
		Localization: localization,
	}, nil
}

func findCommunityPromotionTemplate(rarity uint32, props []*model.WeaponProp) (uint32, error) {
	for _, candidate := range excel.WeaponExcelConfigData {
		if candidate.RankLevel != rarity {
			continue
		}
		var candidateProps []*model.WeaponProp
		valid := true
		for _, value := range candidate.WeaponProp {
			if value.InitValue == 0 || value.PropType == excel.FIGHT_PROP_NONE {
				continue
			}
			prop, err := buildWeaponProp(value.InitValue, value.PropType, value.Type)
			if err != nil {
				valid = false
				break
			}
			candidateProps = append(candidateProps, prop)
		}
		if valid && equalWeaponProps(candidateProps, props) {
			return candidate.WeaponPromoteId, nil
		}
	}
	return 0, fmt.Errorf(
		"no live promotion template matches rarity %d and community weapon_prop",
		rarity,
	)
}

func findCommunityWeaponVersions(root fs.FS, id uint32) ([]string, error) {
	if id == 0 {
		return nil, nil
	}
	entries, err := fs.ReadDir(root, "pipeline/community/data")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan community data versions: %w", err)
	}
	filename := strconv.FormatUint(uint64(id), 10) + ".json"
	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := path.Join("pipeline/community/data", entry.Name(), "weapons", filename)
		if _, err := fs.Stat(root, candidate); err == nil {
			versions = append(versions, entry.Name())
		}
	}
	slices.Sort(versions)
	return versions, nil
}

func compareLiveAndCommunityWeapon(live, community *weaponBuildRef) error {
	if live.Name != community.Name {
		return fmt.Errorf("display name differs: live=%q community=%q", live.Name, community.Name)
	}
	if live.Model.Id != community.Model.Id ||
		live.Model.Key != community.Model.Key ||
		live.Model.Rarity != community.Model.Rarity ||
		live.Model.WeaponClass != community.Model.WeaponClass ||
		live.Model.ImageName != community.Model.ImageName ||
		!reflect.DeepEqual(live.Model.BaseStats, community.Model.BaseStats) {
		return errors.New("identity, weapon type, rarity, icon, or base stats differ")
	}
	return nil
}

func equalWeaponProps(a, b []*model.WeaponProp) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].PropType != b[i].PropType ||
			a[i].InitialValue != b[i].InitialValue ||
			a[i].Curve != b[i].Curve {
			return false
		}
	}
	return true
}

func communityWeaponLocalization(record *communityWeaponRecord, fallback string) map[string]string {
	out := make(map[string]string, len(languages))
	for _, lang := range languages {
		out[lang] = fallback
	}

	raw, ok := record.Sources["nanoka"]
	if !ok {
		raw = record.Sources["lunaris"]
	}
	var localized map[string]json.RawMessage
	if err := json.Unmarshal(raw, &localized); err != nil {
		return out
	}
	langMap := map[string]string{
		"en": "EN",
		"zh": "CHS",
		"de": "DE",
		"es": "ES",
		"ja": "JP",
		"ko": "KR",
		"ru": "RU",
	}
	for sourceLang, pipelineLang := range langMap {
		var source communityLocalizedSource
		if err := json.Unmarshal(localized[sourceLang], &source); err == nil && source.Payload.Name != "" {
			out[pipelineLang] = source.Payload.Name
		}
	}
	return out
}

func communityAwakenIcon(icon string) string {
	if icon == "" || strings.HasSuffix(icon, "_Awaken") {
		return icon
	}
	return icon + "_Awaken"
}

func mustParseCommunityID(value string) uint64 {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		panic(err)
	}
	return id
}
