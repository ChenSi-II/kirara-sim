package linnea

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 1 {
		c.Core.Events.Subscribe(event.OnMoondriftHarmony, func(...any) {
			c.onMoondriftHarmony()
		}, "linnea-harmony")
		c.Core.Events.Subscribe(event.OnLunarReactionAttack, c.c1LunarCrystallize, "linnea-c1")
		c.Core.Events.Subscribe(event.OnApplyAttack, c.c1DirectLunarCrystallize, "linnea-c1-direct")
	}
	if c.Base.Cons >= 6 {
		c.c6Init()
	}
}

func (c *char) c1OnSkill() {
	if c.Base.Cons < 1 {
		return
	}
	c.addFieldCatalog()
}

func (c *char) addFieldCatalog() {
	if c.Base.Cons >= 6 {
		c.fieldCatalogStacks = 18
	} else {
		c.fieldCatalogStacks = min(c.fieldCatalogStacks+6, 18)
	}
	c.fieldCatalogSrc = c.Core.F
	src := c.fieldCatalogSrc
	c.AddStatus("linnea-field-catalog", 10*60, true)
	c.QueueCharTask(func() {
		if c.fieldCatalogSrc == src {
			c.fieldCatalogStacks = 0
		}
	}, 10*60)
}

func (c *char) c1LunarCrystallize(args ...any) {
	atk := args[1].(*info.AttackEvent)
	if atk.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
		return
	}
	c.applyFieldCatalogBonus(atk)
}

func (c *char) c1DirectLunarCrystallize(args ...any) {
	atk := args[0].(*info.AttackEvent)
	if atk.Info.AttackTag != attacks.AttackTagDirectLunarCrystallize ||
		atk.Info.Abil == "Million Ton Crush" {
		return
	}
	c.applyFieldCatalogBonus(atk)
}

func (c *char) applyFieldCatalogBonus(atk *info.AttackEvent) {
	if !c.StatusIsActive("linnea-field-catalog") || c.fieldCatalogStacks == 0 {
		return
	}
	consume, bonusScale := 1, 0.75
	if c.Base.Cons >= 6 {
		consume, bonusScale = 2, 0.75*1.5
	}
	consume = min(consume, c.fieldCatalogStacks)
	c.fieldCatalogStacks -= consume
	atk.Info.FlatDmg += bonusScale * c.TotalDef(false)
}

func (c *char) consumeCatalogForMillion() float64 {
	if c.Base.Cons < 1 || !c.StatusIsActive("linnea-field-catalog") {
		return 0
	}
	originalStacks := min(c.fieldCatalogStacks, 5)
	if c.Base.Cons >= 6 {
		originalStacks = min(c.fieldCatalogStacks/2, 5)
		c.fieldCatalogStacks -= originalStacks * 2
		return float64(originalStacks) * 1.5 * 1.5 * c.TotalDef(false)
	}
	c.fieldCatalogStacks -= originalStacks
	return float64(originalStacks) * 1.5 * c.TotalDef(false)
}

func (c *char) onMoondriftHarmony() {
	c.addFieldCatalog()
	if c.Base.Cons >= 2 {
		c.c2HarmonyBuff()
	}
	if c.Base.Cons >= 4 {
		c.c4HarmonyBuff()
	}
}

func (c *char) c2HarmonyBuff() {
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.CD] = 0.40
	for _, ch := range c.Core.Player.Chars() {
		if ch.Base.Element != attributes.Hydro && ch.Base.Element != attributes.Geo {
			continue
		}
		ch.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("linnea-c2-cd", 8*60),
			AffectedStat: attributes.CD,
			Amount: func() []float64 {
				return buff
			},
		})
	}
}

func (c *char) c4HarmonyBuff() {
	selfBuff := make([]float64, attributes.EndStatType)
	selfBuff[attributes.DEFP] = 0.25
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("linnea-c4-self", 5*60),
		AffectedStat: attributes.DEFP,
		Amount: func() []float64 {
			return selfBuff
		},
	})

	active := c.Core.Player.ActiveChar()
	activeBuff := make([]float64, attributes.EndStatType)
	activeBuff[attributes.DEFP] = 0.25
	active.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("linnea-c4-active", 5*60),
		AffectedStat: attributes.DEFP,
		Amount: func() []float64 {
			return activeBuff
		},
	})
}

func (c *char) c2TriggerHarmony() {
	if c.Base.Cons < 2 || c.Core.Player.GetMoonsignLevel() < 2 {
		return
	}
	target := c.Core.Combat.PrimaryTarget()
	if target == nil {
		return
	}
	var contributors [info.MaxChars]bool
	for _, ch := range c.Core.Player.Chars() {
		if ch.Base.Element == attributes.Hydro || ch.Base.Element == attributes.Geo {
			contributors[ch.Index()] = true
		}
	}
	c.Core.Events.Emit(event.OnMoondriftHarmony, target, nil)
	reactable.DoLCrAttackWithContrib(contributors, target, c.Core, c.Index())
}

func (c *char) c6Init() {
	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagDirectLunarCrystallize {
			atk.Info.Elevation += 0.25
		}
	}, "linnea-c6-direct")
	c.Core.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
			atk.Info.Elevation += 0.25
		}
	}, "linnea-c6-reaction")
}
