package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.StatusIsActive(spiritbladeArmedKey) {
		return c.spiritbladeSkill()
	}

	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spiritblade: Rise", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: skill[lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 30, 30)
	c.composure = 0
	c.DeleteStatus(composureKey)
	if c.Base.Cons >= 2 {
		c.composure = 6
		c.AddStatus(composureKey, 20*60, true)
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
	f := frames.InitAbilSlice(60)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 60, CanQueueAfter: 44, State: action.SkillState}, nil
}

func (c *char) spiritbladeSkill() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	base := skill[lvl]
	stage := c.specialStage

	switch stage {
	case 0:
		c.queueVesnaSkillAttack("Spiritblade: Thrust", base, 20)
		c.queueSpiritbladeAttack("Spiritblade: Thrust Blade", base, 20)
		c.specialStage = 1
	case 1:
		c.queueVesnaSkillAttack("Spiritblade: Fall", 1.5*base, 24)
		c.queueSpiritbladeAttack("Spiritblade: Fall Blade", 2.5*base, 28)
		c.specialStage = 2
	default:
		for i := 0; i < 4; i++ {
			delay := 14 + i*6
			c.queueVesnaSkillAttack("Spiritblade: Dance", base, delay)
			c.queueSpiritbladeAttack("Spiritblade: Dance Blade", base, delay)
		}
		c.queueVesnaSkillAttack("Spiritblade: Dance Finale", 3.5*base, 42)
		c.queueSpiritbladeAttack("Spiritblade: Dance Blade Finale", 3.5*base, 42)
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
	if c.danceCount >= maxDance {
		c.endSpiritbladeArmament()
	}

	f := frames.InitAbilSlice(54)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 54, CanQueueAfter: 38, State: action.SkillState}, nil
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
