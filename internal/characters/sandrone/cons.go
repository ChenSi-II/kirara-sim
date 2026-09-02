package sandrone

import (
	"strings"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 1 {
		c.initC1()
	}
	if c.Base.Cons >= 2 {
		c.initC2()
	}
	if c.Base.Cons < 4 {
		return
	}
	c.initC4()
}

func (c *char) initC1() {
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		target.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("sandrone-c1-c6", -1), Amount: func(ai info.AttackInfo) float64 {
			return c.c1ReactBonus(target, ai)
		}})
	}
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		if atk.Info.ActorIndex != c.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			atk.Info.Elevation += .20
		}
	}, "sandrone-c6-elevation")
}

func (c *char) c1ReactBonus(target *character.CharWrapper, ai info.AttackInfo) float64 {
	if c.StatusIsActive("sandrone-resolution") {
		switch ai.AttackTag {
		case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			return .30
		}
		return 0
	}
	if c.Base.Cons < 6 || target.Index() != c.Index() {
		return 0
	}
	return .20
}

func (c *char) initC2() {
	c.AddAttackMod(character.AttackMod{Base: modifier.NewBase("sandrone-c2-ray-cd", -1), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
		if atk.Info.ActorIndex != c.Index() || !strings.Contains(atk.Info.Abil, "Ray") {
			return nil
		}
		out := make([]float64, attributes.EndStatType)
		out[attributes.CD] = .40 + .20*float64(min(max(c.resolutionRays-1, 0), 3))
		return out
	}})
}

func (c *char) initC4() {
	last := -4 * 60
	hook := func(args ...any) {
		if c.Core.F-last < 4*60 {
			return
		}
		last = c.Core.F
		mult, tag := 1.25, attacks.AttackTagReactionStarSuperconduct
		if c.Core.StarReactions.DiffusionActive {
			mult, tag = 1.875, attacks.AttackTagReactionStarDiffusionCryo
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Prismatic Resonance Cannon (C4)", AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: mult}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
	}
	c.Core.Events.Subscribe(event.OnStarSuperconduct, hook, "sandrone-c4-conduct")
	c.Core.Events.Subscribe(event.OnStarDiffusion, hook, "sandrone-c4-diffusion")
}
