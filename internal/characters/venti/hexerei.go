package venti

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) hexActive() bool { return c.IsHexerei && c.Core.Player.GetHexereiCount() >= 2 }
func (c *char) eyeActive() bool { return c.Core.F >= c.burstSrc+burstStart && c.Core.F < c.burstEnd }

func (c *char) hexInit() {
	if !c.IsHexerei {
		return
	}
	hook := func(args ...any) {
		if !c.hexActive() || !c.eyeActive() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != c.Core.Player.Active() {
			return
		}
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.DmgP] = .5
		c.Core.Player.ActiveChar().AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag("venti-hexerei-dmg", 4*60), AffectedStat: attributes.DmgP,
			Amount: func() []float64 { return buff },
		})
		c.AddStatus("venti-hexerei-eye", 4*60, false)
	}
	for _, ev := range []event.Event{event.OnSwirlPyro, event.OnSwirlHydro, event.OnSwirlCryo, event.OnSwirlElectro, event.OnStarDiffusion} {
		c.Core.Events.Subscribe(ev, hook, "venti-hexerei")
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != c.Index() {
			return
		}
		if a.Info.AttackTag == attacks.AttackTagElementalBurst && c.StatusIsActive("venti-hexerei-eye") {
			a.Info.Mult *= 1.35
		}
		if c.Base.Cons < 6 {
			return
		}
		e := args[0].(*enemy.Enemy)
		if e.StatusIsActive("venti-c6-anemo") {
			a.Snapshot.Stats[attributes.CD] += 1
		}
	}, "venti-hexerei-damage")
}

func (c *char) hexC4() {
	if !c.IsHexerei || c.Base.Cons < 4 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		if ch.Index() != c.Index() && ch.Index() != c.Core.Player.Active() {
			continue
		}
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.AnemoP] = .25
		ch.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("venti-c4", 10*60), AffectedStat: attributes.AnemoP, Amount: func() []float64 { return buff }})
	}
}

func (c *char) hurricaneHit(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.eyeActive() && c.burstExtensions < 2 && !c.StatusIsActive("venti-eye-extension-icd") {
		c.burstExtensions++
		c.burstEnd += 60
		c.ReduceActionCooldown(action.ActionBurst, -30)
		c.AddStatus("venti-eye-extension-icd", 6, false)
	}
	if c.Base.Cons >= 2 && c.Core.Rand.Float64() < .25 {
		c.AddStatus("venti-winds-advent", 15*60, true)
	}
	if c.Base.Cons < 1 || c.StatusIsActive("venti-hurricane-c1-icd") {
		return
	}
	c.AddStatus("venti-hurricane-c1-icd", 15, false)
	ai := a.AttackEvent.Info
	ai.Abil = "Hurricane Arrow (C1)"
	ai.Mult *= .2
	for range 2 {
		c.Core.QueueAttack(ai, combat.NewSingleTargetHit(a.Target.Key()), 0, 1)
	}
}
