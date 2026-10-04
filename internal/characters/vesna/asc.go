package vesna

import (
	"github.com/genshinsim/gcsim/internal/characters/vodyanitsa"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.Core.Events.Subscribe(event.OnStarDiffusion, func(...any) {
		duration := 8 * 60
		for _, ch := range c.Core.Player.Chars() {
			if ch.Base.Key == keys.Vodyanitsa && ch.Base.Ascension >= 1 && ch.StatusIsActive(vodyanitsa.SongKey) {
				duration += 4 * 60
				break
			}
		}
		c.AddStatus(radianceKey, duration, true)
	}, "vesna-radiance")

	c.AddStarDamageMod("vesna-star-diffusion-base-dmg", func(atk *info.AttackEvent) {
		switch atk.Info.AttackTag {
		case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			atk.Info.BaseDmgBonus += min(c.TotalAtk()/100*.007, .14)
		}
	})

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
