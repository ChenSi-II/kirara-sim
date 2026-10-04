package qiqi

import (
	"encoding/json"
	"os"
	"testing"
)

// Keep the source excerpt in testdata: origin_data is an ignored local cache.
func TestEnhancedTalentTablesMatchSource(t *testing.T) {
	data, err := os.ReadFile("testdata/enhanced_multipliers.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Skill []float64 `json:"skill_coordinated"`
		Burst []float64 `json:"burst_stellar"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		got, want []float64
	}{{"skillCoordinated", skillCoordinated[:], source.Skill}, {"burstStellar", burstStellar[:], source.Burst}} {
		if len(tc.want) != 15 || len(tc.got) != len(tc.want) {
			t.Fatalf("%s: source/implementation must cover all 15 levels", tc.name)
		}
		for i, want := range tc.want {
			if tc.got[i] != want {
				t.Errorf("%s level %d: got %v, want %v", tc.name, i+1, tc.got[i], want)
			}
		}
	}
}
