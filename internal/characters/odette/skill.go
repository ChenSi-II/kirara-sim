package odette

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	doubleKey = "odette-solo-dance-double"
	codaKey   = "odette-coda-ready"
)

func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.StatusIsActive(codaKey) && c.StatusIsActive(doubleKey) {
		return c.coda()
	}
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Phantom Night Dancers", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skillParam[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 28, 28, c.skillParticle)
	c.summonDouble(int(skillParam[10][lvl]*60), false)
	c.AddStatus(codaKey, 6*60, true)
	c.SetCD(action.ActionSkill, int(skillParam[11][lvl]*60))
	// Image 5 E1: hit 0.466s, end 0.766s; see PLACEHOLDER_FRAMES.md.
	f := frames.InitAbilSlice(46)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 46, CanQueueAfter: 28, State: action.SkillState}, nil
}

func (c *char) coda() (action.Info, error) {
	// The source table has a separate 15s Coda cooldown. A new E/Q
	// window cannot bypass it; the conflicting window text remains 6s.
	c.codaCooldownUntil = c.Core.F + int(skillParam[12][c.TalentLvlSkill()]*60)
	c.DeleteStatus(codaKey)
	c.AddStatus("odette-double-enhanced", c.StatusDuration(doubleKey), true)
	lvl := c.TalentLvlSkill()
	// Image 5's immediate EE column has its own double timeline. Replace
	// pending E-only attacks, without resummoning or granting Splendor again.
	// A one-frame scheduler delay shifts the recorded EE trace by one frame.
	elapsed := c.Core.F - c.doubleSrc
	if !c.doubleFromBurst && elapsed >= 46 && elapsed <= 47 {
		origin := c.Core.F - 46
		c.doubleTimeline++
		c.scheduleDoubleHits(origin, []int{161, 397, 631, 865, 1098}, []int{271, 507, 742, 976, 1207})
		// EE double disappearance: 21.516s from the initial E.
		c.AddStatus(doubleKey, origin+1291-c.Core.F, true)
	}
	// Image 5 E2 offsets are relative to E2 start, not the preceding E/Q.
	for _, hitmark := range []int{16, 26, 33} {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Coda at Dawn's Tolling", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skillParam[1][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), hitmark, hitmark)
	}
	final := c.stellarAttack("Coda Finale", skillParam[2][lvl], skillParam[3][lvl], attacks.AttackTagReactionStarSuperconduct)
	c.Core.QueueAttack(final, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 66, 66)
	if c.Base.Cons >= 1 {
		extra := c.stellarAttack("Coda Finale (C1)", 3, 4.5, attacks.AttackTagReactionStarSuperconduct)
		c.Core.QueueAttack(extra, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 66, 66)
	}
	f := frames.InitAbilSlice(82)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 82, CanQueueAfter: 70, State: action.SkillState}, nil
}

func (c *char) summonDouble(dur int, fromBurst bool) {
	c.DeleteStatus("odette-double-enhanced")
	c.doubleSrc = c.Core.F
	c.doubleTimeline++
	c.doubleFromBurst = fromBurst
	src := c.doubleSrc
	// Image 5 standalone E and Q columns. Plume and Wing are alternating
	// attacks with roughly a four-second cycle for each, not 90-frame ticks.
	// The diagram's paired stellar hits occur at the same timestamp.
	plume := []int{164, 395, 631, 864, 1097}
	wing := []int{269, 507, 741, 974, 1206}
	startup := 89 // E double disappears at 21.483s after a 20s lifetime.
	if fromBurst {
		plume = []int{257, 490, 723, 958, 1191}
		wing = []int{363, 598, 831, 1068}
		startup = 72 // Q double disappears at 21.200s.
	}
	// Immediate EE replaces this schedule in coda(). Delayed E2 and QE
	// remain distinct recordings, not interchangeable with immediate EE.
	c.AddStatus(doubleKey, dur+startup, true)
	c.grantSplendor(src)
	c.scheduleDoubleHits(c.Core.F, plume, wing)
}

func (c *char) scheduleDoubleHits(origin int, plume, wing []int) {
	for _, hitmark := range plume {
		if delay := origin + hitmark - c.Core.F; delay >= 0 {
			c.QueueCharTask(c.doubleTick(c.doubleSrc, c.doubleTimeline, true), delay)
		}
	}
	for _, hitmark := range wing {
		if delay := origin + hitmark - c.Core.F; delay >= 0 {
			c.QueueCharTask(c.doubleTick(c.doubleSrc, c.doubleTimeline, false), delay)
		}
	}
}

func (c *char) doubleTick(src, timeline int, plume bool) func() {
	return func() {
		if src != c.doubleSrc || timeline != c.doubleTimeline || !c.StatusIsActive(doubleKey) {
			return
		}
		lvl := c.TalentLvlSkill()
		name, normal, conduct, swirl := "Wing", skillParam[7][lvl], skillParam[8][lvl], skillParam[9][lvl]
		if plume {
			name, normal, conduct, swirl = "Plume", skillParam[4][lvl], skillParam[5][lvl], skillParam[6][lvl]
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Dance Double " + name, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: normal}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		if c.StatusIsActive("odette-double-enhanced") && (c.SuperconductRadiance() || c.DiffusionRadiance()) {
			star := c.stellarAttack("Dance Double "+name+" Stellar", conduct, swirl, attacks.AttackTagElementalArt)
			c.Core.QueueAttack(star, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		}
	}
}

func (c *char) stellarAttack(name string, conduct, swirl float64, fallback attacks.AttackTag) info.AttackInfo {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: name, AttackTag: fallback, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: conduct}
	if c.SuperconductRadiance() {
		ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
	}
	if !c.SuperconductRadiance() && c.DiffusionRadiance() {
		ai.AttackTag, ai.Mult = attacks.AttackTagReactionStarDiffusionCryo, swirl
	}
	return ai
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("odette-particle-icd") {
		c.AddStatus("odette-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
	}
}
