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
	c.Core.Events.Subscribe(event.OnApplyAttack, c.nightingaleBuff, "illuga-nightingale")
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
		ch.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("illuga-lightkeepers-oath", 20*60), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.Element != attributes.Geo {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.CR], out[attributes.CD] = .05, .10
			if c.Core.Player.GetMoonsignLevel() >= 2 {
				out[attributes.EM] = 50
			}
			if c.Base.Cons >= 6 {
				out[attributes.CR], out[attributes.CD] = .10, .30
				if c.Core.Player.GetMoonsignLevel() >= 2 {
					out[attributes.EM] = 80
				}
			}
			return out
		}})
	}
}

func (c *char) nightingaleBuff(args ...any) {
	if c.nightingaleStacks == 0 || !c.StatusIsActive(orioleSongKey) {
		return
	}
	atk := args[0].(*info.AttackEvent)
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
	atk.Info.FlatDmg += (burstParam[2][lvl] + extra) * c.Stat(attributes.EM)
	c.nightingaleStacks--
	c.consumedStacks++
	if c.Base.Cons >= 2 && c.consumedStacks%7 == 0 {
		c.c2Attack()
	}
}
