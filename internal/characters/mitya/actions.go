package mitya

import (
	"fmt"
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

// All animation/hit/periodic timings below are UNMEASURED placeholders.
// Talent durations/cooldowns are source data; see PLACEHOLDER_FRAMES.md.
const normalHit = 18
const normalFrames = 42
const skillHit = 24
const skillFrames = 50
const holdSkillFrames = 70
const chargeHit = 30
const chargeFrames = 60
const beaconInterval = 60
const steadyInterval = 120
const burstHit = 45
const burstFrames = 90

func actionInfo(length int, state action.AnimationState) action.Info {
	return action.Info{Frames: frames.NewAbilFunc(frames.InitAbilSlice(length)), AnimationLength: length, CanQueueAfter: length, State: state}
}
func (c *char) hit(name string, mult float64, tag attacks.AttackTag, icd attacks.ICDTag, delay int) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: name, AttackTag: tag, ICDTag: icd, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Electro, Durability: 25, Mult: mult, IgnoreInfusion: true}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, delay)
}
func (c *char) starHit(name string, mult float64, useEM bool, delay int, core bool) {
	if core && c.Base.Cons >= 6 && !c.overloaded {
		mult *= 1.15
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: name, AttackTag: attacks.AttackTagReactionStarSuperconduct, ICDTag: attacks.ICDTagNone, Element: attributes.Electro, Mult: mult, UseEM: useEM, IgnoreInfusion: true}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, delay)
}
func (c *char) Attack(map[string]int) (action.Info, error) {
	i := c.NormalCounter
	c.hit(fmt.Sprintf("Normal %d", i+1), normalParams[c.TalentLvlAttack()][i], attacks.AttackTagNormal, attacks.ICDTagNormalAttack, normalHit)
	c.AdvanceNormalIndex()
	return actionInfo(normalFrames, action.NormalAttackState), nil
}
func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.reactorActive() {
		c.finishCore()
	}
	c.reactorSrc++
	src := c.reactorSrc
	c.overloaded = p["hold"] != 0
	c.reactorUntil = c.Core.F + int(skillParams[c.TalentLvlSkill()][5]*60)
	idx, length := 0, skillFrames
	if c.overloaded {
		idx, length = 1, holdSkillFrames
	}
	c.hit("Variable Detection", skillParams[c.TalentLvlSkill()][idx], attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, skillHit)
	c.SetCD(action.ActionSkill, int(skillParams[c.TalentLvlSkill()][6]*60))
	if c.Base.Cons >= 1 {
		c.addBeacon()
	}
	c.refreshShred()
	if !c.overloaded {
		c.QueueCharTask(c.steadyTick(src), steadyInterval)
	}
	c.QueueCharTask(func() {
		if c.reactorSrc == src {
			c.finishCore()
		}
	}, c.reactorUntil-c.Core.F)
	return actionInfo(length, action.SkillState), nil
}
func (c *char) steadyTick(src int) func() {
	return func() {
		if src != c.reactorSrc || !c.reactorActive() || c.overloaded {
			return
		}
		mult := skillParams[c.TalentLvlSkill()][2]
		if c.Base.Cons >= 6 {
			mult *= 1.15
		}
		c.hit("Steady Core", mult, attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, 0)
		c.QueueCharTask(c.steadyTick(src), steadyInterval)
	}
}
func (c *char) Burst(map[string]int) (action.Info, error) {
	c.SetCD(action.ActionBurst, int(burstParams[c.TalentLvlBurst()][3]*60))
	c.ConsumeEnergy(0)
	c.QueueCharTask(func() {
		p := burstParams[c.TalentLvlBurst()]
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Theory of Gradation", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Electro, Durability: 25, Mult: p[0], FlatDmg: p[1] * c.Stat(attributes.EM), IgnoreInfusion: true}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5), 0, 0)
		for i := 0; i < int(p[2]); i++ {
			c.addBeacon()
		}
	}, burstHit)
	return actionInfo(burstFrames, action.BurstState), nil
}
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if !c.reactorActive() || !c.overloaded || c.beaconCount() == 0 {
		c.hit("Charged Attack", normalParams[c.TalentLvlAttack()][3], attacks.AttackTagExtra, attacks.ICDTagExtraAttack, chargeHit)
		return actionInfo(chargeFrames, action.ChargeAttackState), nil
	}
	duration := chargeHit + max(1, c.beaconCount())*beaconInterval
	if d, ok := p["duration"]; ok {
		duration = max(chargeFrames, d)
	}
	c.channelSrc++
	src := c.channelSrc
	c.hit("Gradation Detonation", normalParams[c.TalentLvlAttack()][8], attacks.AttackTagExtra, attacks.ICDTagExtraAttack, chargeHit)
	for d := chargeHit + beaconInterval; d <= duration; d += beaconInterval {
		c.QueueCharTask(func() {
			if src != c.channelSrc || c.Core.Player.Active() != c.Index() || !c.reactorActive() || !c.overloaded {
				return
			}
			if c.spendBeacon() {
				c.starHit("Gradation Beacon", normalParams[c.TalentLvlAttack()][9], true, 0, false)
			}
		}, d)
	}
	stop := func() {
		if src == c.channelSrc {
			c.channelSrc++
		}
	}
	c.QueueCharTask(stop, duration+1)
	a := actionInfo(duration, action.ChargeAttackState)
	a.OnRemoved = func(action.AnimationState) { stop() }
	return a, nil
}
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error)  { return c.plunge(p, false) }
func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) { return c.plunge(p, true) }
func (c *char) plunge(p map[string]int, high bool) (action.Info, error) {
	c.Core.Player.SetAirborne(player.Grounded)
	v := normalParams[c.TalentLvlAttack()]
	if p["collision"] != 0 {
		c.hit("Plunge Collision", v[5], attacks.AttackTagPlunge, attacks.ICDTagNone, 30)
	}
	idx := 6
	if high {
		idx = 7
	}
	c.hit("Plunge", v[idx], attacks.AttackTagPlunge, attacks.ICDTagNone, 40)
	return actionInfo(65, action.PlungeAttackState), nil
}
