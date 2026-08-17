package illuga

import "github.com/genshinsim/gcsim/pkg/core/event"

func (c *char) initConstellations() {
	if c.Base.Cons < 1 {
		return
	}
	last := -15 * 60
	restore := func(...any) {
		if c.Core.F-last < 15*60 {
			return
		}
		last = c.Core.F
		c.AddEnergy("illuga-c1", 12)
	}
	c.Core.Events.Subscribe(event.OnShielded, restore, "illuga-c1-crystallize")
	c.Core.Events.Subscribe(event.OnLunarCrystallize, restore, "illuga-c1-lunar-crystallize")
}
