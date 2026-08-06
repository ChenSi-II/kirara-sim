package iansan

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {}

func (c *char) a1Buff() {
	if c.Base.Ascension < 1 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.ATKP] = .20
	c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("iansan-precise-movement", 15*60), AffectedStat: attributes.ATKP, Amount: func() []float64 { return buff }})
}
