package disasterandremorse

import (
	"math"
	"testing"
)

func TestRefinementAndHexereiValues(t *testing.T) {
	if got := damageBonus(1, false); got != 0.40 {
		t.Fatalf("R1 bonus: got %v", got)
	}
	if got := damageBonus(5, false); got != 0.80 {
		t.Fatalf("R5 bonus: got %v", got)
	}
	if got := damageBonus(1, true); math.Abs(got-0.70) > 1e-9 {
		t.Fatalf("R1 Hexerei bonus: got %v", got)
	}
}
