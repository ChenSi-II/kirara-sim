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
	c.phaseSrc = c.Core.F
	src := c.phaseSrc
	c.AddStatus(lunarPhaseKey, int(skillParam[3][lvl]*60), true)
	for delay := 60; delay <= int(skillParam[3][lvl]*60); delay += 60 {
		c.QueueCharTask(func() {
			if src == c.phaseSrc && c.StatusIsActive(lunarPhaseKey) {
				c.addPhase(10)
			}
		}, delay)
	}
	c.AddStatus("zibai-selenic-descent", 4*60, true)
	c.SetCD(action.ActionSkill, int(skillParam[4][lvl]*60))
	f := frames.InitAbilSlice(42)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 42, CanQueueAfter: 28, State: action.SkillState}, nil
}

func (c *char) spiritSteed() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	consumed := 70.0
	if c.Base.Cons >= 6 {
		consumed = c.phase
	}
	c.phase -= consumed
	c.strides++
	first := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spirit Steed's Stride 1", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: skillParam[0][lvl]}
	second := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spirit Steed's Stride 2", AttackTag: attacks.AttackTagDirectLunarCrystallize, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, UseDef: true, Mult: skillParam[1][lvl]}
	if c.StatusIsActive("zibai-selenic-descent") {
		second.FlatDmg += .6 * c.TotalDef(false)
	}
	if c.Base.Cons >= 2 && c.Core.Player.GetMoonsignLevel() >= 2 {
		second.FlatDmg += 5.5 * c.TotalDef(false)
	}
	if c.c1FirstStride {
		second.BaseDmgBonus += 2.2
		c.c1FirstStride = false
	}
	if c.Base.Cons >= 6 && consumed > 70 {
		c.c6Elevation = .016 * (consumed - 70)
		second.Elevation += c.c6Elevation
		c.AddStatus("zibai-c6-elevation", 3*60, true)
	}
	c.Core.QueueAttack(first, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 18, 18, c.skillParticle)
	c.Core.QueueAttack(second, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 34, 34)
	if c.Base.Cons >= 4 {
		c.scattermoon = true
	}
	if c.strides >= c.maxStrides() {
		c.DeleteStatus(lunarPhaseKey)
	}
	f := frames.InitAbilSlice(58)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 58, CanQueueAfter: 44, State: action.SkillState}, nil
}

func (c *char) addPhase(v int) {
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
