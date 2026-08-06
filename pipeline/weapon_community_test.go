package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/genshinsim/gcsim/pkg/model"
)

func communityTestPromotion(uint32, []*model.WeaponProp) ([]*model.PromotionData, error) {
	return []*model.PromotionData{{MaxLevel: 20}}, nil
}

func TestCommunityWeaponsLoad(t *testing.T) {
	root := os.DirFS("..")
	for _, tc := range []struct {
		id   uint32
		name string
	}{
		{id: 15435, name: "jadevista"},
		{id: 15436, name: "covenantoffrostandsnow"},
	} {
		t.Run(strconv.FormatUint(uint64(tc.id), 10), func(t *testing.T) {
			cfg := &Config{
				Kind:    KindWeapon,
				Name:    tc.name,
				Source:  "community",
				Version: "6.7.54",
				Override: Override{
					Id: tc.id,
				},
			}
			ref, err := loadCommunityWeaponRefWithPromotion(root, cfg, communityTestPromotion)
			if err != nil {
				t.Fatal(err)
			}
			if ref.Model.Id != tc.id || ref.Model.Key != tc.name {
				t.Fatalf("unexpected model identity: %+v", ref.Model)
			}
			if ref.Model.WeaponClass != model.WeaponType_WEAPON_BOW {
				t.Fatalf("unexpected weapon type: %v", ref.Model.WeaponClass)
			}
			if len(ref.Model.BaseStats.BaseProps) != 2 {
				t.Fatalf("unexpected base props: %+v", ref.Model.BaseStats.BaseProps)
			}
			if len(ref.Attributes) != 1 || len(ref.Attributes[0].Values) == 0 {
				t.Fatalf("refinement attributes were not normalized: %+v", ref.Attributes)
			}
			if ref.Localization["EN"] == "" || ref.Localization["CHS"] == "" {
				t.Fatalf("community localization is missing: %+v", ref.Localization)
			}
		})
	}
}

func TestEveryCommunityWeaponRecordLoads(t *testing.T) {
	entries, err := os.ReadDir("community/data/6.7.54/weapons")
	if err != nil {
		t.Fatal(err)
	}
	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile("community/data/6.7.54/weapons/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		var record communityWeaponRecord
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if record.Kind != "weapon" {
			continue
		}
		id, err := strconv.ParseUint(record.Identity.ID, 10, 32)
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		cfg := communityConfig(uint32(id), record.GCSim.ConfigSeed.Name)
		ref, err := loadCommunityWeaponRefWithPromotion(
			os.DirFS(".."),
			cfg,
			communityTestPromotion,
		)
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if ref.Model.Id != uint32(id) || ref.Model.Key != cfg.Name {
			t.Fatalf("%s: unexpected normalized identity %+v", entry.Name(), ref.Model)
		}
		if len(ref.Attributes) != 1 || len(ref.Attributes[0].Values) == 0 {
			t.Fatalf("%s: refinements were not normalized", entry.Name())
		}
		loaded++
	}
	if loaded != 16 {
		t.Fatalf("loaded %d weapon records, want 16", loaded)
	}
}

