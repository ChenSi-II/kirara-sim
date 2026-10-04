package diona

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
	hook := func(args ...any) {
		if !c.StatusIsActive("diona-extra-paws") || c.StatusIsActive("diona-extra-paws-icd") {
			return
		}
		t := args[0].(info.Target)
		if t.Type() != info.TargettableEnemy {
			return
		}
		c.AddStatus("diona-extra-paws-icd", 210, false)
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Icy Paw (Stellar)", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Cryo, Durability: 25, Mult: paw[c.TalentLvlSkill()]}
		for i := range 3 {
			c.Core.QueueAttack(ai, combat.NewSingleTargetHit(t.Key()), 0, 10+i)
		}
	}
	for _, ev := range []event.Event{event.OnSuperconduct, event.OnStarSuperconduct, event.OnSwirlCryo, event.OnStarDiffusion} {
		c.Core.Events.Subscribe(ev, hook, "diona-extra-paws")
	}
	if c.Base.Cons < 6 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.HPP] = .25
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("diona-c6-hp", -1), AffectedStat: attributes.HPP, Amount: func() []float64 { return buff }})
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("diona-c6-stellar", -1), Amount: func(ai info.AttackInfo) float64 {
			if ch.Index() == c.Core.Player.Active() && c.Core.Status.Duration("diona-q") > 0 && c.Core.Combat.Player().IsWithinArea(c.burstBuffArea) && c.RadiantReaction(ai.AttackTag) {
				return .4
			}
			return 0
		}})
	}
}
