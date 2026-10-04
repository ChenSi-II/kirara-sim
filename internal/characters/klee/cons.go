package klee

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c1(delay int) {
	if c.Base.Cons < 1 {
		return
	}
	// 0.1 base change, + 0.08 every failure
	if c.Core.Rand.Float64() > c.c1Chance {
		// failed
		c.c1Chance += 0.08
		return
	}
	c.c1Chance = 0.1
	if c.IsHexerei {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.ATKP] = .6
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("klee-c1-atk", 12*60), AffectedStat: attributes.ATKP, Amount: func() []float64 { return buff }})
	}

	ai := info.AttackInfo{
		ActorIndex:         c.Index(),
		Abil:               "Sparks'n'Splash (C1)",
		AttackTag:          attacks.AttackTagElementalBurst,
		ICDTag:             attacks.ICDTagElementalBurst,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeDefault,
		Element:            attributes.Pyro,
		Durability:         25,
		Mult:               1.2 * burst[c.TalentLvlBurst()],
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}
	// TODO: should center on target hit by attack that triggered c1
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1.5), 0, delay)
}

func (c *char) c2(a info.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	reduction := -.233
	if c.IsHexerei {
		reduction = -.23
	}
	e.AddDefMod(info.DefMod{
		Base:  modifier.NewBaseWithHitlag("kleec2", 10*60),
		Value: reduction,
	})
}
