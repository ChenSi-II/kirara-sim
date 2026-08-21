package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("sandrone-star-base", -1), Amount: func(ai info.AttackInfo) float64 {
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				return min(c.TotalAtk()/100*.007, .14)
			}
			return 0
		}})
	}
	if c.Base.Ascension < 4 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("sandrone-a4", -1), Extra: true, AffectedStat: attributes.EM, Amount: func() []float64 {
		buff[attributes.EM] = min(c.TotalAtk()/100*8, 160)
		return buff
	}})
}
