package prune

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	register := func(evt event.Event, ele attributes.Element) {
		c.Core.Events.Subscribe(evt, func(args ...any) {
			if len(args) < 2 {
				return
			}
			atk, ok := args[1].(*info.AttackEvent)
			if !ok || atk.Info.ActorIndex != c.Index() {
				return
			}
			if atk.Info.Abil == "Hexhunter Chime" && atk.Info.Element == attributes.Anemo {
				// Use the actual reaction, including when it consumes all aura.
				// Keep the first successful conversion for this E window.
				if !c.StatusIsActive(conversionKey) {
					c.converted = ele
					c.AddStatus(conversionKey, 6*60, true)
				}
				return
			}
			if c.Base.Ascension < 1 || !c.StatusIsActive(bellKey) || atk.Info.Abil != "Witchlure Bell" {
				return
			}
			target := args[0].(info.Target)
			// Image 7: the passive hammer lands 0.800s after the bell hit.
			c.QueueCharTask(func() { c.oathhammer(ele, target) }, 48)
		}, fmt.Sprintf("prune-a1-%d", evt))
	}
	register(event.OnSwirlPyro, attributes.Pyro)
	register(event.OnSwirlHydro, attributes.Hydro)
	register(event.OnSwirlElectro, attributes.Electro)
	register(event.OnSwirlCryo, attributes.Cryo)
	register(event.OnStarDiffusion, attributes.Cryo)
}

// Each summoned hammer has its own hit callback and may bounce only once.
func (c *char) oathhammer(ele attributes.Element, target info.Target) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Banehunter Oathhammer", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: ele, Mult: 1.5}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(target, nil, 3), 0, 0, c.convertedHammerHit(ele, attacks.AttackTagElementalBurst, "Banehunter Oathhammer Bounce"))
}

func (c *char) convertedHammerHit(ele attributes.Element, tag attacks.AttackTag, bounceName string) info.AttackCBFunc {
	bounced := false
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		c.tollingRally()
		if c.Base.Cons >= 2 && c.StatusIsActive(bellKey) {
			c.c2Stacks = min(6, c.c2Stacks+1)
		}
		if c.Base.Cons >= 1 && !c.StatusIsActive("prune-c1-icd") {
			c.AddStatus("prune-c1-icd", 108, true)
			c.AddEnergy("prune-c1", 2)
		}
		if c.Base.Cons < 4 || bounced || bounceName == "" {
			return
		}
		bounced = true
		bounce := info.AttackInfo{ActorIndex: c.Index(), Abil: bounceName, AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: ele, Mult: .80}
		c.Core.QueueAttack(bounce, combat.NewCircleHitOnTarget(a.Target, nil, 3), 64, 64, c.convertedHammerHit(ele, tag, ""))
	}
}

func (c *char) tollingRally() {
	if c.Base.Ascension < 4 {
		return
	}
	bonus := min(max(c.TotalAtk()-2000, 0)*.00025, .50)
	for _, ch := range c.Core.Player.Chars() {
		if ch.Index() == c.Index() {
			continue
		}
		ch.AddStatus("prune-tolling-rally", 5*60, true)
		ch.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("prune-tolling-rally-damage", 5*60), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge, attacks.AttackTagElementalArt, attacks.AttackTagElementalBurst:
			default:
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.DmgP] = bonus
			return out
		}})
	}
}
