package hereticsmoltenblade

import "testing"

func TestMovementAndRefinementValues(t *testing.T) {
	if got := movementBonus(1, 0); got != 0.18 {
		t.Fatalf("R1 minimum: got %v", got)
	}
	if got := movementBonus(1, 7); got != 0.36 {
		t.Fatalf("R1 maximum: got %v", got)
	}
	if got := movementBonus(5, 0); got != 0.36 {
		t.Fatalf("R5 minimum: got %v", got)
	}
	if got := movementBonus(5, 7); got != 0.72 {
		t.Fatalf("R5 maximum: got %v", got)
	}
}
