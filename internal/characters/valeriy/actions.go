package valeriy

import (
	"fmt"
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"math"
)

// UNMEASURED placeholder animation, hitmarks and barrage cadence. No image
// data exists for this character. See PLACEHOLDER_FRAMES.md.
const normalHit = 18
const normalFrames = 42
const skillHit = 24
const skillFrames = 50
const chargeHit = 30
const chargeFrames = 60
const burstHit = 45
const burstFrames = 90
const barrageInterval = 60

func actionInfo(length int, state action.AnimationState) action.Info {
	return action.Info{Frames: frames.NewAbilFunc(frames.InitAbilSlice(length)), AnimationLength: length, CanQueueAfter: length, State: state}
}
func (c *char) hit(name string, mult float64, ele attributes.Element, tag attacks.AttackTag, icd attacks.ICDTag, delay int, enhanced bool) {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: name, AttackTag: tag, ICDTag: icd, ICDGroup: attacks.ICDGroupDefault, Element: ele, Durability: 25, Mult: mult}
	var cb []info.AttackCBFunc
	if enhanced && c.Base.Cons >= 6 {
		cb = append(cb, c.c6Hit)
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), delay, delay, cb...)
}
func (c *char) Attack(map[string]int) (action.Info, error) {
	i := c.NormalCounter
	c.hit(fmt.Sprintf("Normal %d", i+1), normalParams[c.TalentLvlAttack()][i], attributes.Physical, attacks.AttackTagNormal, attacks.ICDTagNormalAttack, normalHit, false)
	c.AdvanceNormalIndex()
	return actionInfo(normalFrames, action.NormalAttackState), nil
}
func (c *char) Skill(map[string]int) (action.Info, error) {
	c.SetCD(action.ActionSkill, int(skillParams[c.TalentLvlSkill()][4]*60))
	n := 40.0
	if c.firstSkill && c.Base.Cons >= 1 {
		n *= 2.5
	}
	c.firstSkill = false
	c.gainPotential(n * c.gainFactor())
	c.createShield()
	c.hit("Bastion Advance", skillParams[c.TalentLvlSkill()][0], attributes.Electro, attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, skillHit, false)
	if c.Base.Cons >= 2 {
		for _, ch := range c.Core.Player.Chars() {
			m := make([]float64, attributes.EndStatType)
			m[attributes.EM] = 100
			ch.AddStatMod(character.StatMod{Base: modifier.NewBase("valeriy-c2-em", 1080), AffectedStat: attributes.EM, Amount: func() []float64 { return m }})
		}
	}
	return actionInfo(skillFrames, action.SkillState), nil
}
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	p := normalParams[c.TalentLvlAttack()]
	if c.potential < 40 {
		c.hit("Charged Attack", p[3], attributes.Physical, attacks.AttackTagExtra, attacks.ICDTagExtraAttack, chargeHit, false)
	} else {
		spent := c.potential
		c.potential = 0
		c.hit("Point-Blank Mortar", p[4], attributes.Electro, attacks.AttackTagExtra, attacks.ICDTagExtraAttack, chargeHit, true)
		// Orders are granted after the triggering special charged attack.
		c.QueueCharTask(func() {
			c.spent = spent
			c.orders = int(p[8])
			c.ordersUntil = c.Core.F + int(p[9]*60)
			if c.Base.Cons >= 1 && !c.StatusIsActive("valeriy-c1-energy") {
				c.AddEnergy("valeriy-c1", min(15, 6+math.Floor((spent-40)/10)*1.5))
				c.AddStatus("valeriy-c1-energy", 900, false)
			}
		}, chargeHit+1)
	}
	return actionInfo(chargeFrames, action.ChargeAttackState), nil
}
func (c *char) Burst(map[string]int) (action.Info, error) {
	p := burstParams[c.TalentLvlBurst()]
	c.SetCD(action.ActionBurst, int(p[3]*60))
	c.ConsumeEnergy(0)
	c.burstSrc++
	src := c.burstSrc
	c.burstGained = 0
	c.burstUntil = c.Core.F + int(p[2]*60)
	c.hit("Thunder Barrage", p[0], attributes.Electro, attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, burstHit, false)
	for d := barrageInterval; d < int(p[2]*60); d += barrageInterval {
		c.QueueCharTask(func() {
			if src == c.burstSrc && c.Core.F < c.burstUntil {
				c.hit("Thunder Shrapnel", p[1], attributes.Electro, attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, 0, true)
			}
		}, d)
	}
	return actionInfo(burstFrames, action.BurstState), nil
}
func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error)  { return c.plunge(p, false) }
func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) { return c.plunge(p, true) }
func (c *char) plunge(p map[string]int, high bool) (action.Info, error) {
	c.Core.Player.SetAirborne(player.Grounded)
	v := normalParams[c.TalentLvlAttack()]
	if p["collision"] != 0 {
		c.hit("Plunge Collision", v[10], attributes.Physical, attacks.AttackTagPlunge, attacks.ICDTagNone, 30, false)
	}
	idx := 11
	if high {
		idx = 12
	}
	c.hit("Plunge", v[idx], attributes.Physical, attacks.AttackTagPlunge, attacks.ICDTagNone, 40, false)
	return actionInfo(65, action.PlungeAttackState), nil
}
