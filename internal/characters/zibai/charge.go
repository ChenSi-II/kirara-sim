package zibai

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	if c.StatusIsActive(lunarPhaseKey) {
		return c.lunarCharge()
	}
	for hit, mult := range charge {
		ai := info.AttackInfo{
			ActorIndex: c.Index(), Abil: "Charged Attack", AttackTag: attacks.AttackTagExtra,
			ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault, Element: attributes.Physical,
			Durability: 25, Mult: mult[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 30+hit*6, 30+hit*6)
	}
	f := frames.InitAbilSlice(52)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 52, CanQueueAfter: 36, State: action.ChargeAttackState}, nil
}

func (c *char) lunarCharge() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	for i, index := range []int{10, 11} {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Lunar Phase Charged Attack", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: skillParam[index][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 28+i*7, 28+i*7)
	}
	f := frames.InitAbilSlice(54)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 54, CanQueueAfter: 40, State: action.ChargeAttackState}, nil
}
