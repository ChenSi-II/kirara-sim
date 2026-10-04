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
			if c.c2Star && c.StatusDuration(c2Key) > 0 && (c.Base.Cons >= 6 || atk.Info.ActorIndex == c.Core.Player.Active()) {
				atk.Snapshot.Stats[attributes.CD] += .60
			}
		}, "vodyanitsa-c2-star-cd")
	}
}

func (c *char) c1Buff() {
	if c.Base.Cons < 1 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATK] = 0.008 * c.MaxHP()
		ch.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("vodyanitsa-c1", 5*60),
			AffectedStat: attributes.ATK,
			Amount:       func() []float64 { return m },
		})
	}
}
