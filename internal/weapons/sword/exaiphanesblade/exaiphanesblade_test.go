package exaiphanesblade

import "testing"

func TestRefinementValues(t *testing.T) {
	if atkByRefine[1] != 0.16 || energyByRefine[1] != 3 {
		t.Fatalf("unexpected R1 values")
	}
	if atkByRefine[5] != 0.40 || energyByRefine[5] != 5 {
		t.Fatalf("unexpected R5 values")
	}
}
