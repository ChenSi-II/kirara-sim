package heartsecho

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
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
	core.RegisterWeaponFunc(keys.HeartsEcho, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)

	const (
		emBuffKey   = "hearts-echo-em"
		starBuffKey = "hearts-echo-star"
	)
	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = 45 + 15*r
	starBonus := 0.12 + 0.04*r

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(emBuffKey, 12*60),
			Amount: func() []float64 {
				return em
			},
		})
	}
	for e := event.ReactionEventStartDelim + 1; e < event.ReactionEventEndDelim; e++ {
		c.Events.Subscribe(e, onReaction, fmt.Sprintf("hearts-echo-reaction-%v-%v", char.Base.Key.String(), e))
	}

	onStarReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == char.Index() {
			char.AddStatus(starBuffKey, 12*60, true)
		}
	}
	c.Events.Subscribe(event.OnStarSuperconduct, onStarReaction, fmt.Sprintf("hearts-echo-superconduct-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, onStarReaction, fmt.Sprintf("hearts-echo-diffusion-%v", char.Base.Key.String()))

	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("hearts-echo-star-dmg", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !char.StatusIsActive(starBuffKey) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct,
				attacks.AttackTagReactionStarDiffusionAnemo,
				attacks.AttackTagReactionStarDiffusionCryo:
				return starBonus
			default:
				return 0
			}
		},
	})
	return w, nil
}
