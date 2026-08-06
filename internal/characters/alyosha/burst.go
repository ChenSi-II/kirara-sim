package alyosha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const hunterAdvanceKey = "alyosha-hunter-advance"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	dur := int(burstParam[2][lvl] * 60)
	if c.Base.Cons >= 2 {
		dur += 6 * 60
	}
	c.SetCD(action.ActionBurst, int(burstParam[3][lvl]*60))
	c.ConsumeEnergy(18)
	c.burstSrc = c.Core.F
	src := c.burstSrc
	c.AddStatus(hunterAdvanceKey, dur, true)
	for delay := 30; delay <= dur; delay += 2 * 60 {
		c.QueueCharTask(c.huntingFieldTick(src), delay)
		c.QueueCharTask(c.tugarinTick(src), delay+12)
	}
	f := frames.InitAbilSlice(72)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 72, CanQueueAfter: 60, State: action.BurstState}, nil
}

func (c *char) huntingFieldTick(src int) func() {
	return func() {
		if src != c.burstSrc || !c.StatusIsActive(hunterAdvanceKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Fulgurite Hunting Field", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Electro, Durability: 25, Mult: burst[0][c.TalentLvlBurst()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
	}
}

func (c *char) tugarinTick(src int) func() {
	return func() {
		if src != c.burstSrc || !c.StatusIsActive(hunterAdvanceKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Tugarin", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Electro, Durability: 25, Mult: burst[1][c.TalentLvlBurst()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
		c.activateMark()
		if c.Base.Ascension >= 1 {
			c.healActive(1.2, "Awakened by the Baying Hounds")
		}
		if c.Base.Cons >= 4 {
			c.healLowest(.6, "Harvest the Spoils")
		}
	}
}

func (c *char) healActive(scale float64, message string) {
	c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: c.Core.Player.Active(), Message: message, Src: scale * c.TotalAtk(), Bonus: c.Stat(attributes.Heal)})
}

func (c *char) healLowest(scale float64, message string) {
	target := c.Core.Player.Active()
	lowest := 2.0
	for _, ch := range c.Core.Player.Chars() {
		ratio := ch.CurrentHPRatio()
		if ratio < lowest {
			lowest, target = ratio, ch.Index()
		}
	}
	c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: target, Message: message, Src: scale * c.TotalAtk(), Bonus: c.Stat(attributes.Heal)})
}
