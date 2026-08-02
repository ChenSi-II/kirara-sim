package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEquipableFiltering(t *testing.T) {
	base := sourceWeapon{
		WeaponType: "WEAPON_BOW",
		WeaponProp: []weaponProp{{PropType: "FIGHT_PROP_BASE_ATTACK"}},
		Stats:      map[string]json.RawMessage{"ATK": nil},
		Ascension: map[string]map[string]float64{
			"1": {"FIGHT_PROP_BASE_ATTACK": 31.1},
		},
		Refinement: map[string]refinement{"1": {}},
	}
	if !base.equipable() {
		t.Fatal("regular weapon was rejected")
	}

	tests := []struct {
		name string
		edit func(*sourceWeapon)
	}{
		{"skin", func(w *sourceWeapon) { w.Skin = true }},
		{"tps", func(w *sourceWeapon) { w.TPS = true }},
		{"invalid class", func(w *sourceWeapon) { w.WeaponType = "ITEM_TPS_WEAPON" }},
		{"missing props", func(w *sourceWeapon) { w.WeaponProp = nil }},
		{"missing stats", func(w *sourceWeapon) { w.Stats = nil }},
		{"missing ascension", func(w *sourceWeapon) { w.Ascension = nil }},
		{"missing refinement", func(w *sourceWeapon) { w.Refinement = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := base
			tc.edit(&got)
			if got.equipable() {
				t.Fatal("non-equipable record was accepted")
			}
		})
	}
}

func TestRenderTextProtoNormalizesAscensionKey(t *testing.T) {
	w := sourceWeapon{
		ID:         15516,
		WeaponType: "WEAPON_BOW",
		Rarity:     5,
		Icon:       "UI_EquipIcon_Bow_Alkonost",
		WeaponProp: []weaponProp{
			{InitValue: 44.3358, PropType: "FIGHT_PROP_BASE_ATTACK", Curve: "GROW_CURVE_ATTACK_304"},
			{InitValue: 0.192, PropType: "FIGHT_PROP_CRITICAL_HURT", Curve: "GROW_CURVE_CRITICAL_301"},
		},
		Ascension: map[string]map[string]float64{
			"1": {"FightPropBaseAttack": 31.1},
			"2": {"FIGHT_PROP_BASE_ATTACK": 62.2},
		},
	}
	cfg := weaponConfig{PackageName: "goldenfrostboundoath", Key: "goldenfrostboundoath", GenshinID: 15516}
	got := string(renderTextProto(w, cfg, 1234))
	for _, want := range []string{
		"id: 15516",
		"initial_value: 44.3358",
		"max_level: 40",
		"value: 31.1",
		"max_level: 50",
		"value: 62.2",
		"name_text_hash_map: 1234",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered textproto does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "package_name:") || strings.Contains(got, "genshin_id:") {
		t.Fatalf("rendered textproto contains config-only metadata:\n%s", got)
	}
}

func TestParseIDs(t *testing.T) {
	got, err := parseIDs("11435, 15516,11435")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !got[11435] || !got[15516] {
		t.Fatalf("unexpected IDs: %#v", got)
	}
	if _, err := parseIDs("not-an-id"); err == nil {
		t.Fatal("invalid ID was accepted")
	}
}

func TestScaffoldPackages(t *testing.T) {
	root := t.TempDir()
	sources := map[int]sourceWeapon{
		15516: {ID: 15516, WeaponType: "WEAPON_BOW"},
	}
	specs := map[int]weaponSpec{
		15516: {
			ID:          15516,
			PackageName: "goldenfrostboundoath",
			Key:         "goldenfrostboundoath",
			KeyConstant: "GoldenFrostboundOath",
		},
	}
	selected := map[int]bool{15516: true}
	if err := scaffoldPackages(root, sources, specs, selected); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(root, "bow", "goldenfrostboundoath")
	config, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(config); !strings.Contains(got, "genshin_id: 15516") || !strings.Contains(got, "key: goldenfrostboundoath") {
		t.Fatalf("unexpected config:\n%s", got)
	}
	stub, err := os.ReadFile(filepath.Join(dir, "goldenfrostboundoath.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stub), "keys.GoldenFrostboundOath") {
		t.Fatalf("stub does not register expected key:\n%s", stub)
	}

	if err := scaffoldPackages(root, sources, specs, selected); err == nil || !strings.Contains(err.Error(), "refusing to scaffold existing directory") {
		t.Fatalf("existing directory was not rejected: %v", err)
	}
}
