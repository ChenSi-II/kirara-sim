package linnea

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func (c *char) initAscensions() {
	c.moonsignBenedictionInit()
	c.a1Init()
	c.a4Init()
}

func (c *char) moonsignBenedictionInit() {
	c.Core.Flags.Custom[reactable.LunarCrystallizeEnableKey] = 1

	apply := func(atk *info.AttackEvent) {
		if atk.Info.AttackTag != attacks.AttackTagDirectLunarCrystallize &&
			atk.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
			return
		}
		atk.Info.BaseDmgBonus += min(c.TotalDef(false)/100*0.007, 0.14)
	}
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		apply(args[1].(*info.AttackEvent))
	}, "linnea-lunar-crystallize-base-bonus-direct")
	c.Core.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) {
		apply(args[1].(*info.AttackEvent))
	}, "linnea-lunar-crystallize-base-bonus-reaction")
}

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	c.QueueCharTask(c.a1Tick, 1)
}

func (c *char) a1Tick() {
	if c.StatusIsActive(lumiKey) {
		res := -0.15
		if c.Core.Player.GetMoonsignLevel() >= 2 {
			res = -0.30
		}
		// Combat targets model no separate Lumi-centered distance query. Lumi
		// attacks the same nearby target set, so refresh the debuff on enemies.
		for _, target := range c.Core.Combat.Enemies() {
			e, ok := target.(*enemy.Enemy)
			if !ok {
				continue
			}
			e.AddResistMod(info.ResistMod{
				Base:  modifier.NewBase("linnea-a1-geo-res", 90),
				Ele:   attributes.Geo,
				Value: res,
			})
		}
	}
	c.QueueCharTask(c.a1Tick, 60)
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		target := ch
		buff := make([]float64, attributes.EndStatType)
		target.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("linnea-a4-em", -1),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				active := c.Core.Player.ActiveChar()
				if active.Moonsign > 0 {
					if target.Index() != active.Index() {
						return nil
					}
				} else if target.Index() != c.Index() {
					return nil
				}
				buff[attributes.EM] = 0.05 * c.TotalDef(false)
				return buff
			},
		})
	}
}
