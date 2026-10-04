package valeriy

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initEffects() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Element != attributes.Electro || a.Info.ActorIndex != c.Core.Player.Active() {
			return
		}
		if c.Base.Ascension >= 1 && c.Core.F < c.burstUntil && a.Info.ActorIndex != c.Index() {
			cap := 80 * c.gainFactor()
			n := min(5*c.gainFactor(), cap-c.burstGained)
			c.burstGained += n
			c.gainPotential(n)
		}
		if c.orders == 0 || c.Core.F >= c.ordersUntil {
			return
		}
		star := a.Info.AttackTag == attacks.AttackTagReactionStarSuperconduct
		if star != c.Core.StarReactions.SuperconductActive {
			return
		}
		if !star {
			switch a.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge, attacks.AttackTagElementalArt, attacks.AttackTagElementalBurst:
			default:
				return
			}
		}
		p := normalParams[c.TalentLvlAttack()]
		scale := p[6]
		a4 := .003
		if star {
			scale = p[7]
			a4 = .009
		}
		if c.Base.Ascension >= 4 {
			scale += a4 * c.spent
		}
		a.Info.FlatDmg += scale * c.TotalAtk()
		c.orders--
	}, "valeriy-orders-and-potential")
	if c.Base.Cons >= 6 {
		for _, ch := range c.Core.Player.Chars() {
			ch.AddAttackMod(character.AttackMod{Base: modifier.NewBase("valeriy-c6-crit", -1), Amount: func(a *info.AttackEvent, t info.Target) []float64 {
				e, ok := t.(*enemy.Enemy)
				if !ok || !e.StatusIsActive("valeriy-c6-mark") || a.Info.Element != attributes.Electro {
					return nil
				}
				m := make([]float64, attributes.EndStatType)
				m[attributes.CR] = .1
				if a.Info.AttackTag == attacks.AttackTagReactionStarSuperconduct {
					m[attributes.CD] = .4
				}
				return m
			}})
		}
		c.AddAttackMod(character.AttackMod{Base: modifier.NewBase("valeriy-c6-damage", -1), Amount: func(a *info.AttackEvent, _ info.Target) []float64 {
			if a.Info.Abil != "Point-Blank Mortar" && a.Info.Abil != "Thunder Shrapnel" {
				return nil
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = 1
			return m
		}})
	}
}
func (c *char) c6Hit(a info.AttackCB) {
	if e, ok := a.Target.(*enemy.Enemy); ok {
		e.AddStatus("valeriy-c6-mark", 420, true)
	}
}
