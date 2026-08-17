package odette

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		target.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("odette-marvelous-splendor", -1), Amount: func(ai info.AttackInfo) float64 {
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			default:
				return 0
			}
			bonus := .15 * float64(c.splendor[target.Index()])
			if c.Base.Ascension >= 4 {
				bonus += min(max(c.TotalAtk()-1000, 0)/100*.015, .30)
			}
			if c.Base.Cons >= 6 {
				bonus += .25
				if target.Index() == c.Index() {
					bonus += .20
				}
			}
			return bonus
		}})
		if c.Base.Cons >= 2 {
			target.AddStatMod(character.StatMod{Base: modifier.NewBase("odette-c2-atk", -1), AffectedStat: attributes.ATKP, Amount: func() []float64 {
				out := make([]float64, attributes.EndStatType)
				out[attributes.ATKP] = .07 * float64(c.splendor[target.Index()])
				return out
			}})
		}
	}
}

func (c *char) grantSplendor(src int) {
	if c.Base.Ascension < 1 {
		return
	}
	for i := range c.splendor {
		c.splendor[i] = 0
	}
	c.splendor[c.Index()] = 4
	if c.Base.Cons >= 1 {
		c.splendor[c.Index()] += 2
	}
	for delay := 60; delay <= 20*60; delay += 60 {
		c.QueueCharTask(func() {
			if src != c.doubleSrc || c.Core.Player.Active() == c.Index() || c.splendor[c.Index()] == 0 {
				return
			}
			move := 1
			if c.Base.Cons >= 1 {
				move = 2
			}
			move = min(move, c.splendor[c.Index()])
			c.splendor[c.Index()] -= move
			c.splendor[c.Core.Player.Active()] += move
		}, delay)
	}
}
