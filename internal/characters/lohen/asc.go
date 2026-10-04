package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if !c.StatusIsActive(masterstrokeKey) {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == c.Index() {
			return
		}
		key := fmt.Sprintf("lohen-will-icd-%d", atk.Info.ActorIndex)
		if c.StatusIsActive(key) {
			return
		}
		c.AddStatus(key, 6, true)
		baseATK := c.Stat(attributes.BaseATK)
		gain := 1
		if args[2].(float64) >= skillParam[14][c.skillLevel()]*baseATK {
			gain = int(skillParam[15][c.skillLevel()])
		}
		if c.Base.Ascension >= 1 && args[2].(float64) >= 30*baseATK {
			gain += 60
		}
		if c.Base.Cons >= 1 {
			gain *= 5
		}
		limit := 100
		if c.Base.Cons >= 1 {
			limit = 300
		}
		c.will = min(limit, c.will+gain)
	}, "lohen-will-to-win")
	if c.Base.Ascension >= 4 {
		hook := func(args ...any) {
			if !c.StatusIsActive(masterstrokeKey) {
				return
			}
			atk := args[1].(*info.AttackEvent)
			if atk.Info.ActorIndex == c.Index() {
				return
			}
			buff := make([]float64, attributes.EndStatType)
			buff[attributes.ATKP] = .15
			c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("lohen-a4-self", 8*60), AffectedStat: attributes.ATKP, Amount: func() []float64 { return buff }})
			if atk.Info.ActorIndex >= 0 && atk.Info.ActorIndex < len(c.Core.Player.Chars()) {
				c.Core.Player.Chars()[atk.Info.ActorIndex].AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("lohen-a4-trigger", 8*60), AffectedStat: attributes.ATKP, Amount: func() []float64 { return buff }})
			}
		}
		for _, evt := range []event.Event{event.OnMelt, event.OnSuperconduct, event.OnFrozen, event.OnSwirlCryo, event.OnCrystallizeCryo} {
			c.Core.Events.Subscribe(evt, hook, fmt.Sprintf("lohen-a4-%d", evt))
		}
	}
}
