package cyno

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const sunriseKey = "cyno-shared-sunrise"

func (c *char) enhancedInit() {
	c.sunriseStacks = make(map[int]int)
	if !c.Enhanced || c.Base.Cons < 2 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("cyno-c2-stellar", -1), Amount: func(ai info.AttackInfo) float64 {
			if ch.StatusIsActive(sunriseKey) && ai.AttackTag == attacks.AttackTagReactionStarSuperconduct {
				return .16 * float64(c.sunriseStacks[ch.Index()])
			}
			return 0
		}})
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		ch := c.Core.Player.ByIndex(ai.ActorIndex)
		if !ch.StatusIsActive(sunriseKey) || ch.StatusIsActive("cyno-sunrise-c2-icd") {
			return
		}
		if ai.AttackTag != attacks.AttackTagNormal && ai.AttackTag != attacks.AttackTagExtra {
			return
		}
		ch.AddStatus("cyno-sunrise-c2-icd", 6, true)
		c.sunriseStacks[ch.Index()] = min(5, c.sunriseStacks[ch.Index()]+1)
	}, "cyno-c2-stellar")
}

func (c *char) enterPactsworn(duration int, fromSkill bool) {
	c.burstExtension = 0
	c.burstFromSkill = fromSkill
	c.c4Counter = 0
	c.c6Stacks = 0
	c.DeleteStatus(c6Key)
	c.DeleteStatus(a1Key)
	c.burstSrc++
	src := c.burstSrc
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.EM] = 100
	c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag(burstKey, duration), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
	if c.Base.Ascension >= 1 {
		c.QueueCharTask(func() {
			if src == c.burstSrc {
				c.a1()
			}
		}, 328)
	}
	var expiry func()
	expiry = func() {
		if src != c.burstSrc {
			return
		}
		if c.StatusIsActive(burstKey) {
			c.QueueCharTask(expiry, c.StatusDuration(burstKey)+1)
			return
		}
		c.onBurstExpiry(src)
	}
	c.QueueCharTask(expiry, duration+1)
	if c.Base.Cons >= 1 && c.SuperconductRadiance() {
		for _, ch := range c.Core.Player.Chars() {
			ch.DeleteStatus(sunriseKey)
		}
		c.giveSunrise(c.CharWrapper, duration)
	}
}

func (c *char) giveSunrise(ch *character.CharWrapper, duration int) {
	if duration <= 0 {
		return
	}
	c.sunriseStacks[ch.Index()] = 0
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.EM] = 200
	ch.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag(sunriseKey, duration), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
}

func (c *char) transferSunrise(next int) {
	if !c.StatusIsActive(sunriseKey) {
		return
	}
	duration := c.StatusDuration(burstKey)
	c.DeleteStatus(sunriseKey)
	c.giveSunrise(c.Core.Player.ByIndex(next), duration)
}

func (c *char) extendPactsworn(duration int) {
	c.ExtendStatus(burstKey, duration)
	if c.StatusIsActive(sunriseKey) {
		c.ExtendStatus(sunriseKey, duration)
	}
}

func (c *char) stellarBolt(ai *info.AttackInfo, stellar bool) {
	if !stellar {
		return
	}
	ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
	ai.Mult = 2
	if c.Base.Ascension >= 4 {
		ai.FlatDmg = 6 * c.Stat(attributes.EM)
	}
}
