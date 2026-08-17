package ifa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const ifaNightsoulKey = "ifa-nightsoul"

func (c *char) Skill(map[string]int) (action.Info, error) {
	c.AddStatus(ifaNightsoulKey, 20*60, true)
	c.SetCD(action.ActionSkill, int(skillParam[4][c.TalentLvlSkill()]*60))
	f := frames.InitAbilSlice(36)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 36, CanQueueAfter: 24, State: action.SkillState}, nil
}

func (c *char) supportingFire() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Tonicshot", AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: skill[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 18, 18, c.tonicshotHit)
	if c.Base.Cons >= 6 && c.Core.Rand.Float64() < .5 {
		extra := ai
		extra.Abil = "Tonicshot (C6)"
		extra.Mult = 1.2
		c.Core.QueueAttack(extra, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 24, 24)
	}
	f := frames.InitAbilSlice(42)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 42, CanQueueAfter: 28, State: action.NormalAttackState}, nil
}

func (c *char) tonicshotHit(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	lvl := c.TalentLvlSkill()
	heal := skillParam[1][lvl]*c.Stat(attributes.EM) + skillParam[2][lvl]
	c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: -1, Message: "Tonicshot", Src: heal, Bonus: c.Stat(attributes.Heal)})
	if c.Base.Cons >= 1 && !c.StatusIsActive("ifa-c1-icd") {
		c.AddStatus("ifa-c1-icd", 8*60, true)
		c.AddEnergy("ifa-c1", 6)
	}
}
