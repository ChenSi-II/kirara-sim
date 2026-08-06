package songofthevigil

import "testing"

func TestRefinementValues(t *testing.T) {
	if got := energyRestore(1); got != 4 {
		t.Fatalf("R1 energy: got %v", got)
	}
	if got := energyRestore(5); got != 8 {
		t.Fatalf("R5 energy: got %v", got)
	}
}
