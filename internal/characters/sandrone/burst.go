package sandrone

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
	c.AddStatus("sandrone-qed", 88, true)
	for i := 0; i < 3; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Prismatic Bombardment %d", i+1), AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: burst[0][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 24+i*8, 24+i*8)
	}
	beam := info.AttackInfo{ActorIndex: c.Index(), Abil: "Convective Inhibition Ray", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: burst[1][lvl]}
	if c.Core.StarReactions.SuperconductActive {
		beam.AttackTag, beam.ICDTag, beam.Mult = attacks.AttackTagReactionStarSuperconduct, attacks.ICDTagNone, burst[2][lvl]
	}
	if c.Core.StarReactions.DiffusionActive {
		beam.AttackTag, beam.ICDTag, beam.Mult = attacks.AttackTagReactionStarDiffusionCryo, attacks.ICDTagNone, burst[3][lvl]
	}
	if (c.Core.StarReactions.SuperconductActive || c.Core.StarReactions.DiffusionActive) && c.tacticStacks > 0 {
		beam.Mult *= 1 + .10*float64(c.tacticStacks)
		c.tacticStacks = 0
	}
	c.Core.QueueAttack(beam, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 56, 56)
	c.SetCD(action.ActionBurst, int(burstParam[3][lvl]*60))
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(88)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 88, CanQueueAfter: 76, State: action.BurstState}, nil
}
