package illuga

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		// Reaction contributions already consumed stacks before aggregation.
		if args[1].(*info.AttackEvent).Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
			return
		}
		c.nightingaleBuff(args...)
	}, "illuga-nightingale")
	c.Core.Events.Subscribe(event.OnLunarReactionAttack, c.nightingaleBuff, "illuga-nightingale-contributor")
	c.Core.Events.Subscribe(event.OnConstructSpawned, func(args ...any) {
		if !c.StatusIsActive(orioleSongKey) || c.constructStacks >= 15 || len(args) == 0 {
			return
		}
		_, ok := args[0].(*construct.Construct)
		if !ok {
			return
		}
		gain := min(5, 15-c.constructStacks)
		c.nightingaleStacks += gain
		c.constructStacks += gain
	}, "illuga-nightingale-construct")
}

func (c *char) lightkeepersOath() {
	if c.Base.Ascension < 1 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		if ch.Index() == c.Index() {
			continue
		}
		ch.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("illuga-oath-em", 20*60), AffectedStat: attributes.EM, Amount: func() []float64 {
			if c.Core.Player.GetMoonsignLevel() < 2 {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.EM] = 50
			if c.Base.Cons >= 6 {
				out[attributes.EM] = 80
			}
			return out
		}})
		ch.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("illuga-lightkeepers-oath", 20*60), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.Element != attributes.Geo {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.CR], out[attributes.CD] = .05, .10

			if c.Base.Cons >= 6 {
				out[attributes.CR], out[attributes.CD] = .10, .30

			}
			return out
		}})
	}
}

func (c *char) nightingaleBuff(args ...any) {
	if c.nightingaleStacks == 0 || !c.StatusIsActive(orioleSongKey) {
		return
	}
	atk := args[1].(*info.AttackEvent)
	if atk.Info.ActorIndex != c.Core.Player.Active() || atk.Info.Element != attributes.Geo || atk.Info.AttackTag == attacks.AttackTagNone {
		return
	}
	lvl := c.TalentLvlBurst()
	extra := 0.0
	if c.Base.Ascension >= 4 {
		count := 0
		for _, ch := range c.Core.Player.Chars() {
			if ch.Base.Element == attributes.Hydro || ch.Base.Element == attributes.Geo {
				count++
			}
		}
		tiers := []float64{0, .07, .14, .24}
		extra = tiers[min(count, 3)]
		if atk.Info.AttackTag == attacks.AttackTagDirectLunarCrystallize || atk.Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
			extra = []float64{0, .48, .96, 1.60}[min(count, 3)]
		}
	}
	param := 2
	if atk.Info.AttackTag == attacks.AttackTagDirectLunarCrystallize || atk.Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
		param = 3
	}
	atk.Info.FlatDmg += (burstParam[param][lvl] + extra) * c.Stat(attributes.EM)
	c.nightingaleStacks--
	c.consumedStacks++
	if c.nightingaleStacks == 0 {
		c.DeleteStatus(orioleSongKey)
	}
	if c.Base.Cons >= 2 && c.consumedStacks%7 == 0 {
		c.c2Attack()
	}
}
