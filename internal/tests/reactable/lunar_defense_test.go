package reactable_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func TestDirectLunarDamageIgnoresEnemyDefense(t *testing.T) {
	for _, tc := range []struct {
		tag  attacks.AttackTag
		mult float64
	}{
		{attacks.AttackTagDirectLunarCharged, 3},
		{attacks.AttackTagDirectLunarBloom, 1},
		{attacks.AttackTagDirectLunarCrystallize, 1.6},
	} {
		for _, level := range []int{1, 90, 200} {
			for _, defReduction := range []float64{0, -0.5} {
				for _, ignoreDef := range []float64{0, 0.5, 1} {
					t.Run(fmt.Sprintf("%v/level=%d/def=%g/ignore=%g", tc.tag, level, defReduction, ignoreDef), func(t *testing.T) {
						c, targets := makeCore(1)
						if err := c.Init(); err != nil {
							t.Fatal(err)
						}
						e := targets[0]
						e.Level = level
						e.AddDefMod(info.DefMod{
							Base: modifier.NewBaseWithHitlag("test-defense", 600), Value: defReduction,
						})
						e.AddResistMod(info.ResistMod{
							Base: modifier.NewBaseWithHitlag("test-resistance", 600), Ele: attributes.Electro, Value: 0.1,
						})
						atk := info.AttackEvent{
							Info: info.AttackInfo{
								ActorIndex: 0, AttackTag: tc.tag, Element: attributes.Electro,
								Mult: 2, FlatDmg: 50, IgnoreDefPercent: ignoreDef, SourceIsSim: true,
							},
							Snapshot: info.Snapshot{CharLvl: 90},
						}
						atk.Snapshot.Stats[attributes.BaseATK] = 1000
						want := (2000*tc.mult + 50) * 0.9
						if got := e.HandleAttack(&atk); math.Abs(got-want) > 1e-9 {
							t.Fatalf("damage = %v, want %v", got, want)
						}
					})
				}
			}
		}
	}
}
