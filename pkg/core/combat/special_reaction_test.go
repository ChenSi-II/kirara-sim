package combat

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func TestCalcSpecialReactionDmgStarMultipliers(t *testing.T) {
	base := CalcReactionBaseDmg(90)
	tests := []struct {
		name string
		tag  attacks.AttackTag
		mult float64
		want float64
	}{
		{name: "star superconduct zero stack exception", tag: attacks.AttackTagReactionStarSuperconduct, mult: 1, want: base},
		{name: "star superconduct one stack", tag: attacks.AttackTagReactionStarSuperconduct, mult: 1.45, want: 1.45 * base},
		{name: "star diffusion anemo", tag: attacks.AttackTagReactionStarDiffusionAnemo, mult: 99, want: 0.75 * base},
		{name: "star diffusion cryo low vortex", tag: attacks.AttackTagReactionStarDiffusionCryo, mult: 2, want: 2 * base},
		{name: "star diffusion cryo high vortex", tag: attacks.AttackTagReactionStarDiffusionCryo, mult: 3, want: 3 * base},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalcSpecialReactionDmg(90, 0, info.AttackInfo{AttackTag: tc.tag, Mult: tc.mult}, 0)
			if math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("got %.12f, want %.12f", got, tc.want)
			}
		})
	}
}
