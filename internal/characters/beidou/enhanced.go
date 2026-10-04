package beidou

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) enhancedHoldCB(hold bool, stacks int) info.AttackCBFunc {
	if !c.Enhanced || !hold || stacks == 0 {
		return nil
	}
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy || c.StatusIsActive("beidou-enhanced-hold-icd") {
			return
		}
		c.AddStatus("beidou-enhanced-hold-icd", 15*60, false)
		c.ReduceActionCooldown(action.ActionSkill, 4*60*stacks)
		c.AddEnergy("beidou-enhanced-hold", float64(8*stacks))
	}
}

func (c *char) enhancedInit() {
	if !c.Enhanced || c.Base.Cons < 6 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.EM] = 200
		ch.AddStatMod(character.StatMod{Base: modifier.NewBase("beidou-c6-em", -1), AffectedStat: attributes.EM, Amount: func() []float64 {
			if c.SuperconductRadiance() && c.StatusIsActive(burstKey) && ch.Index() == c.Core.Player.Active() {
				return buff
			}
			return nil
		}})
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		e := args[0].(*enemy.Enemy)
		if c.SuperconductRadiance() && c.StatusIsActive(burstKey) && e.IsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5)) {
			e.AddResistMod(info.ResistMod{Base: modifier.NewBase("beidou-c6-cryo", 1), Ele: attributes.Cryo, Value: -.15})
		} else {
			e.DeleteResistMod("beidou-c6-cryo")
		}
	}, "beidou-c6-cryo")
}
