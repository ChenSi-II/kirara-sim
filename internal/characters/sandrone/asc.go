package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension < 4 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("sandrone-a4", -1), Extra: true, AffectedStat: attributes.EM, Amount: func() []float64 {
		buff[attributes.EM] = min(c.TotalAtk()/100*8, 160)
		return buff
	}})
}
