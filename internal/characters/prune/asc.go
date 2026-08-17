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
	if c.Base.Ascension < 1 {
		return
	}
	register := func(evt event.Event, ele attributes.Element) {
		c.Core.Events.Subscribe(evt, func(...any) { c.oathhammer(ele) }, fmt.Sprintf("prune-a1-%d", evt))
	}
	register(event.OnSwirlPyro, attributes.Pyro)
	register(event.OnSwirlHydro, attributes.Hydro)
	register(event.OnSwirlElectro, attributes.Electro)
	register(event.OnSwirlCryo, attributes.Cryo)
	register(event.OnStarDiffusion, attributes.Cryo)
}

func (c *char) oathhammer(ele attributes.Element) {
	if !c.StatusIsActive(bellKey) {
		return
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Banehunter Oathhammer", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: ele, Mult: 1.5}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
	c.converted = ele
	c.tollingRally()
	if c.Base.Cons >= 1 && !c.StatusIsActive("prune-c1-icd") {
		c.AddStatus("prune-c1-icd", 108, true)
		c.AddEnergy("prune-c1", 2)
	}
}

func (c *char) tollingRally() {
	if c.Base.Ascension < 4 {
		return
	}
	bonus := min(max(c.TotalAtk()-2000, 0)*.00025, .50)
	for _, ch := range c.Core.Player.Chars() {
		ch.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("prune-tolling-rally", 5*60), Amount: func(*info.AttackEvent, info.Target) []float64 {
			out := make([]float64, attributes.EndStatType)
			out[attributes.DmgP] = bonus
			return out
		}})
	}
}
