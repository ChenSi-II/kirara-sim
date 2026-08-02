package watchsong

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
	core.RegisterWeaponFunc(keys.Watchsong, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)
	energy := 3 + r

	const energyICDKey = "watchsong-energy-icd"
	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(energyICDKey) {
			return
		}
		char.AddStatus(energyICDKey, 9*60, true)
		char.AddEnergy("watchsong", energy)
	}
	for e := event.ReactionEventStartDelim + 1; e < event.ReactionEventEndDelim; e++ {
		c.Events.Subscribe(e, onReaction, fmt.Sprintf("watchsong-reaction-%v-%v", char.Base.Key.String(), e))
	}

	atkBuff := make([]float64, attributes.EndStatType)
	atkBuff[attributes.ATKP] = 0.15 + 0.05*r
	onStarReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag("watchsong-star-atk", 12*60),
			Amount: func() []float64 {
				return atkBuff
			},
		})
	}
	c.Events.Subscribe(event.OnStarSuperconduct, onStarReaction, fmt.Sprintf("watchsong-star-superconduct-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, onStarReaction, fmt.Sprintf("watchsong-star-diffusion-%v", char.Base.Key.String()))
	return w, nil
}
