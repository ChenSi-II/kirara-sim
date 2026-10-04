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
	hook := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if !attacks.AttackTagIsStar(atk.Info.AttackTag) {
			return
		}
		if c.Core.F-last < 210 {
			return
		}
		last = c.Core.F
		mult, tag := .66, attacks.AttackTagReactionStarSuperconduct
		if !c.Core.StarReactions.SuperconductActive && c.Core.StarReactions.DiffusionActive {
			mult, tag = .99, attacks.AttackTagReactionStarDiffusionCryo
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Bluebird Coordinated Attack (C4)", AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: mult}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(args[0].(info.Target), nil, 4), 0, 0)
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, hook, "odette-c4-hit")
}