func TestLiveWeapon11519StillUsesLiveResolver(t *testing.T) {
	want := &weaponBuildRef{
		Name:  "Lightbearing Moonshard",
		Model: &model.WeaponData{Id: 11519, Key: "lightbearingmoonshard"},
	}
	called := false
	got, err := resolveWeaponBuildRefWithLive(fstest.MapFS{}, &Config{
		Kind: KindWeapon,
		Name: "lightbearingmoonshard",
		Override: Override{
			Id: 11519,
		},
	}, func(cfg *Config) (*weaponBuildRef, error) {
		called = true
		if cfg.Override.Id != 11519 {
			t.Fatalf("live resolver received ID %d", cfg.Override.Id)
		}
		return want, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called || got != want {
		t.Fatal("default source did not preserve the live resolver")
	}
}

func TestDefaultSourceDoesNotFallBackToCommunity(t *testing.T) {
	liveErr := errors.New("live refs=0")
	_, err := resolveWeaponBuildRefWithLive(os.DirFS(".."), &Config{
		Kind: KindWeapon,
		Name: "jadevista",
		Override: Override{
			Id: 15435,
		},
	}, func(*Config) (*weaponBuildRef, error) {
		return nil, liveErr
	})
	if !errors.Is(err, liveErr) {
		t.Fatalf("default source silently used community data: %v", err)
	}
}

func TestMissingLiveWeaponFallsBackToCommunity(t *testing.T) {
	want := &weaponBuildRef{
		Name:  "Jade Vista",
		Model: &model.WeaponData{Id: 15435, Key: "jadevista"},
	}
	called := false
	got, err := resolveWeaponBuildRefWithResolvers(
		os.DirFS(".."),
		&Config{
			Kind: KindWeapon,
			Name: "jadevista",
			Override: Override{
				Id: 15435,
			},
		},
		func(*Config) (*weaponBuildRef, error) {
			return nil, errLiveWeaponNotFound
		},
		func(_ fs.FS, cfg *Config) (*weaponBuildRef, error) {
			called = true
			if cfg.Source != "community" || cfg.Version != "6.7.54" {
				t.Fatalf("unexpected fallback config: %+v", cfg)
			}
			return want, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !called || got != want {
		t.Fatal("missing live weapon did not use the community fallback")
	}
}

func TestCommunityWeaponMissingFile(t *testing.T) {
	root := communityFixtureFS(t, 15435, nil)
	cfg := communityConfig(15436, "covenantoffrostandsnow")
	_, err := loadCommunityWeaponRefWithPromotion(root, cfg, communityTestPromotion)
	if err == nil || !strings.Contains(err.Error(), "community weapon 15436 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommunityWeaponValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*communityWeaponRecord)
		wantErr string
	}{
		{
			name: "identity ID mismatch",
			mutate: func(record *communityWeaponRecord) {
				record.Identity.ID = "15436"
			},
			wantErr: "identity.id 15436 does not match override.id 15435",
		},
		{
			name: "config seed override ID mismatch",
			mutate: func(record *communityWeaponRecord) {
				record.GCSim.ConfigSeed.OverrideID = "15436"
			},
			wantErr: "gcsim.config_seed.override_id 15436 does not match override.id 15435",
		},
		{
			name: "slug mismatch",
			mutate: func(record *communityWeaponRecord) {
				record.Identity.GCSimSlug = "wrongslug"
			},
			wantErr: "identity.gcsim_slug",
		},
		{
			name: "name mismatch",
			mutate: func(record *communityWeaponRecord) {
				record.GCSim.ConfigSeed.Name = "wrongname"
			},
			wantErr: "gcsim.config_seed.name",
		},
		{
			name: "missing base stats",
			mutate: func(record *communityWeaponRecord) {
				record.ImplementationInputs.Confirmed.BaseStats = nil
			},
			wantErr: "base_stats is missing",
		},
		{
			name: "missing refinements",
			mutate: func(record *communityWeaponRecord) {
				record.ImplementationInputs.Confirmed.Refinements = nil
			},
			wantErr: "refinements is missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			record := readCommunityRecord(t, 15435)
			tc.mutate(record)
			root := communityFixtureFS(t, 15435, record)
			_, err := loadCommunityWeaponRefWithPromotion(
				root,
				communityConfig(15435, "jadevista"),
				communityTestPromotion,
			)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnsupportedCommunityVersion(t *testing.T) {
	cfg := communityConfig(15435, "jadevista")
	cfg.Version = "9.9.9"
	_, err := loadCommunityWeaponRefWithPromotion(fstest.MapFS{}, cfg, communityTestPromotion)
	if err == nil || !strings.Contains(err.Error(), `unsupported community version "9.9.9"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func communityConfig(id uint32, name string) *Config {
	return &Config{
		Kind:    KindWeapon,
		Name:    name,
		Source:  "community",
		Version: "6.7.54",
		Override: Override{
			Id: id,
		},
	}
}

func readCommunityRecord(t *testing.T, id uint32) *communityWeaponRecord {
	t.Helper()
	data, err := os.ReadFile(
		"community/data/6.7.54/weapons/" + strconv.FormatUint(uint64(id), 10) + ".json",
	)
	if err != nil {
		t.Fatal(err)
	}
	var record communityWeaponRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return &record
}

func communityFixtureFS(
	t *testing.T,
	id uint32,
	record *communityWeaponRecord,
) fs.FS {
	t.Helper()
	files := fstest.MapFS{
		"pipeline/community/data/6.7.54/manifest.json": &fstest.MapFile{
			Data: []byte(`{"schema_version":1,"data_track":"preview"}`),
		},
	}
	if record != nil {
		data, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		files["pipeline/community/data/6.7.54/weapons/"+
			strconv.FormatUint(uint64(id), 10)+".json"] = &fstest.MapFile{Data: data}
	}
	return files
}
