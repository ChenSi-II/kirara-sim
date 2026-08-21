package prune

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

const bellKey = "prune-hunter-seeker"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "The Bell Tolls!", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Anemo, Durability: 25, Mult: burst[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 30, 30)
	dur := int(burstParam[2][lvl] * 60)
	if c.Base.Cons >= 6 {
		dur += 4 * 60
	}
	c.bellSrc = c.Core.F
	c.c2Stacks = 0
	c.AddStatus(bellKey, dur, true)
	if c.Base.Cons >= 2 {
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("prune-c2-huntress", dur), AffectedStat: attributes.ATKP, Amount: func() []float64 {
			out := make([]float64, attributes.EndStatType)
			out[attributes.ATKP] = .10 + .05*float64(c.c2Stacks)
			return out
		}})
	}
	for delay := 90; delay <= dur; delay += 90 {
		c.QueueCharTask(c.bellTick(c.bellSrc), delay)
	}
	c.SetCD(action.ActionBurst, int(burstParam[3][lvl]*60))
	c.ConsumeEnergy(70)
	f := frames.InitAbilSlice(72)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 72, CanQueueAfter: 60, State: action.BurstState}, nil
}

func (c *char) bellTick(src int) func() {
	return func() {
		if src != c.bellSrc || !c.StatusIsActive(bellKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Witchlure Bell", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: burst[1][c.TalentLvlBurst()]}
		if c.Base.Cons >= 2 {
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0, func(info.AttackCB) { c.c2Stacks = min(6, c.c2Stacks+1) })
			return
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
	}
}
