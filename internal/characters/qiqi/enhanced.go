package qiqi

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) enhancedInit() {
	if !c.Enhanced {
		return
	}
	c.InitDiffusionRadiance(true)
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("qiqi-stellar-bonus", -1), Amount: func(ai info.AttackInfo) float64 {
			if c.StatusIsActive(skillBuffKey) && c.RadiantReaction(ai.AttackTag) {
				return .5
			}
			return 0
		}})
	}
	if c.Base.Cons >= 2 {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.ATKP] = .5
		c.AddStatMod(character.StatMod{Base: modifier.NewBase("qiqi-c2-atk", -1), AffectedStat: attributes.ATKP, Amount: func() []float64 {
			if c.StellarRadiance() {
				return buff
			}
			return nil
		}})
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if !c.StatusIsActive(skillBuffKey) || c.StatusIsActive("qiqi-coordinated-icd") || a.Info.ActorIndex != c.Core.Player.Active() {
			return
		}
		// Reactions and the Herald itself cannot recursively trigger this attack.
		if (a.Info.AttackTag >= attacks.AttackTagNoneStat && !a.Info.IsDirectStarDamage()) || a.Info.Abil == "Herald of Frost (Coordinated)" {
			return
		}
		c.AddStatus("qiqi-coordinated-icd", 132, false)
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Herald of Frost (Coordinated)", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skillCoordinated[c.TalentLvlSkill()]}
		t := args[0].(info.Target)
		var baseC1 info.AttackCBFunc
		if c.Base.Cons >= 1 {
			baseC1 = c.c1
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(t, nil, 2.5), 0, 1, baseC1, c.enhancedC1)
	}, "qiqi-coordinated")
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex == c.Index() || a.Info.ActorIndex != c.Core.Player.Active() || !c.StatusIsActive("qiqi-c6-stacks") || c.c6Stacks <= 0 {
			return
		}
		// Only direct talent damage consumes a charge, never a reaction contribution.
		if !a.Info.IsDirectStarDamage() {
			return
		}
		c.c6Stacks--
		a.Info.FlatDmg += 6 * c.TotalAtk()
	}, "qiqi-c6-stellar")
}

func (c *char) enhancedC1(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || !c.Enhanced || c.Base.Cons < 1 || !c.StellarRadiance() || c.StatusIsActive("qiqi-c1-stellar-icd") {
		return
	}
	c.AddStatus("qiqi-c1-stellar-icd", 6*60, false)
	c.AddEnergy("qiqi-c1-stellar", 6)
}

func (c *char) healLowest(amount float64) {
	var lowest *character.CharWrapper
	for _, ch := range c.Core.Player.Chars() {
		if ch.CurrentHP() > 0 && (lowest == nil || ch.CurrentHPRatio() < lowest.CurrentHPRatio()) {
			lowest = ch
		}
	}
	if lowest != nil {
		c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: lowest.Index(), Message: "Fortune-Preserving Talisman (C4)", Src: amount, Bonus: c.Stat(attributes.Heal)})
	}
}
