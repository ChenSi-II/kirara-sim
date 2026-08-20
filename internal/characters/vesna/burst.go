package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Spiritblade: Burst", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: burst[c.TalentLvlBurst()] * c.spiritbladeBonus()}
	if c.StatusIsActive(radianceKey) {
		ai.AttackTag = attacks.AttackTagReactionStarDiffusionAnemo
		ai.ICDTag = attacks.ICDTagNone
		ai.Durability = 0
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 30, 30)
	c.addMagic(1)
	c.addComposure()
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(72)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 72, CanQueueAfter: 60, State: action.BurstState}, nil
}
