package ateaspoonoftranscendence

import "testing"

func TestRefinementValues(t *testing.T) {
	if got := atkBonus(1); got != 0.28 {
		t.Fatalf("R1 ATK bonus: got %v", got)
	}
	if got := atkBonus(5); got != 0.56 {
		t.Fatalf("R5 ATK bonus: got %v", got)
	}
}
