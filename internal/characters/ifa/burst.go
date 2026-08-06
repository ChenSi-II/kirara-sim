package ifa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	for i, mult := range burst {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Compound Sedation Field", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: mult[lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 40+i*10, 40+i*10)
	}
	if c.Base.Cons >= 4 {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.EM] = 100
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("ifa-c4", 15*60), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
	}
	c.SetCD(action.ActionBurst, int(burstParam[2][lvl]*60))
	c.ConsumeEnergy(40)
	f := frames.InitAbilSlice(78)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 78, CanQueueAfter: 66, State: action.BurstState}, nil
}
