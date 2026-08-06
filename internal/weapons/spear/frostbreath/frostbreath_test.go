package frostbreath

import "testing"

func TestRefinementValues(t *testing.T) {
	if atkBonus(1) != 0.20 || teamEnergy(1) != 6 {
		t.Fatalf("unexpected R1 values")
	}
	if atkBonus(5) != 0.40 || teamEnergy(5) != 12 {
		t.Fatalf("unexpected R5 values")
	}
}
