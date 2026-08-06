package kachina

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) initConstellations() {
	if c.Base.Cons < 1 {
		return
	}
	last := -5 * 60
	restore := func(...any) {
		if c.Core.F-last < 5*60 {
			return
		}
		last = c.Core.F
		c.AddEnergy("kachina-c1", 3)
	}
	c.Core.Events.Subscribe(event.OnShielded, restore, "kachina-c1-crystallize")
	c.Core.Events.Subscribe(event.OnLunarCrystallize, restore, "kachina-c1-lunar-crystallize")
	if c.Base.Cons >= 6 {
		c.Core.Events.Subscribe(event.OnShieldBreak, func(...any) {
			if c.StatusIsActive("kachina-c6-icd") {
				return
			}
			c.AddStatus("kachina-c6-icd", 5*60, true)
			ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Turbo Twirly Shield Response", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Geo, UseDef: true, Mult: 2}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
		}, "kachina-c6")
	}
}
