package albedo

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"strings"
)

type silverBloom struct {
	expiry int
	pos    info.Point
}

func (c *char) liveSilver() []silverBloom {
	n := 0
	for _, s := range c.silver {
		if s.expiry > c.Core.F {
			c.silver[n] = s
			n++
		}
	}
	c.silver = c.silver[:n]
	return c.silver
}

func (c *char) inBlossomArea(t info.Target) bool {
	if c.skillActive {
		return t.IsWithinArea(c.skillArea)
	}
	// Only one Moonsilver substitutes for the lost Solar Isotoma.
	if s := c.liveSilver(); len(s) > 0 {
		return t.IsWithinArea(combat.NewCircleHitOnTarget(s[0].pos, nil, 10))
	}
	return false
}

func (c *char) hexOnSkill() {
	if !c.IsHexerei {
		return
	}
	if c.Core.Player.GetHexereiCount() >= 2 {
		c.AddStatus("albedo-silver-window", 20*60, false)
	}
	if c.Base.Cons >= 1 {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.DEFP] = .5
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("albedo-c1-def", 20*60), AffectedStat: attributes.DEFP, Amount: func() []float64 { return buff }})
	}
}

func (c *char) hexTeamBuff(silver bool) {
	if !c.IsHexerei || c.Core.Player.GetHexereiCount() < 2 {
		return
	}
	key, rate, cap := "albedo-hex-isotoma", .00004, .12
	if silver {
		key, rate, cap = "albedo-hex-silver", .0001, .30
	}
	amount := min(c.TotalDef(false)*rate, cap)
	for _, ch := range c.Core.Player.Chars() {
		if silver && !ch.IsHexerei {
			continue
		}
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.DmgP] = amount
		ch.AddAttackMod(character.AttackMod{Base: modifier.NewBase(key, 20*60), Amount: func(a *info.AttackEvent, _ info.Target) []float64 {
			switch a.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge, attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
				return buff
			}
			return nil
		}})
	}
}

func (c *char) hexInit() {
	if !c.IsHexerei {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != c.Index() {
			return
		}
		if c.Base.Ascension >= 1 && a.Info.Abil == skillAbilTick && len(c.liveSilver()) > 0 {
			a.Info.FlatDmg += 2.4 * c.TotalDef(false)
		}
		if c.StatusIsActive("albedo-c6-blossom") && strings.Contains(a.Info.Abil, "Blossom") {
			a.Info.FlatDmg += 2.5 * c.TotalDef(false)
		}
	}, "albedo-hex-damage")
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != c.Index() || a.Info.AttackTag != attacks.AttackTagElementalArt || !c.StatusIsActive("albedo-silver-window") {
			return
		}
		t := args[0].(info.Target)
		c.liveSilver()
		if len(c.silver) == 2 {
			c.silver = c.silver[1:]
		}
		c.silver = append(c.silver, silverBloom{c.Core.F + 10*60, t.Pos()})
		c.hexTeamBuff(true)
	}, "albedo-create-silver")
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnActionExec, func(args ...any) {
		if args[1].(action.Action) != action.ActionJump {
			return
		}
		for i, s := range c.liveSilver() {
			if !c.Core.Combat.Player().IsWithinArea(combat.NewCircleHitOnTarget(s.pos, nil, 5)) {
				continue
			}
			c.silver = append(c.silver[:i], c.silver[i+1:]...)
			ch := c.Core.Player.ActiveChar()
			buff := make([]float64, attributes.EndStatType)
			buff[attributes.DmgP] = .3
			ch.AddAttackMod(character.AttackMod{Base: modifier.NewBaseWithHitlag("albedo-silver-plunge", 3*60), Amount: func(a *info.AttackEvent, _ info.Target) []float64 {
				if a.Info.AttackTag != attacks.AttackTagPlunge || strings.Contains(strings.ToLower(a.Info.Abil), "collision") {
					return nil
				}
				return buff
			}})
			_ = c.Core.Player.SetAirborne(player.AirborneXianyun)
			break
		}
	}, "albedo-silver-jump")
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.AttackTag != attacks.AttackTagPlunge || strings.Contains(strings.ToLower(a.Info.Abil), "collision") {
			return
		}
		ch := c.Core.Player.ByIndex(a.Info.ActorIndex)
		if ch.StatusIsActive("albedo-silver-plunge") {
			c.Core.Tasks.Add(func() { ch.DeleteStatus("albedo-silver-plunge") }, 6)
		}
	}, "albedo-silver-plunge-consume")
}

func (c *char) hexC2() {
	if !c.IsHexerei || c.Base.Cons < 2 || c.c2stacks < 4 || c.Core.Player.Active() == c.Index() {
		return
	}
	c.c2stacks = 0
	c.DeleteStatus(c2key)
	c.a4()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Fatal Blossom (C2)", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: 3}
	for i := range 3 {
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3), 0, 1+i*5)
	}
}
