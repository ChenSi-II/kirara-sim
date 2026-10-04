package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Action/hit timings: user image 2; see PLACEHOLDER_FRAMES.md for the observed
// values and the inferred EE1 -> EE2 transition.
func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.Base.Cons >= 6 && c.StatusIsActive(stepReadyKey) {
		return c.spiritbladeStep()
	}
	if c.StatusIsActive(spiritbladeArmedKey) {
		return c.spiritbladeSkill()
	}

	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spiritblade: Rise", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: skill[lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), initialSkillHitmark, initialSkillHitmark)
	c.composure = 0
	c.composureExpiries = nil
	c.DeleteStatus(composureKey)
	if c.Base.Cons >= 2 && c.Base.Ascension >= 1 {
		for range 6 {
			c.addComposure()
		}
	}
	c.magic = 2
	c.specialStage = 0
	c.danceCount = 0
	c.freeDance = c.Base.Cons >= 1
	c.armedSrc = c.Core.F
	src := c.armedSrc
	c.AddStatus(spiritbladeArmedKey, 15*60, true)
	c.QueueCharTask(func() {
		if src == c.armedSrc {
			c.endSpiritbladeArmament()
		}
	}, 15*60+1)
	c.SetCD(action.ActionSkill, 18*60)
	// The hit is later than the observed animation end; it must survive the
	// next action beginning at frame 27.
	f := frames.InitAbilSlice(initialSkillLength)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: initialSkillLength, CanQueueAfter: initialSkillLength, State: action.SkillState}, nil
}

func (c *char) spiritbladeSkill() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	base := skill[lvl]
	stage := c.specialStage

	switch stage {
	case 0:
		c.queueVesnaSkillAttack("Spiritblade: Thrust", base, 13)
		c.specialStage = 1
	case 1:
		c.queueVesnaSkillAttack("Spiritblade: Fall", 1.5*base, 21)
		c.queueSpiritbladeAttack("Spiritblade: Fall Blade", skillFallBlade[lvl], 52)
		c.specialStage = 2
	default:
		for i := 0; i < 4; i++ {
			delay := 28 + i*9
			c.queueSpiritbladeAttack("Spiritblade: Dance Blade", skillDanceBlade[lvl], delay)
		}
		c.queueSpiritbladeAttack("Spiritblade: Dance Blade Finale", skillDanceFinale[lvl], 75)
		c.danceCount++
	}

	if stage != 2 || !c.freeDance {
		c.magic--
	} else {
		c.freeDance = false
	}
	c.addComposure()

	maxDance := 3
	if c.Base.Cons >= 1 {
		maxDance++
	}
	if c.Base.Cons >= 6 && stage == 2 {
		c.AddStatus(stepReadyKey, 5*60, true)
	}
	if c.danceCount >= maxDance {
		c.endSpiritbladeArmament()
	}

	animation := spiritbladeSkillLengths[stage]
	f := frames.InitAbilSlice(animation)
	canQueue := animation
	if stage == 0 {
		// Inferred from EE123 ending at 142 rather than 20+52+76=148.
		f[action.ActionSkill] = 14
		canQueue = 14
	}
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: animation, CanQueueAfter: canQueue, State: action.SkillState}, nil
}

func (c *char) spiritbladeStep() (action.Info, error) {
	c.DeleteStatus(stepReadyKey)
	base := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spiritblade: Step", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Anemo, Durability: 25, Mult: 1.5}
	c.Core.QueueAttack(base, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 18, 18)
	star := base
	star.Abil = "Starblade: Step"
	star.Mult = 2 * c.spiritbladeBonus()
	if c.StatusIsActive(radianceKey) {
		star.AttackTag = attacks.AttackTagReactionStarDiffusionAnemo
		star.ICDTag = attacks.ICDTagNone
		star.Durability = 0
	}
	c.Core.QueueAttack(star, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 43, 43)
	if c.StatusIsActive(spiritbladeArmedKey) {
		// The image does not resolve this additional feather's travel time.
		c.queueFeather(43)
	}
	f := frames.InitAbilSlice(97)
	f[action.ActionSkill] = 48
	f[action.ActionDash] = 76
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 97, CanQueueAfter: 48, State: action.SkillState}, nil
}

func (c *char) queueVesnaSkillAttack(abil string, mult float64, delay int) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: abil, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Anemo, Durability: 25, Mult: mult}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, delay)
}

func (c *char) queueSpiritbladeAttack(abil string, mult float64, delay int) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: abil, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Anemo, Durability: 25, Mult: mult * c.spiritbladeBonus()}
	if c.StatusIsActive(radianceKey) {
		ai.AttackTag = attacks.AttackTagReactionStarDiffusionAnemo
		ai.ICDTag = attacks.ICDTagNone
		ai.Durability = 0
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, delay)
}

func (c *char) queueFeather(delay int) {
	if !c.StatusIsActive(spiritbladeArmedKey) {
		return
	}
	src := c.armedSrc
	c.QueueCharTask(func() {
		if src != c.armedSrc || !c.StatusIsActive(spiritbladeArmedKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spirit Feather", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Anemo, Durability: 25, Mult: 0.26 * skill[c.TalentLvlSkill()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
		c.addMagic(1)
	}, delay)
}
