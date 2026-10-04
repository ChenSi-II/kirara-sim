package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Armed A1+CA timings are reconstructed from image 2; ordinary CA remains a
// placeholder. The inferred action boundary is recorded in PLACEHOLDER_FRAMES.md.
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	ele := attributes.Physical
	hitmark, animation, canQueue := 30, 52, 36
	if c.StatusIsActive(spiritbladeArmedKey) {
		ele = attributes.Anemo
		hitmark, animation, canQueue = armedChargeHitmark, armedChargeLength, armedChargeLength
		c.queueFeather(armedChargeFeather)
		c.queueFeather(armedChargeFeather)
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Charged Attack", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: ele, Durability: 25, Mult: charge[c.TalentLvlAttack()]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), hitmark, hitmark)
	f := frames.InitAbilSlice(animation)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: animation, CanQueueAfter: canQueue, State: action.ChargeAttackState}, nil
}
