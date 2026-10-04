package mitya

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
	c.Core.StarReactions.SuperconductEnabled = true
	c.Core.StarReactions.BeaconSuperconduct = true
	c.Core.Events.Subscribe(event.OnStarSuperconduct, func(...any) {
		// Refresh a single generation window; repeated reactions do not create
		// overlapping periodic generators. Cadence is stated in the passive.
		dur := 180
		if c.Base.Ascension >= 1 && c.reactorActive() && c.overloaded {
			dur = 360
		}
		c.generationUntil = c.Core.F + dur
		if !c.generationRunning {
			c.generationRunning = true
			c.QueueCharTask(c.generateBeacon, 90)
		}
	}, "mitya-beacons")
	for _, ch := range c.Core.Player.Chars() {
		ally := ch

		ally.AddStatMod(character.StatMod{Base: modifier.NewBase("mitya-core-buffs", -1), AffectedStat: attributes.NoStat, Extra: true, Amount: func() []float64 {
			if !c.reactorActive() {
				return nil
			}
			m := make([]float64, attributes.EndStatType)
			if c.Base.Ascension >= 1 && !c.overloaded {
				m[attributes.EM] += 30 * float64(c.beaconCount())
			}
			if c.Base.Cons >= 1 && c.Core.Player.Active() == ally.Index() {
				m[attributes.CD] += .5
			}
			if c.Base.Cons >= 2 {
				if !c.overloaded {
					m[attributes.EM] += 100
				} else if c.Core.Player.Active() == ally.Index() {
					m[attributes.EM] += 200
				}
			}
			return m
		}})
	}
	if c.Base.Ascension >= 4 {
		c.AddStatMod(character.StatMod{Base: modifier.NewBase("mitya-a4", -1), AffectedStat: attributes.CR, Amount: func() []float64 {
			n := 0
			for _, e := range c.critStacks {
				if e > c.Core.F {
					n++
				}
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.CR] = .05 * float64(n)
			return m
		}})
	}
	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		a := args[0].(*info.AttackEvent)
		if a.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
			return
		}
		a.Info.BaseDmgBonus += min(.14, c.Stat(attributes.EM)*.00028)
		if c.Base.Cons >= 6 {
			a.Info.Elevation += .2
		}
	}, "mitya-star-base-and-elevation")
}
func (c *char) generateBeacon() {
	if c.Core.F > c.generationUntil {
		c.generationRunning = false
		return
	}
	c.addBeacon()
	if c.Core.F+90 <= c.generationUntil {
		c.QueueCharTask(c.generateBeacon, 90)
	} else {
		c.generationRunning = false
	}
}
func (c *char) refreshShred() {
	if c.Base.Cons < 2 {
		return
	}
	for _, t := range c.Core.Combat.Enemies() {
		e, ok := t.(*enemy.Enemy)
		if !ok {
			continue
		}
		for _, ele := range []attributes.Element{attributes.Cryo, attributes.Electro} {
			value := 0.0
			dur := 1
			if c.reactorActive() && c.beaconCount() > 0 {
				value = -.2
				dur = c.reactorUntil - c.Core.F
			}
			e.AddResistMod(info.ResistMod{Base: modifier.NewBase("mitya-c2-"+ele.String(), dur), Ele: ele, Value: value})
		}
	}
}
