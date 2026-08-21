package zibai

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func (c *char) initAscensions() {
	c.Core.Flags.Custom[reactable.LunarCrystallizeEnableKey] = 1
	c.Core.Events.Subscribe(event.OnMoondriftHarmony, func(...any) {
		c.AddStatus("zibai-selenic-descent", 4*60, true)
	}, "zibai-selenic-descent-harmony")
	geo, hydro := 0, 0
	for _, ch := range c.Core.Player.Chars() {
		if ch.Index() != c.Index() && ch.Base.Element == attributes.Geo {
			geo++
		}
		if ch.Base.Element == attributes.Hydro {
			hydro++
		}
	}
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.DEFP], buff[attributes.EM] = .15*float64(geo), 60*float64(hydro)
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("zibai-a4-party", -1), Amount: func() []float64 { return buff }})
	apply := func(atk *info.AttackEvent) {
		if atk.Info.AttackTag == attacks.AttackTagDirectLunarCrystallize || atk.Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
			atk.Info.BaseDmgBonus += min(c.TotalDef(false)/100*.007, .14)
			if c.Base.Cons >= 6 && c.StatusIsActive("zibai-c6-elevation") {
				atk.Info.Elevation += c.c6Elevation
			}
		}
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) { apply(args[1].(*info.AttackEvent)) }, "zibai-lunar-base")
	c.Core.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) { apply(args[1].(*info.AttackEvent)) }, "zibai-lunar-reaction-base")
	last := -4 * 60
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(...any) {
		if c.Core.Player.GetMoonsignLevel() >= 2 && c.Core.F-last >= 4*60 {
			last = c.Core.F
			c.addPhase(35)
		}
	}, "zibai-phase-lunar-crystallize")
}
