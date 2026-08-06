package kachina

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(...any) {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.GeoP] = .20
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("kachina-a1", 12*60), AffectedStat: attributes.GeoP, Amount: func() []float64 { return buff }})
	}, "kachina-a1")
}
