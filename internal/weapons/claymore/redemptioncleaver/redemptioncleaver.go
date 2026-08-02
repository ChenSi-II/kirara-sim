package redemptioncleaver

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.RedemptionCleaver, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = 48 + 16*r
	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("redemption-cleaver-em", 12*60),
			AffectedStat: attributes.EM,
			Amount:       func() []float64 { return em },
		})
	}
	for evt := event.ReactionEventStartDelim + 1; evt < event.ReactionEventEndDelim; evt++ {
		c.Events.Subscribe(evt, onReaction, fmt.Sprintf("redemption-cleaver-reaction-%v-%v", evt, char.Base.Key.String()))
	}

	atk := make([]float64, attributes.EndStatType)
	atk[attributes.ATKP] = 0.12 + 0.04*r
	onStarReaction := func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("redemption-cleaver-atk", 12*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return atk },
		})
	}
	c.Events.Subscribe(event.OnStarSuperconduct, onStarReaction, fmt.Sprintf("redemption-cleaver-star-sc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, onStarReaction, fmt.Sprintf("redemption-cleaver-star-diff-%v", char.Base.Key.String()))

	return w, nil
}
