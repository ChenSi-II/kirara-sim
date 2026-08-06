package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"slices"
	"strconv"

	"github.com/genshinsim/gcsim/pkg/model"
	"github.com/shizukayuki/excel-hk4e"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type WeaponSpec struct {
	Name         string            `yaml:"name,omitempty"`
	Model        *model.WeaponData `yaml:"model,omitempty"`
	Localization map[string]string `yaml:"localization,omitempty"`
}

func (s *WeaponSpec) ClearRef() {}

type weaponBuildRef struct {
	Name         string
	Model        *model.WeaponData
	Attributes   []*AttributeSpec
	Localization map[string]string
}

var errLiveWeaponNotFound = errors.New("live weapon not found")

func buildWeaponSpec(cfg *Config) (*WeaponSpec, error) {
	ref, err := resolveWeaponBuildRef(projectRoot.FS(), cfg)
	if err != nil {
		return nil, err
	}
	cfg.Attributes = append(cfg.Attributes, ref.Attributes...)
	return &WeaponSpec{
		Name:         ref.Name,
		Model:        ref.Model,
		Localization: ref.Localization,
	}, nil
}

func resolveWeaponBuildRef(root fs.FS, cfg *Config) (*weaponBuildRef, error) {
	return resolveWeaponBuildRefWithLive(root, cfg, buildLiveWeaponRef)
}

func resolveWeaponBuildRefWithLive(
	root fs.FS,
	cfg *Config,
	live func(*Config) (*weaponBuildRef, error),
) (*weaponBuildRef, error) {
	return resolveWeaponBuildRefWithResolvers(root, cfg, live, loadCommunityWeaponRef)
}

func resolveWeaponBuildRefWithResolvers(
	root fs.FS,
	cfg *Config,
	live func(*Config) (*weaponBuildRef, error),
	community func(fs.FS, *Config) (*weaponBuildRef, error),
) (*weaponBuildRef, error) {
	switch cfg.Source {
	case "live":
		if cfg.Version != "" {
			return nil, errors.New("weapon version requires source: community")
		}
		return live(cfg)
	case "":
		if cfg.Version != "" {
			return nil, errors.New("weapon version requires source: community")
		}
		liveRef, liveErr := live(cfg)
		versions, err := findCommunityWeaponVersions(root, cfg.Override.Id)
		if err != nil {
			return nil, err
		}
		if liveErr == nil {
			if len(versions) == 0 {
				return liveRef, nil
			}
			if len(versions) > 1 {
				return nil, fmt.Errorf(
					"weapon %d exists in multiple community versions %v; select source and version explicitly",
					cfg.Override.Id,
					versions,
				)
			}
			communityCfg := *cfg
			communityCfg.Source = "community"
			communityCfg.Version = versions[0]
			communityRef, err := community(root, &communityCfg)
			if err != nil {
				return nil, err
			}
			if err := compareLiveAndCommunityWeapon(liveRef, communityRef); err != nil {
				return nil, fmt.Errorf("live/community conflict for weapon %d: %w", cfg.Override.Id, err)
			}
			return liveRef, nil
		}
		if !errors.Is(liveErr, errLiveWeaponNotFound) || len(versions) == 0 {
			return nil, liveErr
		}
		if len(versions) > 1 {
			return nil, fmt.Errorf(
				"weapon %d exists in multiple community versions %v; select source and version explicitly",
				cfg.Override.Id,
				versions,
			)
		}
		communityCfg := *cfg
		communityCfg.Source = "community"
		communityCfg.Version = versions[0]
		return community(root, &communityCfg)
	case "community":
		if cfg.Version == "" {
			return nil, errors.New("community weapon source requires version")
		}
		return community(root, cfg)
	default:
		return nil, fmt.Errorf("unsupported weapon source %q", cfg.Source)
	}
}

