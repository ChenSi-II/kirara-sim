package vodyanitsa

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 2 {
		c.Core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
			atk, ok := args[1].(*info.AttackEvent)
			if !ok {
				return
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			default:
				return
			}
			active := c.c2StarBuffs[:0]
			for _, buff := range c.c2StarBuffs {
				if buff.expiry <= c.Core.F {
					continue
				}
				active = append(active, buff)
				if buff.target == -1 || buff.target == atk.Info.ActorIndex {
					atk.Snapshot.Stats[attributes.CD] += .60
				}
			}
			c.c2StarBuffs = active
		}, "vodyanitsa-c2-star-cd")
	}
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnHeal, func(args ...any) {
		source := args[0].(*info.HealInfo)
		amount := args[2].(float64)
		if source.Caller != c.Index() || amount <= 0 {
			return
		}
		for _, ch := range c.Core.Player.Chars() {
			m := make([]float64, attributes.EndStatType)
			m[attributes.ATK] = 0.007 * c.MaxHP()
			ch.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("vodyanitsa-c1", 3*60),
				AffectedStat: attributes.ATK,
				Amount:       func() []float64 { return m },
			})
		}
	}, "vodyanitsa-c1-heal")
}
