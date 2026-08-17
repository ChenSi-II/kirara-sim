package jahoda

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const robotsKey = "jahoda-assistance-robots"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Hidden Aces", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: burst[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 34, 34)
	dur := int(burstParam[2][lvl] * 60)
	c.robotSrc = c.Core.F
	c.AddStatus(robotsKey, dur, true)
	if c.Base.Cons >= 4 {
		c.AddEnergy("jahoda-c4", 4)
	}
	interval := 2 * 60
	count, damageScale, healScale := c.elementalRobotBonus()
	if count == 4 {
		interval = 108
	}
	for delay := 70; delay <= dur; delay += interval {
		for robot := 0; robot < count; robot++ {
			c.QueueCharTask(c.robotTick(c.robotSrc, damageScale, healScale), delay+robot*6)
		}
	}
	c.SetCD(action.ActionBurst, int(burstParam[7][lvl]*60))
	c.ConsumeEnergy(34)
	f := frames.InitAbilSlice(76)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 76, CanQueueAfter: 64, State: action.BurstState}, nil
}

func (c *char) elementalRobotBonus() (int, float64, float64) {
	count, damage, heal := 2, 1.0, 1.0
	if c.Base.Ascension < 1 {
		return count, damage, heal
	}
	counts := map[attributes.Element]int{}
	for _, ch := range c.Core.Player.Chars() {
		counts[ch.Base.Element]++
	}
	best, bestN := attributes.Pyro, -1
	for _, ele := range []attributes.Element{attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo} {
		if counts[ele] > bestN {
			best, bestN = ele, counts[ele]
		}
	}
	switch best {
	case attributes.Pyro:
		damage = 1.3
	case attributes.Hydro:
		heal = 1.2
	case attributes.Electro:
		count++
	}
	return count, damage, heal
}

func (c *char) robotTick(src int, damageScale, healScale float64) func() {
	return func() {
		if src != c.robotSrc || !c.StatusIsActive(robotsKey) {
			return
		}
		lvl := c.TalentLvlBurst()
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Purrsonal Assistance Robot", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Anemo, Durability: 25, Mult: burst[1][lvl] * damageScale}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
		heal := (burstParam[3][lvl]*c.TotalAtk() + burstParam[4][lvl]) * healScale
		active := c.Core.Player.ActiveChar()
		c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: active.Index(), Message: "Assistance Robot", Src: heal, Bonus: c.Stat(attributes.Heal)})
		if c.Base.Ascension >= 4 && active.CurrentHPRatio() > .7 {
			buff := make([]float64, attributes.EndStatType)
			buff[attributes.EM] = 100
			active.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("jahoda-a4", 6*60), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
		}
	}
}
