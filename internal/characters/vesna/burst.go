package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// User image 2 gives Q end at 133, damage at 135 and Q -> EE3 at 131.
// See PLACEHOLDER_FRAMES.md; unmeasured transitions use the animation end.
func (c *char) Burst(map[string]int) (action.Info, error) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spiritblade: Burst", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: burst[c.TalentLvlBurst()] * c.spiritbladeBonus()}
	if c.StatusIsActive(radianceKey) {
		ai.AttackTag = attacks.AttackTagReactionStarDiffusionAnemo
		ai.ICDTag = attacks.ICDTagNone
		ai.Durability = 0
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), burstHitmark, burstHitmark)
	c.addMagic(1)
	c.addComposure()
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(burstLength)
	canQueue := burstLength
	// Skill prioritizes the C6 Step when it is ready. The measured 131f
	// transition only applies to Dance; Q -> Step is unmeasured and uses 133f.
	stepReady := c.Base.Cons >= 6 && c.StatusIsActive(stepReadyKey)
	if c.StatusIsActive(spiritbladeArmedKey) && c.specialStage == 2 && !stepReady {
		f[action.ActionSkill] = burstToDance
		canQueue = burstToDance
	}
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: burstLength, CanQueueAfter: canQueue, State: action.BurstState}, nil
}
