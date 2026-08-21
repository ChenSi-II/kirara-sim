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
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 28, 28, c.skillHit(ele))
	if index == 0 {
		c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	}
	f := frames.InitAbilSlice(58)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 58, CanQueueAfter: 40, State: action.SkillState}, nil
}

func (c *char) skillHit(ele attributes.Element) info.AttackCBFunc {
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		if ele == attributes.Anemo {
			target, ok := a.Target.(info.Reactable)
			if !ok {
				return
			}
			for _, candidate := range []attributes.Element{attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo} {
				if target.AuraContains(candidate) {
					c.converted = candidate
					c.AddStatus(conversionKey, 6*60, true)
					break
				}
			}
			c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Anemo, c.ParticleDelay)
			return
		}
		c.tollingRally()
		if c.Base.Cons >= 2 {
			c.c2Stacks = min(6, c.c2Stacks+1)
		}
		if c.Base.Cons >= 1 && !c.StatusIsActive("prune-c1-icd") {
			c.AddStatus("prune-c1-icd", 108, true)
			c.AddEnergy("prune-c1", 2)
		}
		if c.Base.Cons >= 4 {
			ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Witch-tribution Ricochet", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: ele, Mult: .8}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 8, 8)
		}
	}
}
