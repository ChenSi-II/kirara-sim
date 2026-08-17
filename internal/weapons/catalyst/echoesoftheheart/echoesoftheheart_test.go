package echoesoftheheart

import "testing"

func TestRefinementValues(t *testing.T) {
	if got := emBonus(1); got != 60 {
		t.Fatalf("R1 EM bonus: got %v", got)
	}
	if got := emBonus(5); got != 120 {
		t.Fatalf("R5 EM bonus: got %v", got)
	}
}
