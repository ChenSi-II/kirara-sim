package venti

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
		Hurricane []float64 `json:"hurricane_multiplier"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.Hurricane) != 15 || len(hurricaneMultiplier) != len(source.Hurricane) {
		t.Fatal("source/implementation must cover all 15 levels")
	}
	for i, want := range source.Hurricane {
		if got := hurricaneMultiplier[i]; got != want {
			t.Errorf("hurricaneMultiplier level %d: got %v, want %v", i+1, got, want)
		}
	}
}
