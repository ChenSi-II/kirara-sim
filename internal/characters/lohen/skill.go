package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const masterstrokeKey = "lohen-masterstroke"

func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.StatusIsActive(masterstrokeKey) && c.joy >= 100 {
		return c.etchedIntoBoneAndSoul()
	}
	lvl := c.TalentLvlSkill()
	c.joy, c.will, c.etchedUses = 0, 0, 0
	if !c.StatusIsActive("lohen-c3-cooldown") {
		momentumDur := 9 * 60
		for _, ch := range c.Core.Player.Chars() {
			if ch.Index() == c.Index() {
				continue
			}
			if ch.TalentLvlAttack() >= lvl || ch.TalentLvlSkill() >= lvl || ch.TalentLvlBurst() >= lvl {
				momentumDur += 6 * 60
				break
			}
		}
		c.AddStatus(c3MomentumKey, momentumDur, true)
		c.AddStatus("lohen-c3-cooldown", 18*60, true)
	}
	dur := int(skillParam[11][lvl] * 60)
	c.AddStatus(masterstrokeKey, dur, true)
	if c.Base.Cons >= 4 {
		if c.Energy < c.EnergyMax {
			c.AddEnergy("lohen-c4", 15)
		} else {
			c.AddStatus("lohen-c4-refund", 15*60, true)
		}
	}
	c.SetCD(action.ActionSkill, int(skillParam[18][lvl]*60))
	f := frames.InitAbilSlice(46)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 46, CanQueueAfter: 30, State: action.SkillState}, nil
}

func (c *char) etchedIntoBoneAndSoul() (action.Info, error) {
	lvl := c.skillLevel()
	bonus := skillParam[17][lvl] * float64(c.will)
	for i := 0; i < 4; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Etched Into Bone and Soul %d", i+1), AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Cryo, Durability: 25, Mult: skillParam[16][lvl], BaseDmgBonus: bonus}
		if c.Base.Cons >= 6 {
			ai.FlatDmg += 0
		}
		cb := c.etchedHit
		if i > 0 {
			cb = nil
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 18+i*7, 18+i*7, cb)
	}
	c.joy = 0
	c.etchedUses++
	if c.Base.Cons >= 6 {
		c.joy = 100
	} else {
		c.will = 0
	}
	if c.Base.Cons >= 2 {
		c.evilsbane = true
	}
	f := frames.InitAbilSlice(58)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 58, CanQueueAfter: 44, State: action.SkillState}, nil
}

func (c *char) attackElement() attributes.Element {
	if c.StatusIsActive(masterstrokeKey) {
		return attributes.Cryo
	}
	return attributes.Physical
}

func (c *char) maxEtchedUses() int {
	if c.Base.Cons >= 6 {
		return 5
	}
	return 3
}

func (c *char) masterstrokeHit(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || !c.StatusIsActive(masterstrokeKey) {
		return
	}
	c.joy = min(100, c.joy+17)
	if c.evilsbane && c.Base.Cons >= 2 {
		c.evilsbane = false
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Evilsbane Blade", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Cryo, Mult: 5}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		for _, ch := range c.Core.Player.Chars() {
			if ch.Index() == c.Index() {
				continue
			}
			buff := make([]float64, attributes.EndStatType)
			buff[attributes.EM] = 200
			ch.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("lohen-c2-em", 8*60), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
		}
	}
}

func (c *char) etchedHit(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("lohen-particle-icd") {
		c.AddStatus("lohen-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
	}
	if c.Base.Cons >= 6 && a.Target.Type() == info.TargettableEnemy {
		if c.c6ExtendReady && !c.StatusIsActive("lohen-c6-extend-icd") {
			c.AddStatus("lohen-c6-extend-icd", 7*60, true)
			if c.StatusIsActive(masterstrokeKey) {
				c.AddStatus(masterstrokeKey, c.StatusDuration(masterstrokeKey)+75, true)
			}
		}
		c.c6ExtendReady = true
	}
}
