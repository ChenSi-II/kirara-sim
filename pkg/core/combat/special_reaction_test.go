package combat

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func TestCalcSpecialReactionDmg(t *testing.T) {
	tests := []struct {
		name        string
		base        float64
		coefficient float64
		want        float64
	}{
		{"direct diffusion and zero-stack superconduct", 2000, 1, 10325},
		{"direct superconduct one stack", 2000, 1.45, 14915},
		{"direct superconduct twelve stacks", 2000, 2, 20525},
		{"reaction diffusion anemo", 1000, .75, 3950},
		{"reaction diffusion cryo low vortex", 1000, 2, 10325},
		{"reaction diffusion cryo high vortex", 1000, 3, 15425},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// EM 1000 gives +200%; reaction bonus is +40%. Base bonus
			// applies before the flat 100, and elevation applies once to both.
			got := CalcSpecialReactionDmg(tc.base, tc.coefficient, .4, info.AttackInfo{
				BaseDmgBonus: .2, FlatDmg: 100, Elevation: .25,
			}, 1000)
			if math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("got %.12f, want %.12f", got, tc.want)
			}
		})
	}
}

func TestCalcLunarReactionDmgCompatibility(t *testing.T) {
	for _, tc := range []struct {
		tag         attacks.AttackTag
		coefficient float64
	}{
		{attacks.AttackTagReactionLunarCharge, 3},
		{attacks.AttackTagReactionLunarCrystallize, 1.6},
	} {
		ai := info.AttackInfo{AttackTag: tc.tag, BaseDmgBonus: .2, FlatDmg: 100, Elevation: .25}
		want := (CalcReactionBaseDmg(90)*tc.coefficient*3.4*1.2 + 100) * 1.25
		if got := CalcLunarReactionDmg(90, .4, ai, 1000); math.Abs(got-want) > 1e-9 {
			t.Fatalf("%v: got %v, want %v", tc.tag, got, want)
		}
	}
}
