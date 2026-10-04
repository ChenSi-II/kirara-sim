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
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 44, 44)
	dur := int(burstParam[2][lvl] * 60)
	if c.Base.Cons >= 6 {
		dur += 4 * 60
	}
	c.bellSrc = c.Core.F
	c.c2Stacks = 0
	// Image 7: first cast hit 0.733s, bell follows after a startup.
	// Include startup so the last C0 (12.733s) / C6 (16.566s) tick lands.
	c.AddStatus(bellKey, dur+45, true)
	if c.Base.Cons >= 2 {
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("prune-c2-huntress", dur+45), AffectedStat: attributes.ATKP, Amount: func() []float64 {
			out := make([]float64, attributes.EndStatType)
			out[attributes.ATKP] = .10 + .05*float64(c.c2Stacks)
			return out
		}})
	}
	hitmarks := []int{175, 292, 409, 526, 646, 764}
	if c.Base.Cons >= 6 {
		hitmarks = append(hitmarks, 877, 994)
	}
	for _, delay := range hitmarks {
		c.QueueCharTask(c.bellTick(c.bellSrc), delay)
	}
	c.SetCD(action.ActionBurst, int(burstParam[3][lvl]*60))
	c.ConsumeEnergy(70)
	// Only the swap endpoint is measured (1.250s). Other transitions use
	// the same endpoint provisionally; see PLACEHOLDER_FRAMES.md.
	f := frames.InitAbilSlice(75)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 75, CanQueueAfter: 60, State: action.BurstState}, nil
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
