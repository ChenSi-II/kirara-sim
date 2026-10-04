package wriothesley

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) enhancedInit() {
	if !c.Enhanced {
		return
	}
	c.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("wriothesley-stellar", -1), Amount: func(ai info.AttackInfo) float64 {
		if !c.SuperconductRadiance() || ai.AttackTag != attacks.AttackTagReactionStarSuperconduct {
			return 0
		}
		bonus := .3
		if (ai.Abil == "Normal 4 (Enhanced)" || ai.Abil == "Normal 4 (Enhanced) (C6)") && c.StatusIsActive("wriothesley-c1-n5-dmg") {
			bonus += .5
		}
		if (ai.Abil == "Stellar: Vaulting Fist" || ai.Abil == "Stellar: Vaulting Fist (C6)") && c.StatusIsActive("wriothesley-c1-charge") {
			bonus += .5
		}
		return bonus
	}})
	if c.Base.Cons < 4 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.AtkSpd] = .1
		if ch.Index() == c.Index() {
			buff[attributes.AtkSpd] = .2
		}
		ch.AddStatMod(character.StatMod{Base: modifier.NewBase("wriothesley-c4-stellar", -1), AffectedStat: attributes.AtkSpd, Amount: func() []float64 {
			if c.SuperconductRadiance() {
				return buff
			}
			return nil
		}})
	}
	c.AddDamageReductionMod(character.DamageReductionMod{Base: modifier.NewBase("wriothesley-c4-stellar-reduction", -1), Amount: func() float64 {
		if c.SuperconductRadiance() {
			return .25
		}
		return 0
	}})
}

func (c *char) maxEdict() bool {
	return c.Enhanced && c.Base.Cons >= 2 && c.Base.Ascension >= 4 && c.StatusIsActive(skillKey) && c.a4Stack >= 5
}

func (c *char) enhanceSnapshot(ai *info.AttackInfo, snap *info.Snapshot) {
	if !c.Enhanced {
		return
	}
	if !c.SuperconductRadiance() {
		if c.maxEdict() {
			switch ai.AttackTag {
			case attacks.AttackTagNormal:
				ai.Mult *= 1.25
			case attacks.AttackTagExtra:
				ai.Mult *= 1.3
			}
		}
		return
	}
	if ai.AttackTag != attacks.AttackTagNormal || !c.skillBuffActive() {
		return
	}
	if c.Base.Cons >= 6 && c.Base.Ascension >= 1 {
		c.addC6Buff(snap)
	}
	ratio := 0.0
	switch ai.Abil {
	case "Normal 2 (Enhanced)":
		ratio = .6
	case "Normal 4 (Enhanced)":
		ratio = .8
		if c.StatusIsActive("wriothesley-c1-n5") {
			c.DeleteStatus("wriothesley-c1-n5")
			c.AddStatus("wriothesley-c1-n5-dmg", 1, false)
		}
	}
	if ratio == 0 {
		return
	}
	if c.maxEdict() {
		ratio *= 1.5
	}
	ai.Mult *= ratio
	ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
}

func (c *char) radiantCharge(ai *info.AttackInfo, snap *info.Snapshot) {
	ai.Abil = "Stellar: Vaulting Fist"
	ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
	ai.HitlagFactor = .03
	ai.HitlagHaltFrames = .12 * 60
	if c.maxEdict() {
		ai.Mult *= 1.5
	}
	if c.Base.Cons >= 6 {
		c.addC6Buff(snap)
	}
}

func (c *char) radiantChargeCB() info.AttackCBFunc {
	done := false
	return func(a info.AttackCB) {
		if done || a.Target.Type() != info.TargettableEnemy {
			return
		}
		done = true
		if !c.StatusIsActive("wriothesley-stellar-heal-icd") {
			c.AddStatus("wriothesley-stellar-heal-icd", 120, true)
			c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: c.Index(), Message: "Stellar: Vaulting Fist", Src: .3 * c.MaxHP(), Bonus: c.Stat(attributes.Heal)})
		}
		if c.Base.Cons >= 1 {
			c.AddStatus("wriothesley-c1-n5", 300, true)
			if !c.c1SkillExtensionProc && c.StatusIsActive(skillKey) {
				c.ExtendStatus(skillKey, c1SkillExtension)
				c.ExtendStatus("wriothesley-a4", c1SkillExtension)
				c.c1SkillExtensionProc = true
			}
		}
		c.radiantIcicle(a)
	}
}

func (c *char) radiantNormalCB(n int) info.AttackCBFunc {
	done := false
	return func(a info.AttackCB) {
		if done || n != 4 || a.Target.Type() != info.TargettableEnemy || a.AttackEvent.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
			return
		}
		done = true
		if c.Base.Cons >= 1 {
			c.AddStatus("wriothesley-c1-charge", 300, true)
		}
		c.radiantIcicle(a)
	}
}

func (c *char) radiantIcicle(a info.AttackCB) {
	if c.Base.Cons < 6 || c.Base.Ascension < 1 {
		return
	}
	ai := a.AttackEvent.Info
	ai.Abil += " (C6)"
	ai.Mult *= .2
	// External flat-damage buffs apply independently to the icicle.
	ai.FlatDmg = 0
	ai.Durability = 0
	ai.HitlagFactor = 0
	ai.HitlagHaltFrames = 0
	snap := c.Snapshot(&ai)
	c.addC6Buff(&snap)
	c.Core.QueueAttackWithSnap(ai, snap, combat.NewCircleHitOnTarget(a.Target, nil, 2), 0)
}
