package lohen

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Masterstroke N1C timing is inferred from image 8; ordinary CA retains
// placeholder timing. See PLACEHOLDER_FRAMES.md.
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	hitmarks := []int{30, 36}
	animation, queue := 52, 36
	if c.StatusIsActive(masterstrokeKey) {
		hitmarks = masterstrokeChargeHitmarks
		animation, queue = 36, 33
	}
	for hit, mult := range charge {
		ai := info.AttackInfo{
			ActorIndex: c.Index(), Abil: "Charged Attack", AttackTag: attacks.AttackTagExtra,
			ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypePierce, Element: c.attackElement(),
			Durability: 25, Mult: mult[c.TalentLvlAttack()],
		}
		if c.StatusIsActive(masterstrokeKey) {
			ai.Mult = skillParam[6][c.skillLevel()]
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), hitmarks[hit], hitmarks[hit], c.masterstrokeHit)
	}
	f := frames.InitAbilSlice(animation)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: animation, CanQueueAfter: queue, State: action.ChargeAttackState}, nil
}
