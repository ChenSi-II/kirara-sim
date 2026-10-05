package zibai

import (
	"slices"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 1 {
		c.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("zibai-c1-stride", -1), Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag == attacks.AttackTagDirectLunarCrystallize && slices.Contains(ai.AdditionalTags, attacks.AdditionalTagZibaiC1) {
				return 2.2
			}
			return 0
		}})
	}
	if c.Base.Cons < 2 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("zibai-c2", -1), Amount: func(ai info.AttackInfo) float64 {
			if c.StatusIsActive(lunarPhaseKey) && (ai.AttackTag == attacks.AttackTagDirectLunarCrystallize || ai.AttackTag == attacks.AttackTagReactionLunarCrystallize) {
				return .30
			}
			return 0
		}})
	}
}
