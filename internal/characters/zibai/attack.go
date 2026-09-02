package zibai

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) Attack(map[string]int) (action.Info, error) {
	if c.StatusIsActive(lunarPhaseKey) {
		return c.lunarAttack()
	}
	stage := c.NormalCounter
	for hit, mult := range attack[stage] {
		ai := info.AttackInfo{
			ActorIndex: c.Index(), Abil: fmt.Sprintf("Normal %d-%d", stage+1, hit+1),
			AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack,
			ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault,
			Element: attributes.Physical, Durability: 25, Mult: mult[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 18+hit*6, 18+hit*6)
	}
	c.AdvanceNormalIndex()
	f := frames.InitNormalCancelSlice(18+(len(attack[stage])-1)*6, 42)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 42, CanQueueAfter: 18, State: action.NormalAttackState}, nil
}

func (c *char) lunarAttack() (action.Info, error) {
	stage := c.NormalCounter
	lvl := c.TalentLvlSkill()
	indices := [][]int{{5}, {6}, {7, 8}, {9}}
	for hit, index := range indices[stage] {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Lunar Phase Normal %d-%d", stage+1, hit+1), AttackTag: attacks.AttackTagNormal, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: skillParam[index][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 18+hit*6, 18+hit*6, c.phaseNormalHit)
	}
	// The extra fourth-hit Lunar Crystallize strike is only granted at
	// Moon Sign: Full Moon, not merely while Lunar Phase Shift is active.
	if stage == 3 && c.Core.Player.GetMoonsignLevel() >= 2 {
		extra := info.AttackInfo{ActorIndex: c.Index(), Abil: "Lunar Phase Fourth-Hit Additional DMG", AttackTag: attacks.AttackTagDirectLunarCrystallize, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeSlash, Element: attributes.Geo, UseDef: true, Mult: skillParam[2][lvl]}
		if c.scattermoon && c.Base.Cons >= 4 {
			extra.Mult *= 2.5
			c.scattermoon = false
		}
		c.Core.QueueAttack(extra, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 28, 28)
	}
	c.AdvanceNormalIndex()
	f := frames.InitNormalCancelSlice(30, 46)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 46, CanQueueAfter: 30, State: action.NormalAttackState}, nil
}

func (c *char) phaseNormalHit(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("zibai-phase-na-icd") {
		c.AddStatus("zibai-phase-na-icd", 30, true)
		c.addPhase(5)
	}
}
