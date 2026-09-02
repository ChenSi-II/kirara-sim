package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	will := c.will
	if c.Base.Cons >= 4 && c.StatusIsActive(masterstrokeKey) {
		if c.Base.Cons >= 1 {
			will = 300
		} else {
			will = 100
		}
	}
	for i := 0; i < 6; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Manifest Judgment %d", i+1), AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Cryo, Durability: 25, Mult: burstParam[0][lvl], BaseDmgBonus: burstParam[1][lvl] * float64(will)}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 20+i*7, 20+i*7)
	}
	if c.StatusIsActive(masterstrokeKey) {
		c.AddStatus(masterstrokeKey, c.StatusDuration(masterstrokeKey)+99, true)
	}
	if c.Base.Cons < 6 {
		c.will = 0
	} else {
		c.joy = 100
	}
	if c.Base.Cons >= 2 {
		c.evilsbane = true
	}
	c.SetCD(action.ActionBurst, int(burstParam[2][lvl]*60))
	c.ConsumeEnergy(60)
	if c.Base.Cons >= 4 && c.StatusIsActive("lohen-c4-refund") {
		c.DeleteStatus("lohen-c4-refund")
		c.AddEnergy("lohen-c4-burst-refund", 15)
	}
	f := frames.InitAbilSlice(78)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 78, CanQueueAfter: 66, State: action.BurstState}, nil
}
