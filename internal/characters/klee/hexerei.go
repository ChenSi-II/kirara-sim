package klee

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) hexActive() bool { return c.IsHexerei && c.Core.Player.GetHexereiCount() >= 2 }

func (c *char) gainSpark() {
	if c.Base.Ascension < 1 {
		return
	}
	c.sparks = min(c.sparks+1, 3)
	c.AddStatus(a1SparkKey, -1, false)
}

func (c *char) consumeSpark() {
	if c.sparks == 0 {
		return
	}
	if c.Base.Cons >= 6 && c.hexActive() && c.Core.Rand.Float64() < .5 {
		return
	}
	c.sparks--
	if c.sparks == 0 {
		c.DeleteStatus(a1SparkKey)
	}
}

func (c *char) medalMultiplier() float64 {
	if !c.hexActive() {
		return 1
	}
	n := 0
	for _, key := range []string{"klee-medal-normal", "klee-medal-skill", "klee-medal-burst"} {
		if c.StatusIsActive(key) {
			n++
		}
	}
	return []float64{1, 1.15, 1.3, 1.5}[n]
}

func (c *char) hexInit() {
	if !c.IsHexerei {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != c.Index() || args[2].(float64) <= 0 {
			return
		}
		key := ""
		switch a.Info.AttackTag {
		case attacks.AttackTagNormal:
			key = "klee-medal-normal"
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
			key = "klee-medal-skill"
			// Every part of Jumpy Dumpty now applies C2, not just its mines.
			c.c2(info.AttackCB{Target: args[0].(info.Target), AttackEvent: a})
		case attacks.AttackTagElementalBurst:
			key = "klee-medal-burst"
		}
		if key != "" && c.hexActive() {
			c.AddStatus(key, 20*60, false)
		}
	}, "klee-hexerei")
}

func (c *char) boomBarrage(delay int) {
	c.consumeSpark()
	ai := info.AttackInfo{
		ActorIndex: c.Index(), Abil: "Boom-Boom Barrage", AttackTag: attacks.AttackTagExtra,
		ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt,
		Element: attributes.Pyro, Durability: 25, PoiseDMG: 180, Mult: charge[c.TalentLvlAttack()] * c.medalMultiplier(),
	}
	snap := c.Snapshot(&ai)
	snap.Stats[attributes.DmgP] += .5
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, c.makeA4CB())
}
