package forgedbythegoldenmelody

import "testing"

func TestRefinementValues(t *testing.T) {
	if atkBonus(1) != 0.18 || emBonus(1) != 120 {
		t.Fatalf("unexpected R1 values")
	}
	if atkBonus(5) != 0.36 || emBonus(5) != 240 {
		t.Fatalf("unexpected R5 values")
	}
}
