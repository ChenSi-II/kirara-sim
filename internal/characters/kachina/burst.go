package kachina

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

const drillFieldKey = "kachina-turbo-drill-field"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Time to Get Serious!", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Geo, Durability: 50, UseDef: true, Mult: burst[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 42, 42)
	dur := int(burstParam[1][lvl] * 60)
	c.AddStatus(drillFieldKey, dur, true)
	if c.Base.Cons >= 2 && !c.StatusIsActive(twirlyKey) {
		c.summonTwirly(false)
	}
	if c.Base.Cons >= 4 {
		for _, ch := range c.Core.Player.Chars() {
			target := ch
			target.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("kachina-c4", dur), AffectedStat: attributes.DEFP, Amount: func() []float64 {
				if target.Index() != c.Core.Player.Active() {
					return nil
				}
				out := make([]float64, attributes.EndStatType)
				count := len(c.Core.Combat.Enemies())
				out[attributes.DEFP] = []float64{0, .08, .12, .16, .20}[min(count, 4)]
				return out
			}})
		}
	}
	c.SetCD(action.ActionBurst, int(burstParam[2][lvl]*60))
	c.ConsumeEnergy(42)
	f := frames.InitAbilSlice(80)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 80, CanQueueAfter: 68, State: action.BurstState}, nil
}
