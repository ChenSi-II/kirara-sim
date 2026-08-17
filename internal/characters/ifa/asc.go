package ifa

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension >= 1 {
		for _, ch := range c.Core.Player.Chars() {
			ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("ifa-rescue-essentials", -1), Amount: func(ai info.AttackInfo) float64 {
				if !c.StatusIsActive(ifaNightsoulKey) {
					return 0
				}
				points := 80.
				if c.Base.Cons >= 2 {
					points = min(points+50, 200)
				}
				switch ai.AttackTag {
				case attacks.AttackTagSwirlPyro, attacks.AttackTagSwirlHydro, attacks.AttackTagSwirlCryo, attacks.AttackTagSwirlElectro, attacks.AttackTagECDamage:
					return .015 * points
				case attacks.AttackTagReactionLunarCharge, attacks.AttackTagDirectLunarCharged, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
					return .0024 * points
				}
				return 0
			}})
		}
	}
	if c.Base.Ascension < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(...any) {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.EM] = 80
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("ifa-a4", 10*60), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
	}, "ifa-a4")
}
