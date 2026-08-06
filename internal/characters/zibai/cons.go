package zibai

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons < 2 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("zibai-c2", -1), Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag == attacks.AttackTagDirectLunarCrystallize || ai.AttackTag == attacks.AttackTagReactionLunarCrystallize {
				return .30
			}
			return 0
		}})
	}
}
