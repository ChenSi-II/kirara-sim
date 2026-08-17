package iansan

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 1 {
		c.Core.Events.Subscribe(event.OnNightsoulConsume, func(args ...any) {
			if args[0].(int) != c.Index() || c.StatusIsActive("iansan-c1-icd") {
				return
			}
			if args[1].(float64) >= 6 {
				c.AddStatus("iansan-c1-icd", 18*60, true)
				c.AddEnergy("iansan-c1", 15)
			}
		}, "iansan-c1")
	}
	if c.Base.Cons >= 4 {
		c.Core.Events.Subscribe(event.OnBurst, func(...any) {
			if c.StatusIsActive(scaleKey) && c.Core.Player.Active() != c.Index() {
				c.surging = 2
			}
		}, "iansan-c4")
	}
}

func (c *char) c6Buff() {
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		target.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("iansan-c6", 3*60), Amount: func(*info.AttackEvent, info.Target) []float64 {
			if target.Index() != c.Core.Player.Active() {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.DmgP] = .25
			return out
		}})
	}
}
