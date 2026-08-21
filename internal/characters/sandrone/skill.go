package sandrone

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	first := info.AttackInfo{ActorIndex: c.Index(), Abil: "Prism Shot 1", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skill[0][lvl]}
	c.Core.QueueAttack(first, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 20, 20, c.particleCB)
	second := first
	second.Abil = "Prism Shot 2"
	if c.Core.StarReactions.SuperconductActive {
		second.AttackTag, second.ICDTag, second.Mult = attacks.AttackTagReactionStarSuperconduct, attacks.ICDTagNone, skill[1][lvl]
	} else if c.Core.StarReactions.DiffusionActive {
		second.AttackTag, second.ICDTag, second.Mult = attacks.AttackTagReactionStarDiffusionCryo, attacks.ICDTagNone, skill[2][lvl]
	}
	if c.Base.Ascension >= 1 && c.resolutionPower > 50 && (c.Core.StarReactions.SuperconductActive || c.Core.StarReactions.DiffusionActive) {
		second.Mult *= 4
		c.reduceResolutionPower(50)
	}
	c.Core.QueueAttack(second, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 32, 32)
	c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	f := frames.InitAbilSlice(54)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 54, CanQueueAfter: 38, State: action.SkillState}, nil
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("sandrone-particle-icd") {
		c.AddStatus("sandrone-particle-icd", 4*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Cryo, c.ParticleDelay)
	}
}
