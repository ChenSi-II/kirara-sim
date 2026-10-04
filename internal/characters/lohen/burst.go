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
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), burstHitmarks[i], burstHitmarks[i])
	}
	if c.StatusIsActive(masterstrokeKey) {
		c.AddStatus(masterstrokeKey, c.StatusDuration(masterstrokeKey)+99, true)
	}
	if c.Base.Cons < 6 {
		c.will = 0
	} else {
		c.will = will
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
	// Image 8 gives Q -> swap at 2.233s, before the remaining hits land.
	// The full animation/other cancels were not supplied: conservatively use
	// one frame after the last hit. See PLACEHOLDER_FRAMES.md.
	f := frames.InitAbilSlice(172)
	f[action.ActionSwap] = 134
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 172, CanQueueAfter: 134, State: action.BurstState}, nil
}