func buildLiveWeaponRef(cfg *Config) (*weaponBuildRef, error) {
	refs := excel.Filter(excel.WeaponExcelConfigData, func(v *excel.Weapon) bool {
		if id := cfg.Override.Id; id != 0 {
			return v.Id == id
		}
		return v.StoryId != 0 && excel.SlugLower(v.Name()) == cfg.Name
	})
	if len(refs) != 1 {
		if len(refs) == 0 {
			return nil, fmt.Errorf("%w: query results in refs=0 but we expect 1", errLiveWeaponNotFound)
		}
		return nil, fmt.Errorf("query results in refs=%v but we expect 1", len(refs))
	}
	ref := refs[0]

	promoData, err := buildPromotionData(ref.WeaponPromoteId)
	if err != nil {
		return nil, fmt.Errorf("weapon_id=%v: %w", ref.Id, err)
	}

	name := ref.Name()
	out := &weaponBuildRef{
		Name: name,
		Model: &model.WeaponData{
			Id:          ref.Id,
			Key:         excel.SlugLower(name),
			Rarity:      ref.RankLevel,
			WeaponClass: ConvertEnum[model.WeaponType](ref.WeaponType, model.WeaponType_value, -1),
			ImageName:   ref.AwakenIcon,
			BaseStats: &model.WeaponStatsData{
				PromoData: promoData,
			},
		},
		Localization: make(map[string]string),
	}
	if out.Model.WeaponClass == -1 {
		return nil, fmt.Errorf("unknown weapon_type=%v", ref.WeaponType)
	}

	for _, add := range ref.WeaponProp {
		if add.InitValue == 0 || add.PropType == excel.FIGHT_PROP_NONE {
			continue
		}
		prop, err := buildWeaponProp(add.InitValue, add.PropType, add.Type)
		if err != nil {
			return nil, err
		}
		out.Model.BaseStats.BaseProps = append(out.Model.BaseStats.BaseProps, prop)
	}

	for num, id := range ref.SkillAffix {
		affixes := excel.Filter(excel.EquipAffixExcelConfigData, func(v *excel.EquipAffix) bool { return v.Id == id })
		slices.SortFunc(affixes, func(a, b *excel.EquipAffix) int { return cmp.Compare(a.Level, b.Level) })
		if len(affixes) == 0 {
			continue
		}
		attr := &AttributeSpec{
			Type:   "passive",
			Name:   affixes[0].NameTextMapHash.String(),
			Desc:   affixes[0].DescTextMapHash.String(),
			Config: affixes[0].OpenConfig,
			Index:  NewTalentIndex(affixes[0].ParamList),
		}
		if num > 0 {
			attr.Type += strconv.Itoa(num)
		}
		attr.SetValues(len(affixes), func(i int) []float64 { return affixes[i].ParamList })
		out.Attributes = append(out.Attributes, attr)
	}

	for _, lang := range languages {
		out.Localization[lang] = ref.NameTextMapHash.Lang(lang)
	}
	return out, nil
}

func buildWeaponProp(initialValue float64, propType, curveType any) (*model.WeaponProp, error) {
	if !slices.ContainsFunc(curveTypes[KindWeapon], func(curve excel.GrowCurveType) bool {
		return curve.String() == fmt.Sprint(curveType)
	}) {
		return nil, fmt.Errorf("curve not listed in known types: %v", curveType)
	}
	typ := ConvertEnum[model.FightPropType](propType, model.FightPropType_value, -1)
	curve := ConvertEnum[model.GrowCurveType](curveType, model.GrowCurveType_value, -1)
	if typ == -1 {
		return nil, fmt.Errorf("unknown prop=%v", propType)
	}
	if curve == -1 {
		return nil, fmt.Errorf("unknown curve=%v", curveType)
	}
	return &model.WeaponProp{
		PropType:     typ,
		InitialValue: initialValue,
		Curve:        curve,
	}, nil
}

func buildPromotionData(promoteID uint32) ([]*model.PromotionData, error) {
	promote := excel.Filter(excel.WeaponPromoteExcelConfigData, func(v *excel.WeaponPromote) bool {
		return v.WeaponPromoteId == promoteID
	})
	if len(promote) == 0 {
		return nil, fmt.Errorf("no promote found for weapon_promote_id=%v", promoteID)
	}
	slices.SortFunc(promote, func(a, b *excel.WeaponPromote) int {
		return cmp.Compare(a.PromoteLevel, b.PromoteLevel)
	})

	out := make([]*model.PromotionData, 0, len(promote))
	for _, v := range promote {
		props, err := ConvertAddProps(v.AddProps)
		if err != nil {
			return nil, err
		}
		out = append(out, &model.PromotionData{
			MaxLevel: v.UnlockMaxLevel,
			AddProps: props,
		})
	}
	return out, nil
}

