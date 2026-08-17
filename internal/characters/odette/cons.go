package odette

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) initConstellations() {
	if c.Base.Cons < 4 {
		return
	}
	last := -210
	hook := func(...any) {
		if c.Core.F-last < 210 {
			return
		}
		last = c.Core.F
		mult, tag := .66, attacks.AttackTagReactionStarSuperconduct
		if c.Core.StarReactions.DiffusionActive {
			mult, tag = .99, attacks.AttackTagReactionStarDiffusionCryo
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Bluebird Coordinated Attack (C4)", AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: mult}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
	}
	c.Core.Events.Subscribe(event.OnStarSuperconduct, hook, "odette-c4-conduct")
	c.Core.Events.Subscribe(event.OnStarDiffusion, hook, "odette-c4-diffusion")
}
