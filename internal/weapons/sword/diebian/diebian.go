package diebian

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

type Weapon struct {
	Index int
	next  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	hook := func(...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		switch w.next {
		case 0:
			m := make([]float64, attributes.EndStatType)
			m[attributes.CD] = 0.40 + 0.16*r
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("diebian-crit-dmg", 10*60),
				AffectedStat: attributes.CD,
				Amount:       func() []float64 { return m },
			})
		case 1:
			bonus := 0.27 + 0.09*r
			char.AddReactBonusMod(character.ReactBonusMod{
				Base: modifier.NewBaseWithHitlag("diebian-star-diffusion", 10*60),
				Amount: func(ai info.AttackInfo) float64 {
					switch ai.AttackTag {
					case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
						return bonus
					default:
						return 0
					}
				},
			})
		case 2:
			if char.StatusDuration("diebian-energy-icd") == 0 {
				char.AddEnergy("diebian", 4.5+0.5*r)
				char.AddStatus("diebian-energy-icd", 4*60, false)
			}
		}
		w.next = (w.next + 1) % 3
	}
	c.Events.Subscribe(event.OnSkill, hook, fmt.Sprintf("diebian-skill-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnBurst, hook, fmt.Sprintf("diebian-burst-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if args[0].(int) != char.Index() {
			return
		}
		char.DeleteStatMod("diebian-crit-dmg")
		char.DeleteReactBonusMod("diebian-star-diffusion")
		w.next = 0
	}, fmt.Sprintf("diebian-swap-%v", char.Base.Key.String()))
	return w, nil
}
