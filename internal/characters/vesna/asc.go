package vesna

import (
	"math"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.Core.Events.Subscribe(event.OnStarDiffusion, func(...any) {
		c.AddStatus(radianceKey, 8*60, true)
	}, "vesna-radiance")

	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase("vesna-star-diffusion-base-dmg", -1),
			Amount: func(ai info.AttackInfo) float64 {
				switch ai.AttackTag {
				case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				default:
					return 0
				}
				return min(math.Floor(c.TotalAtk()/100)*0.007, 0.14)
			},
		})
	}

	if c.Base.Ascension < 4 {
		return
	}
	cryoAnemoCount := 0
	otherCount := 0
	for _, ch := range c.Core.Player.Chars() {
		if ch.Base.Element == attributes.Cryo || ch.Base.Element == attributes.Anemo {
			cryoAnemoCount++
		} else {
			otherCount++
		}
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("vesna-a4", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() []float64 {
			for i := range m {
				m[i] = 0
			}
			if !c.StatusIsActive(radianceKey) {
				return m
			}
			mult := 1.0
			if c.Base.Cons >= 4 {
				mult = 3
			}
			m[attributes.ATKP] = 0.06 * float64(cryoAnemoCount) * mult
			m[attributes.EM] = 25 * float64(otherCount) * mult
			return m
		},
	})
}
