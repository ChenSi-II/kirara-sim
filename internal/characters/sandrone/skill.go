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
	c.skillStart = c.Core.F
	lvl := c.TalentLvlSkill()
	first := info.AttackInfo{ActorIndex: c.Index(), Abil: "Prism Shot 1", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skill[0][lvl]}
	// Images 1/6: E shots at .600/.800s. They remain scheduled after
	// the .616s animation finishes; see PLACEHOLDER_FRAMES.md.
	c.Core.QueueAttack(first, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 36, 36, c.particleCB)
	second := first
	second.Abil = "Prism Shot 2"
	if c.Core.StarReactions.SuperconductActive {
		second.AttackTag, second.ICDTag, second.Mult = attacks.AttackTagReactionStarSuperconduct, attacks.ICDTagNone, skill[1][lvl]
	} else if c.Core.StarReactions.DiffusionActive {
		second.AttackTag, second.ICDTag, second.Mult = attacks.AttackTagReactionStarDiffusionCryo, attacks.ICDTagNone, skill[2][lvl]
	}
	if c.Base.Ascension >= 1 && c.resolutionPower > 50 && (c.Core.StarReactions.SuperconductActive || c.Core.StarReactions.DiffusionActive) {
		second.Mult *= 4
	}
	// E repairs Faggio regardless of Ascension or Stellar status. The exact
	// repair curve is unmeasured: retain the 50-point lump as an explicit
	// approximation, but do not incorrectly gate repair on the A1 damage buff.
	c.reduceResolutionPower(50)
	c.Core.QueueAttack(second, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 48, 48)
	c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	f := frames.InitAbilSlice(37)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 37, CanQueueAfter: 37, State: action.SkillState}, nil
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("sandrone-particle-icd") {
		c.AddStatus("sandrone-particle-icd", 4*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 2, attributes.Cryo, c.ParticleDelay)
	}
}
