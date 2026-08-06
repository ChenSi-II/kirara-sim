package kachina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const twirlyKey = "kachina-turbo-twirly"

func (c *char) Skill(p map[string]int) (action.Info, error) {
	c.summonTwirly(p["mount"] != 0)
	c.SetCD(action.ActionSkill, int(skillParam[3][c.TalentLvlSkill()]*60))
	f := frames.InitAbilSlice(42)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 42, CanQueueAfter: 28, State: action.SkillState}, nil
}

func (c *char) summonTwirly(mounted bool) {
	c.twirlySrc = c.Core.F
	src := c.twirlySrc
	c.AddStatus(twirlyKey, 12*60, true)
	index := 1
	if mounted {
		index = 0
	}
	for delay := 30; delay <= 12*60; delay += 2 * 60 {
		c.QueueCharTask(c.twirlyTick(src, index), delay)
	}
}

func (c *char) twirlyTick(src, index int) func() {
	return func() {
		if src != c.twirlySrc || !c.StatusIsActive(twirlyKey) {
			return
		}
		lvl := c.TalentLvlSkill()
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Turbo Twirly", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Geo, Durability: 50, UseDef: true, Mult: skill[index][lvl]}
		if c.Base.Ascension >= 4 {
			ai.FlatDmg = .2 * c.TotalDef(false)
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0, c.twirlyParticle)
	}
}

func (c *char) twirlyParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("kachina-particle-icd") {
		c.AddStatus("kachina-particle-icd", 4*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
	}
}
