package ateaspoonoftranscendence

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func atkBonus(refine int) float64 { return 0.21 + 0.07*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus(p.Refine)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("a-teaspoon-of-transcendence-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount:       func() []float64 { return m },
	})

	const stackKey = "a-teaspoon-of-transcendence-stacks"
	const icdKey = "a-teaspoon-of-transcendence-icd"
	stacks := 0
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("a-teaspoon-of-transcendence-star", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if char.StatusDuration(stackKey) == 0 || !attacks.AttackTagIsStar(ai.AttackTag) {
				return 0
			}
			return float64(stacks) * (0.12 + 0.04*float64(p.Refine))
		},
	})
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || atk.Info.AttackTag != attacks.AttackTagExtra || char.StatusDuration(icdKey) > 0 {
			return
		}
		if char.StatusDuration(stackKey) == 0 {
			stacks = 0
		}
		stacks = min(stacks+1, 3)
		char.AddStatus(stackKey, 5*60, true)
		char.AddStatus(icdKey, 12, false)
	}, fmt.Sprintf("a-teaspoon-of-transcendence-charge-%v", char.Index()))
	return w, nil
}
