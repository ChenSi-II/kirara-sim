package iansan

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

const scaleKey = "iansan-kinetic-energy-scale"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "The Three Principles of Power", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Electro, Durability: 50, Mult: burst[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 38, 38)
	c.restoreNightsoul(15)
	dur := int(burstParam[5][lvl] * 60)
	if c.Base.Cons >= 6 {
		dur += 3 * 60
	}
	c.AddStatus(scaleKey, dur, true)
	iansanAtk := c.TotalAtk()
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		target.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag(scaleKey, dur), AffectedStat: attributes.ATK, Amount: func() []float64 {
			if target.Index() != c.Core.Player.Active() {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			bonus := min(burstParam[3][lvl], burstParam[2][lvl]*float64(c.nightsoul)*iansanAtk)
			if c.nightsoul >= 42 {
				bonus = min(burstParam[3][lvl], burstParam[1][lvl]*iansanAtk)
			}
			out[attributes.ATK] = bonus
			if c.Base.Cons >= 2 && target.Index() != c.Index() {
				out[attributes.ATKP] = .30
			}
			return out
		}})
	}
	c.a1Buff()
	c.SetCD(action.ActionBurst, int(burstParam[6][lvl]*60))
	c.ConsumeEnergy(38)
	f := frames.InitAbilSlice(76)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 76, CanQueueAfter: 64, State: action.BurstState}, nil
}
