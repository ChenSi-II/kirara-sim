package yaemiko

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) kitsuneDuration() int {
	duration := 900 - skillStart
	if c.Enhanced {
		duration += 600
	}
	return duration
}

func (c *char) enhancedInit() {
	if !c.Enhanced {
		return
	}
	hook := func(args ...any) {
		if !c.StatusIsActive("yaemiko-empower-icd") {
			c.AddStatus("yaemiko-empower-icd", 150, false)
			c.AddStatus("yaemiko-empowered-sakura", -1, false)
		}
		if args[1].(*info.AttackEvent).Info.ActorIndex == c.Index() {
			c.enhancedC1()
		}
	}
	c.Core.Events.Subscribe(event.OnSuperconduct, hook, "yaemiko-enhanced")
	c.Core.Events.Subscribe(event.OnStarSuperconduct, hook, "yaemiko-enhanced")
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex == c.Index() && a.Info.AttackTag == attacks.AttackTagReactionStarSuperconduct {
			c.enhancedC1()
		}
	}, "yaemiko-c1-stellar")
	if c.Base.Cons >= 2 {
		for _, ch := range c.Core.Player.Chars() {
			buff := make([]float64, attributes.EndStatType)
			ch.AddStatMod(character.StatMod{Base: modifier.NewBase("yaemiko-c2-em", -1), AffectedStat: attributes.EM, Amount: func() []float64 {
				if len(c.kitsunes) == 0 || (ch.Index() != c.Index() && ch.Index() != c.Core.Player.Active()) {
					return nil
				}
				buff[attributes.EM] = []float64{60, 90, 120, 200}[min(3, len(c.kitsunes))]
				return buff
			}})
		}
	}
	if c.Base.Cons >= 4 {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.DmgP] = 1
		c.AddAttackMod(character.AttackMod{Base: modifier.NewBase("yaemiko-c4-burst", -1), Amount: func(a *info.AttackEvent, _ info.Target) []float64 {
			if a.Info.AttackTag == attacks.AttackTagElementalBurst {
				return buff
			}
			return nil
		}})
	}
	if c.Base.Cons >= 6 {
		apply := func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.ActorIndex == c.Index() && a.Info.AttackTag == attacks.AttackTagReactionStarSuperconduct {
				a.Snapshot.Stats[attributes.CD] += 2
			}
		}
		c.Core.Events.Subscribe(event.OnStarReactionAttack, apply, "yaemiko-c6-stellar")
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			if args[1].(*info.AttackEvent).Info.IsDirectStarDamage() {
				apply(args...)
			}
		}, "yaemiko-c6-stellar-direct")
	}
}

func (c *char) enhancedC1() {
	if c.Base.Cons < 1 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.ElectroP] = .5
		ch.AddStatMod(character.StatMod{Base: modifier.NewBase("yaemiko-c1-electro", 600), AffectedStat: attributes.ElectroP, Amount: func() []float64 { return buff }})
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("yaemiko-c1-stellar", 600), Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag == attacks.AttackTagReactionStarSuperconduct {
				return .5
			}
			return 0
		}})
	}
}

func (c *char) enhancedSkill() {
	if !c.Enhanced || c.Base.Ascension < 1 || len(c.kitsunes) < 3 {
		return
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Sesshou Sakura (A1)", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Electro, Durability: 25, Mult: .4}
	if c.SuperconductRadiance() {
		ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
		ai.Mult = .5
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1), skillStart, skillStart)
}

func (c *char) enhancedTickCB(empowered bool) info.AttackCBFunc {
	done := false
	return func(a info.AttackCB) {
		if !empowered || done || a.Target.Type() != info.TargettableEnemy || !c.SuperconductRadiance() {
			return
		}
		done = true
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Sesshou Sakura (Stellar)", AttackTag: attacks.AttackTagReactionStarSuperconduct, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Electro, Mult: 2}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(a.Target, nil, .5), 0, 1)
	}
}
