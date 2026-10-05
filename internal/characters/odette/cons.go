package odette

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 2 {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			target, ok := args[0].(*enemy.Enemy)
			if !ok {
				return
			}
			// Evaluate at each hit so expired Radiance or a missing double
			// cannot leave a stale aura debuff. Targets are treated as nearby.
			for _, ele := range []attributes.Element{attributes.Cryo, attributes.Electro, attributes.Anemo} {
				key := "odette-c2-res-" + ele.String()
				target.DeleteResistMod(key)
				if !c.StatusIsActive(doubleKey) {
					continue
				}
				conduct, diffusion := c.SuperconductRadiance(), c.DiffusionRadiance()
				if (ele == attributes.Cryo && (conduct || diffusion)) || (ele == attributes.Electro && conduct) || (ele == attributes.Anemo && diffusion) {
					target.AddResistMod(info.ResistMod{Base: modifier.NewBase(key, 1), Ele: ele, Value: -.20})
				}
			}
		}, "odette-c2-res")
	}
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
		if !c.SuperconductRadiance() && c.DiffusionRadiance() {
			mult, tag = .99, attacks.AttackTagReactionStarDiffusionCryo
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Bluebird Coordinated Attack (C4)", AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: mult}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(args[0].(info.Target), nil, 4), 0, 0)
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, hook, "odette-c4-hit")
}
