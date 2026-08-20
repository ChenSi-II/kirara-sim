package vesna

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 1 {
		c.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase("vesna-c1-star-diffusion", -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !c.StatusIsActive(spiritbladeArmedKey) {
					return 0
				}
				switch ai.AttackTag {
				case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
					return 0.20
				default:
					return 0
				}
			},
		})
	}
	if c.Base.Cons < 2 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("vesna-c2", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			m[attributes.ATKP] = 0
			if c.composure >= 6 && c.StatusIsActive(composureKey) {
				m[attributes.ATKP] = 0.6
			}
			return m
		},
	})
}
