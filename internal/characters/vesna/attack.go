package vesna

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Armed timings come from image 2; ordinary physical attacks still use the
// generated placeholder table. See PLACEHOLDER_FRAMES.md.
func (c *char) Attack(map[string]int) (action.Info, error) {
	if c.StatusIsActive(stepReadyKey) && c.Base.Cons >= 6 {
		return c.spiritbladeStep()
	}
	stage := c.NormalCounter
	ele := attributes.Physical
	armed := c.StatusIsActive(spiritbladeArmedKey)
	hitmarks := make([]int, len(attack[stage]))
	for hit := range hitmarks {
		hitmarks[hit] = 18 + hit*6
	}
	animation := 42
	if armed {
		ele = attributes.Anemo
		hitmarks = armedNormalHitmarks[stage]
		animation = armedNormalLengths[stage]
	}
	for hit, mult := range attack[stage] {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Normal %d-%d", stage+1, hit+1), AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: ele, Durability: 25, Mult: mult[c.TalentLvlAttack()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), hitmarks[hit], hitmarks[hit])
	}
	lastHit := hitmarks[len(hitmarks)-1]
	if armed {
		c.queueFeather(lastHit + armedNormalFeatherDelay)
	}
	c.AdvanceNormalIndex()
	// Unmeasured non-normal cancel points retain the hitmark approximation.
	f := frames.InitNormalCancelSlice(lastHit, animation)
	if armed && stage == 0 {
		f[action.ActionCharge] = armedNormalChargeCancel
	}
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: animation, CanQueueAfter: hitmarks[0], State: action.NormalAttackState}, nil
}
