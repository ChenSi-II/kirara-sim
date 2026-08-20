package vodyanitsa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	mult := burst[c.TalentLvlBurst()]
	if c.StatusIsActive(microphoneKey) {
		mult *= 1.48
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Water Nymph Aria", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Hydro, Durability: 25, UseHP: true, Mult: mult}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 30, 30)
	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(72)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 72, CanQueueAfter: 60, State: action.BurstState}, nil
}
