package jahoda

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnSkill, func(...any) {
		if c.Core.Player.Active() != c.Index() || !c.StatusIsActive(flaskKey) {
			return
		}
		for _, ch := range c.Core.Player.Chars() {
			if ch.Moonsign == 0 {
				continue
			}
			buff := make([]float64, attributes.EndStatType)
			buff[attributes.CR], buff[attributes.CD] = .05, .40
			ch.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("jahoda-c6", 20*60), Amount: func() []float64 { return buff }})
		}
	}, "jahoda-c6")
}
