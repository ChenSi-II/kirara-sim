package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.AddStarDamageMod("sandrone-star-base", func(atk *info.AttackEvent) {
		atk.Info.BaseDmgBonus += min(c.TotalAtk()/100*.007, .14)
	})
	if c.Base.Ascension < 4 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("sandrone-a4", -1), Extra: true, AffectedStat: attributes.EM, Amount: func() []float64 {
		buff[attributes.EM] = min(c.TotalAtk()/100*8, 160)
		return buff
	}})
}
