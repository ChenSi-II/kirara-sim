package prune

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const conversionKey = "prune-witch-tribution"

func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ele, index := attributes.Anemo, 0
	if c.StatusIsActive(conversionKey) && c.converted != attributes.NoElement {
		ele, index = c.converted, 1
		c.DeleteStatus(conversionKey)
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Hexhunter Chime", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: ele, Durability: 25, Mult: skill[index][lvl]}
	// Image 7: initial hammer hits at 0.516s, converted hammer at 0.683s.
	// Their complete/cancel animations remain provisional (58f).
	hitmark := 31
	if index == 1 {
		hitmark = 41
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), hitmark, hitmark, c.skillHit(ele))
	if index == 0 {
		c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	}
	f := frames.InitAbilSlice(58)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 58, CanQueueAfter: 40, State: action.SkillState}, nil
}

func (c *char) skillHit(ele attributes.Element) info.AttackCBFunc {
	if ele != attributes.Anemo {
		return c.convertedHammerHit(ele, attacks.AttackTagElementalArt, "Witch-tribution Ricochet")
	}
	generated := false
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy || generated {
			return
		}
		generated = true
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Anemo, c.ParticleDelay)
	}
}
