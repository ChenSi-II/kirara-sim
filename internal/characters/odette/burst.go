package odette

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const dreamKey = "odette-snow-swans-dream"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	// Image 5 Q hits are absolute offsets from Q start, including damage
	// after Q ends at 1.816s; see PLACEHOLDER_FRAMES.md.
	for i, hitmark := range []int{119, 137, 142} {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Bluebird Slash %d", i+1), AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: burstParam[0][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), hitmark, hitmark)
	}
	final := info.AttackInfo{ActorIndex: c.Index(), Abil: "Bluebird Finale", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: burstParam[1][lvl]}
	c.Core.QueueAttack(final, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 148, 148)
	c.AddStatus(dreamKey, int(burstParam[3][lvl]*60), true)
	c.summonDouble(int(burstParam[4][lvl]*60), true)
	c.AddStatus(codaKey, 6*60, true)
	c.SetCD(action.ActionBurst, int(burstParam[5][lvl]*60))
	c.ConsumeEnergy(54)
	f := frames.InitAbilSlice(109)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 109, CanQueueAfter: 109, State: action.BurstState}, nil
}
