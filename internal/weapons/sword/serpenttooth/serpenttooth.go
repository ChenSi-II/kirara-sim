package serpenttooth

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SerpentTooth, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

const (
	trailKey = "serpent-tooth-trail"
	icdKey   = "serpent-tooth-icd"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)
	minimumATK := make([]float64, attributes.EndStatType)
	minimumATK[attributes.ATKP] = 0.15 + 0.05*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("serpent-tooth-minimum-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			if !char.StatusIsActive(trailKey) {
				return nil
			}
			return minimumATK
		},
	})

	// TODO: The simulator does not track distance travelled by the player, so
	// the movement-scaled increase above the guaranteed minimum cannot be
	// implemented without inventing a value. Keep the minimum, activation,
	// duration, cooldown, and swap removal accurate.
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(trailKey, 14*60, true)
		char.AddStatus(icdKey, 14*60, true)
	}, fmt.Sprintf("serpent-tooth-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev == char.Index() {
			char.DeleteStatus(trailKey)
		}
	}, fmt.Sprintf("serpent-tooth-swap-%v", char.Base.Key.String()))

	return w, nil
}
