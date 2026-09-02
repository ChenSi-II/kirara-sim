package nefer

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
	bonus := burstParam[4][lvl] * float64(c.veils)
	for i := 0; i < 2; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("True Eye's Phantasm %d", i+1), AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, Durability: 25, Mult: burstParam[i*2][lvl], FlatDmg: burstParam[i*2+1][lvl] * c.Stat(attributes.EM), BaseDmgBonus: bonus}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 28+i*16, 28+i*16)
	}
	c.veils = 0
	c.AddStatus("nefer-true-eye", 60, true)
	c.SetCD(action.ActionBurst, int(burstParam[5][lvl]*60))
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(76)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 76, CanQueueAfter: 64, State: action.BurstState}, nil
}
