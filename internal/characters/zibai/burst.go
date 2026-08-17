package zibai

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	first := info.AttackInfo{ActorIndex: c.Index(), Abil: "Tri-Sphere Eminence 1", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: burstParam[0][lvl]}
	second := info.AttackInfo{ActorIndex: c.Index(), Abil: "Tri-Sphere Eminence 2", AttackTag: attacks.AttackTagDirectLunarCrystallize, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, UseDef: true, Mult: burstParam[1][lvl]}
	c.Core.QueueAttack(first, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 26, 26)
	c.Core.QueueAttack(second, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 42, 42)
	if c.StatusIsActive(lunarPhaseKey) {
		c.AddStatus(lunarPhaseKey, c.StatusDuration(lunarPhaseKey)+102, true)
	}
	c.SetCD(action.ActionBurst, int(burstParam[2][lvl]*60))
	c.ConsumeEnergy(26)
	f := frames.InitAbilSlice(76)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 76, CanQueueAfter: 64, State: action.BurstState}, nil
}
