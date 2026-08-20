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

func (c *char) Attack(map[string]int) (action.Info, error) {
	stage := c.NormalCounter
	ele := attributes.Physical
	if c.StatusIsActive(spiritbladeArmedKey) {
		ele = attributes.Anemo
	}
	for hit, mult := range attack[stage] {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Normal %d-%d", stage+1, hit+1), AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: ele, Durability: 25, Mult: mult[c.TalentLvlAttack()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 18+hit*6, 18+hit*6)
	}
	c.queueFeather(18 + (len(attack[stage])-1)*6)
	c.AdvanceNormalIndex()
	f := frames.InitNormalCancelSlice(18+(len(attack[stage])-1)*6, 42)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 42, CanQueueAfter: 18, State: action.NormalAttackState}, nil
}
