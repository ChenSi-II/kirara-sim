package yaemiko

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// When Sesshou Sakura lightning hits opponents, the Electro DMG Bonus of all nearby party members is increased by 20% for 5s.
func (c *char) c4() {
	if c.Enhanced && !c.StatusIsActive("yaemiko-c4-energy-icd") {
		c.AddStatus("yaemiko-c4-energy-icd", 300, false)
		c.AddEnergy("yaemiko-c4", 8)
	}
	// TODO: does this trigger for yaemiko too? assuming it does
	for _, char := range c.Core.Player.Chars() {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("yaemiko-c4", 5*60),
			AffectedStat: attributes.ElectroP,
			Amount: func() []float64 {
				return c.c4buff
			},
		})
	}
}