func (c *Compiled) GenerateWeapons() error {
	kind := KindWeapon
	inputs := excel.Filter(c.Configuration, func(v *Config) bool { return v.Kind == kind })

	imports := ImportTmpl{Kind: kind}
	keys := KeysTmpl{Kind: kind}
	assets := AssetsTmpl{Kind: kind, Variable: "weaponMap"}
	shortcut := ShortcutTmpl{Kind: kind, Variable: "WeaponNameToKey", Type: "keys.Weapon"}
	catalog := CatalogTmpl{Kind: kind, Variable: "WeaponMap", Type: "keys.Weapon", ModelName: reflect.TypeFor[*model.WeaponData]().String()}
	doc := DocTmpl{Kind: kind}
	issues := IssuesTmpl{Kind: kind}
	fields := FieldsTmpl{Kind: kind}
	hitbox := HitboxTmpl{Kind: kind}
	hitlag := HitlagTmpl{Kind: kind}
	params := ParamsTmpl{Kind: kind}
	defer imports.Write()
	defer keys.Write()
	defer assets.Write()
	defer shortcut.Write()
	defer catalog.Write()
	defer doc.Write()
	defer issues.Write()
	defer fields.Write()
	defer hitbox.Write()
	defer hitlag.Write()
	defer params.Write()

	models := &model.WeaponDataMap{Data: make(map[string]*model.WeaponData)}
	for _, config := range inputs {
		spec := config.Weapon

		imports.PackagePath = append(imports.PackagePath, config.Dir())
		keys.Name = append(keys.Name, spec.Name)
		shortcut.Slug = append(shortcut.Slug, excel.Slug(spec.Name))
		shortcut.Names = append(shortcut.Names, append([]string{spec.Model.Key}, config.Shortcuts...))

		doc.Key = append(doc.Key, spec.Model.Key)
		doc.Name = append(doc.Name, spec.Name)
		issues.Key = append(issues.Key, spec.Model.Key)
		issues.Data = append(issues.Data, config.Docs.Issues)
		fields.Key = append(fields.Key, spec.Model.Key)
		fields.Data = append(fields.Data, config.Docs.Fields)
		hitbox.Key = append(hitbox.Key, spec.Model.Key)
		hitbox.Data = append(hitbox.Data, config.Abilities)
		hitlag.Key = append(hitlag.Key, spec.Model.Key)
		hitlag.Data = append(hitlag.Data, config.Abilities)
		params.Key = append(params.Key, spec.Model.Key)
		params.Data = append(params.Data, config.Abilities)

		assets.Key = append(assets.Key, spec.Model.Key)
		assets.Image = append(assets.Image, spec.Model.ImageName)

		catalog.Slug = append(catalog.Slug, excel.Slug(spec.Name))
		catalog.Model = append(catalog.Model, proto.Clone(spec.Model))

		m := proto.Clone(spec.Model).(*model.WeaponData)
		m.BaseStats = nil
		models.Data[m.Key] = m

		b := bytes.NewBuffer(nil)
		for _, attr := range config.Attributes {
			b.WriteString(attr.EmitDesc("// "))
		}
		fmt.Fprintf(b, "package %s\n", path.Base(config.Dir()))
		b.WriteString("import (\n")
		for _, pkg := range []string{
			path.Join(baseModule, "pkg/core"),
			path.Join(baseModule, "pkg/core/keys"),
		} {
			fmt.Fprintf(b, "\t\"%s\"\n", pkg)
		}
		b.WriteString(")\n")

		b.WriteString("func init() {\n")
		fmt.Fprintf(b, "core.Register%[2]sFunc(keys.%[1]s, New%[2]s)\n", excel.Slug(spec.Name), "Weapon")
		b.WriteString("}\n")

		if err := emitTalents(b, config.Talents, config.Attributes); err != nil {
			return fmt.Errorf("%v: %w", config.Path, err)
		}
		writeFile(path.Join(config.Dir(), fmt.Sprintf("zz_%s.dm.go", spec.Model.Key)), b.Bytes())
	}

	data, err := protojson.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal %v models: %w", kind, err)
	}
	writeFile(fmt.Sprintf("ui/packages/ui/src/Data/%s.dm.json", kind), data)

	return nil
}
