// Chernaya (秘星典谕), Nanoka 7.1.51, weapon ID 14525.
package chernaya

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

func init() { core.RegisterWeaponFunc(keys.Chernaya, NewWeapon) }

type Weapon struct {
	Index         int
	c             *core.Core
	char          *character.CharWrapper
	refine        int
	stacks, until int
}

func (w *Weapon) SetIndex(i int) { w.Index = i }
func NewWeapon(c *core.Core, ch *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{c: c, char: ch, refine: p.Refine}, nil
}
func (w *Weapon) Init() error {
	r := float64(w.refine)
	w.char.AddStatMod(character.StatMod{Base: modifier.NewBase("chernaya-stats", -1), AffectedStat: attributes.NoStat, Amount: func() []float64 {
		m := make([]float64, attributes.EndStatType)
		m[attributes.CD] = .18 + .06*r
		if w.c.F < w.until {
			m[attributes.EM] = 36 + 12*r + float64(w.stacks)*(18+6*r)
		}
		return m
	}})
	w.c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if w.c.Player.Active() != w.char.Index() {
			return
		}
		w.stacks = 0
		w.until = w.c.F + 1200
	}, fmt.Sprintf("chernaya-skill-%d", w.char.Index()))
	for e := event.ReactionEventStartDelim + 1; e < event.ReactionEventEndDelim; e++ {
		w.c.Events.Subscribe(e, func(args ...any) {
			if len(args) < 2 || w.c.F >= w.until {
				return
			}
			a, ok := args[1].(*info.AttackEvent)
			if !ok || a.Info.ActorIndex < 0 || a.Info.ActorIndex >= len(w.c.Player.Chars()) {
				return
			}
			w.stacks = min(3, w.stacks+1)
		}, fmt.Sprintf("chernaya-react-%d-%d", w.char.Index(), e))
	}
	w.char.AddAttackMod(character.AttackMod{Base: modifier.NewBase("chernaya-star-cd", -1), Amount: func(a *info.AttackEvent, _ info.Target) []float64 {
		if w.c.F >= w.until || (!w.c.StarReactions.SuperconductActive && !w.c.StarReactions.DiffusionActive) || a.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
			return nil
		}
		m := make([]float64, attributes.EndStatType)
		m[attributes.CD] = float64(w.stacks) * (.09 + .03*r)
		return m
	}})
	return nil
}
