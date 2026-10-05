package prune

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initReactionBuffs() {
	for ev := event.ReactionEventStartDelim + 1; ev < event.ReactionEventEndDelim; ev++ {
		reaction := ev
		c.Core.Events.Subscribe(ev, func(args ...any) {
			atk := args[1].(*info.AttackEvent)
			if atk.Info.ActorIndex < 0 || atk.Info.ActorIndex >= len(c.Core.Player.Chars()) {
				return
			}
			trigger := c.Core.Player.ByIndex(atk.Info.ActorIndex)
			if !trigger.StatusIsActive("prune-tolling-rally") {
				return
			}
			if c.IsHexerei && c.Core.Player.GetHexereiCount() >= 2 && trigger.IsHexerei {
				c.grantAttack(c.CharWrapper, "prune-hex-self", attributes.ATKP, .60)
				switch reaction {
				case event.OnSwirlPyro, event.OnSwirlHydro, event.OnSwirlElectro, event.OnSwirlCryo, event.OnStarDiffusion:
					c.grantAttack(trigger, "prune-hex-trigger", attributes.ATKP, .30)
				}
			}
			if c.Base.Cons < 6 {
				return
			}
			c.grantAttack(c.CharWrapper, "prune-c6-self", attributes.ATK, 350)
			active := c.Core.Player.ActiveChar()
			if active.Index() != c.Index() && active.StatusIsActive("prune-tolling-rally") {
				c.grantAttack(active, "prune-c6-active", attributes.ATK, 350)
			}
		}, "prune-reaction-buffs")
	}
}

func (c *char) grantAttack(target *character.CharWrapper, key string, stat attributes.Stat, value float64) {
	buff := make([]float64, attributes.EndStatType)
	buff[stat] = value
	target.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag(key, 5*60), AffectedStat: stat, Amount: func() []float64 { return buff }})
}
