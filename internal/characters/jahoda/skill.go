package jahoda

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const flaskKey = "jahoda-purr-loined-treasure-flask"

func (c *char) Skill(p map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	index := 0
	if p["filled"] != 0 {
		index = 2
	} else if p["contact"] != 0 {
		index = 1
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Splitting the Spoils", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: skill[index][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 28, 28, c.skillParticle)
	c.flaskSrc = c.Core.F
	c.AddStatus(flaskKey, int(skillParam[3][lvl]*60), true)
	if p["filled"] != 0 {
		for delay := 2 * 60; delay <= int(skillParam[3][lvl]*60); delay += 210 {
			c.QueueCharTask(c.meowball(c.flaskSrc), delay)
		}
	}
	c.SetCD(action.ActionSkill, int(skillParam[5][lvl]*60))
	f := frames.InitAbilSlice(54)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 54, CanQueueAfter: 38, State: action.SkillState}, nil
}

func (c *char) meowball(src int) func() {
	return func() {
		if src != c.flaskSrc || !c.StatusIsActive(flaskKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Fluffy Meowball", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: skill[3][c.TalentLvlSkill()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
		if c.Base.Cons >= 1 && c.Core.Rand.Float64() < .5 {
			bounce := ai
			bounce.Abil = "Fluffy Meowball Bounce"
			c.Core.QueueAttack(bounce, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 8, 8)
		}
		c.AddEnergy("jahoda-meowball", 2)
	}
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("jahoda-particle-icd") {
		c.AddStatus("jahoda-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Anemo, c.ParticleDelay)
	}
}
