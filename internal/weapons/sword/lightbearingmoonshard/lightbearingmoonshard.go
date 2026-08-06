package lightbearingmoonshard

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

const lunarCrystallizeBuffKey = "lightbearing-moonshard-lunar-crystallize"

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

// Increases DEF by 20/25/30/35/40%. Lunar-Crystallize reaction DMG increases
// by 64/80/96/112/128% for 5s after the equipping character uses an
// Elemental Skill.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := p.Refine

	m := make([]float64, attributes.EndStatType)
	m[attributes.DEFP] = 0.15 + 0.05*float64(r)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("lightbearing-moonshard-def", -1),
		AffectedStat: attributes.DEFP,
		Amount: func() []float64 {
			return m
		},
	})

	lunarCrystallizeBonus := 0.48 + 0.16*float64(r)
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarCrystallizeBuffKey, 5*60),
			Amount: func(atk info.AttackInfo) float64 {
				switch atk.AttackTag {
				case attacks.AttackTagDirectLunarCrystallize,
					attacks.AttackTagReactionLunarCrystallize:
					return lunarCrystallizeBonus
				default:
					return 0
				}
			},
		})
	}, fmt.Sprintf("lightbearing-moonshard-%v", char.Base.Key.String()))

	return w, nil
}
