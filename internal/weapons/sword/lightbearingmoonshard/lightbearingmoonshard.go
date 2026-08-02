package lightbearingmoonshard

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.LightbearingMoonshard, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const lunarCrystallizeBuffKey = "lightbearing-moonshard-lunar-crystallize"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	def := make([]float64, attributes.EndStatType)
	def[attributes.DEFP] = 0.15 + 0.05*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("lightbearing-moonshard-def", -1),
		AffectedStat: attributes.DEFP,
		Amount:       func() []float64 { return def },
	})

	lunarCrystallizeBonus := 0.48 + 0.16*r
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBaseWithHitlag(lunarCrystallizeBuffKey, 5*60),
			Amount: func(ai info.AttackInfo) float64 {
				switch ai.AttackTag {
				case attacks.AttackTagReactionLunarCrystallize, attacks.AttackTagDirectLunarCrystallize:
					return lunarCrystallizeBonus
				default:
					return 0
				}
			},
		})
	}, fmt.Sprintf("lightbearing-moonshard-skill-%v", char.Base.Key.String()))

	return w, nil
}
