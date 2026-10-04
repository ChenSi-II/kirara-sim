package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Outside Masterstroke, hitmarks/cancels remain placeholders; see
// PLACEHOLDER_FRAMES.md for the measured stance timeline and inferred splits.
func (c *char) Attack(map[string]int) (action.Info, error) {
	stage := c.NormalCounter
	hitmarks := make([]int, len(attack[stage]))
	animation := 42
	for hit := range hitmarks {
		hitmarks[hit] = 18 + hit*6
	}
	masterstroke := c.StatusIsActive(masterstrokeKey)
	if masterstroke {
		hitmarks = masterstrokeNormalHitmarks[stage]
		animation = masterstrokeNormalFrames[stage]
	}
	for hit, mult := range attack[stage] {
		ai := info.AttackInfo{
			ActorIndex: c.Index(), Abil: fmt.Sprintf("Normal %d-%d", stage+1, hit+1),
			AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack,
			ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce,
			Element: c.attackElement(), Durability: 25, Mult: mult[c.TalentLvlAttack()],
		}
		if masterstroke {
			param := []int{0, 1, 2, 3, 4}[stage]
			if stage == 4 {
				param += hit
			}
			ai.Mult = skillParam[param][c.skillLevel()]
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), hitmarks[hit], hitmarks[hit], c.masterstrokeHit)
	}
	c.AdvanceNormalIndex()
	f := frames.InitNormalCancelSlice(hitmarks[len(hitmarks)-1], animation)
	if masterstroke {
		// Only N1C is recorded. Other normal -> CA transitions remain
		// conservatively equal to that normal's inferred duration.
		if stage == 0 {
			f[action.ActionCharge] = hitmarks[0]
		}
	}
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: animation, CanQueueAfter: hitmarks[0], State: action.NormalAttackState}, nil
}
