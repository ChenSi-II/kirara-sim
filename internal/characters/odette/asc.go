package odette

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.AddStarDamageMod("odette-star-base-and-elevation", func(atk *info.AttackEvent) {
		atk.Info.BaseDmgBonus += min(c.TotalAtk()/100*.007, .14)
		if c.Base.Cons >= 6 {
			if c.StatusIsActive(doubleKey) && c.splendor[atk.Info.ActorIndex] > 0 {
				atk.Info.Elevation += .25
			}
			if atk.Info.ActorIndex == c.Index() {
				atk.Info.Elevation += .20
			}
		}
	})
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		target.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("odette-marvelous-splendor", -1), Amount: func(ai info.AttackInfo) float64 {
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			default:
				return 0
			}
			bonus := 0.0
			if c.StatusIsActive(doubleKey) {
				bonus = .15 * float64(c.splendor[target.Index()])
			}
			// Snow Swan's Dream increases Odette's own Stellar reaction damage;
			// C4 extends half of that bonus to the rest of the party.
			if c.StatusIsActive(dreamKey) {
				dream := burstParam[2][c.TalentLvlBurst()]
				if target.Index() == c.Index() {
					bonus += dream
				} else if c.Base.Cons >= 4 {
					bonus += dream * .5
				}
			}
			if c.Base.Ascension >= 4 && target.Index() == c.Index() {
				bonus += min(max(c.TotalAtk()-1000, 0)/100*.015, .30)
			}
			return bonus
		}})
		if c.Base.Cons >= 2 {
			target.AddStatMod(character.StatMod{Base: modifier.NewBase("odette-c2-atk", -1), AffectedStat: attributes.ATKP, Amount: func() []float64 {
				out := make([]float64, attributes.EndStatType)
				if c.StatusIsActive(doubleKey) {
					out[attributes.ATKP] = .07 * float64(c.splendor[target.Index()])
				}
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
	for delay := 60; delay <= c.StatusDuration(doubleKey); delay += 60 {
		c.QueueCharTask(func() {
			if src != c.doubleSrc || !c.StatusIsActive(doubleKey) || c.Core.Player.Active() == c.Index() || c.splendor[c.Index()] == 0 {
				return
			}
			move := 1
			if c.Base.Cons >= 1 {
				move = 2
			}
			move = min(move, c.splendor[c.Index()])
			if c.Base.Cons < 6 {
				c.splendor[c.Index()] -= move
			}
			for _, ally := range c.Core.Player.Chars() {
				if ally.Index() != c.Index() {
					c.splendor[ally.Index()] = c.splendor[ally.Index()] + move
				}
			}
		}, delay)
	}
}
