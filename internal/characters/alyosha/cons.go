package alyosha

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/event"
)

func (c *char) initConstellations() {
	if c.Base.Cons < 1 {
		return
	}
	last := -18 * 60
	hook := func(...any) {
		if c.Core.F-last < 18*60 {
			return
		}
		last = c.Core.F
		c.AddEnergy("alyosha-c1", 15)
	}
	for _, evt := range []event.Event{event.OnElectroCharged, event.OnOverload, event.OnSuperconduct, event.OnQuicken, event.OnAggravate, event.OnHyperbloom} {
		c.Core.Events.Subscribe(evt, hook, fmt.Sprintf("alyosha-c1-%d", evt))
	}
}
