package sourceofkindling

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
	core.RegisterWeaponFunc(keys.SourceOfKindling, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	atkBuffKey    = "source-of-kindling-atk"
	starBuffKey   = "source-of-kindling-star"
	reactBonusKey = "source-of-kindling-reaction-bonus"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	atkBonus := 0.12 + 0.04*r
	starBonus := 0.12 + 0.04*r

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus
	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(atkBuffKey, 12*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return m },
		})
	}
	for evt := event.ReactionEventStartDelim + 1; evt < event.ReactionEventEndDelim; evt++ {
		c.Events.Subscribe(evt, onReaction, fmt.Sprintf("source-of-kindling-reaction-%v-%v", evt, char.Base.Key.String()))
	}

	onStarReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == char.Index() {
			char.AddStatus(starBuffKey, 12*60, true)
		}
	}
	c.Events.Subscribe(event.OnStarSuperconduct, onStarReaction, fmt.Sprintf("source-of-kindling-star-sc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, onStarReaction, fmt.Sprintf("source-of-kindling-star-diff-%v", char.Base.Key.String()))

	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(reactBonusKey, -1),
		Amount: func(ai info.AttackInfo) float64 {
			if char.StatusIsActive(starBuffKey) && attacks.AttackTagIsStar(ai.AttackTag) {
				return starBonus
			}
			return 0
		},
	})
	return w, nil
}
