package zibai

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const lunarPhaseKey = "zibai-lunar-phase-shift"

func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.StatusIsActive(lunarPhaseKey) {
		return c.spiritSteed()
	}
	lvl := c.TalentLvlSkill()
	c.phase, c.strides = 0, 0
	if c.Base.Cons >= 1 {
		c.phase = 100
		c.c1FirstStride = true
	}
	c.phaseSrc++
	src := c.phaseSrc
	c.AddStatus(lunarPhaseKey, int(skillParam[3][lvl]*60), true)
	c.QueueCharTask(c.phaseTick(src), 60)
	c.watchPhaseExpiry(src)
	c.AddStatus("zibai-selenic-descent", 4*60, true)
	c.SetCD(action.ActionSkill, int(skillParam[4][lvl]*60))
	f := frames.InitAbilSlice(phaseSkillFrames)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: phaseSkillFrames, CanQueueAfter: phaseSkillFrames, State: action.SkillState}, nil
}

func (c *char) spiritSteed() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	consumed := 70.0
	if c.Base.Cons >= 6 {
		consumed = c.phase
	}
	c.phase -= consumed
	c.lastPhaseReaction = c.Core.F - 4*60
	c.strides++
	first := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spirit Steed's Stride 1", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: skillParam[0][lvl]}
	second := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spirit Steed's Stride 2", AttackTag: attacks.AttackTagDirectLunarCrystallize, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, UseDef: true, Mult: skillParam[1][lvl]}
	if c.Base.Ascension >= 1 && c.StatusIsActive("zibai-selenic-descent") {
		second.FlatDmg += .6 * c.TotalDef(false)
	}
	if c.Base.Ascension >= 1 && c.StatusIsActive("zibai-selenic-descent") && c.Base.Cons >= 2 && c.Core.Player.GetMoonsignLevel() >= 2 {
		second.FlatDmg += 5.5 * c.TotalDef(false)
	}
	if c.c1FirstStride {
		second.AdditionalTags = append(second.AdditionalTags, attacks.AdditionalTagZibaiC1)
		c.c1FirstStride = false
	}
	if c.Base.Cons >= 6 {
		c.c6Elevation = .016 * (consumed - 70)
		c.AddStatus("zibai-c6-elevation", 3*60, true)
	}
	src := c.phaseSrc
	grantScattermoon := func(a info.AttackCB) {
		if a.Target.Type() == info.TargettableEnemy && c.Base.Cons >= 4 && src == c.phaseSrc && c.StatusIsActive(lunarPhaseKey) {
			c.scattermoon = true
		}
	}
	c.Core.QueueAttack(first, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 18, 18, c.skillParticle, grantScattermoon)
	c.Core.QueueAttack(second, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 34, 34, grantScattermoon)
	if c.strides >= c.maxStrides() {
		c.endPhase()
	}
	f := frames.InitAbilSlice(phaseSkillFrames)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: phaseSkillFrames, CanQueueAfter: phaseSkillFrames, State: action.SkillState}, nil
}

func (c *char) addPhase(v int) {
	if !c.StatusIsActive(lunarPhaseKey) {
		return
	}
	gain := float64(v)
	if c.Base.Cons >= 6 && c.StatusIsActive(lunarPhaseKey) {
		gain *= 1.5
	}
	c.phase = min(100, c.phase+gain)
}

func (c *char) maxStrides() int {
	if c.Base.Cons >= 1 {
		return 5
	}
	return 4
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("zibai-particle-icd") {
		c.AddStatus("zibai-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Geo, c.ParticleDelay)
	}
}

// Reschedule from the live status so Burst extensions retain natural recovery.
func (c *char) phaseTick(src int) func() {
	return func() {
		if src != c.phaseSrc || !c.StatusIsActive(lunarPhaseKey) {
			return
		}
		c.addPhase(10)
		c.QueueCharTask(c.phaseTick(src), 60)
	}
}

func (c *char) watchPhaseExpiry(src int) {
	c.QueueCharTask(func() {
		if src != c.phaseSrc {
			return
		}
		if c.StatusIsActive(lunarPhaseKey) {
			c.watchPhaseExpiry(src)
			return
		}
		c.endPhase()
	}, c.StatusDuration(lunarPhaseKey)+1)
}

func (c *char) endPhase() {
	c.DeleteStatus(lunarPhaseKey)
	c.phaseSrc++
	c.phase = 0
	c.scattermoon = false
	c.Character.ResetNormalCounter()
}
