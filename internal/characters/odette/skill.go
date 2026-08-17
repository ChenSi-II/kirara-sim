package odette

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	doubleKey = "odette-solo-dance-double"
	codaKey   = "odette-coda-ready"
)

func (c *char) Skill(map[string]int) (action.Info, error) {
	if c.StatusIsActive(codaKey) && c.StatusIsActive(doubleKey) {
		return c.coda()
	}
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Phantom Night Dancers", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skillParam[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 30, 30, c.skillParticle)
	c.summonDouble(int(skillParam[10][lvl] * 60))
	c.AddStatus(codaKey, 6*60, true)
	c.SetCD(action.ActionSkill, int(skillParam[11][lvl]*60))
	f := frames.InitAbilSlice(62)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 62, CanQueueAfter: 46, State: action.SkillState}, nil
}

func (c *char) coda() (action.Info, error) {
	c.DeleteStatus(codaKey)
	lvl := c.TalentLvlSkill()
	for i := 0; i < 3; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Coda at Dawn's Tolling", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: skillParam[1][lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4), 20+i*12, 20+i*12)
	}
	final := c.stellarAttack("Coda Finale", skillParam[2][lvl], skillParam[3][lvl], attacks.AttackTagElementalArt)
	c.Core.QueueAttack(final, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 62, 62)
	if c.Base.Cons >= 1 {
		extra := c.stellarAttack("Coda Finale (C1)", 3, 4.5, attacks.AttackTagReactionStarSuperconduct)
		c.Core.QueueAttack(extra, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 68, 68)
	}
	f := frames.InitAbilSlice(82)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 82, CanQueueAfter: 70, State: action.SkillState}, nil
}

func (c *char) summonDouble(dur int) {
	c.doubleSrc = c.Core.F
	src := c.doubleSrc
	c.AddStatus(doubleKey, dur, true)
	c.grantSplendor(src)
	for delay := 90; delay <= dur; delay += 90 {
		c.QueueCharTask(c.doubleTick(src, delay/90), delay)
	}
}

func (c *char) doubleTick(src, tick int) func() {
	return func() {
		if src != c.doubleSrc || !c.StatusIsActive(doubleKey) {
			return
		}
		lvl := c.TalentLvlSkill()
		plume := tick%2 == 1
		name, normal, conduct, swirl := "Wing", skillParam[7][lvl], skillParam[8][lvl], skillParam[9][lvl]
		if plume {
			name, normal, conduct, swirl = "Plume", skillParam[4][lvl], skillParam[5][lvl], skillParam[6][lvl]
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Dance Double " + name, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: normal}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		if c.Core.StarReactions.SuperconductActive || c.Core.StarReactions.DiffusionActive {
			star := c.stellarAttack("Dance Double "+name+" Stellar", conduct, swirl, attacks.AttackTagElementalArt)
			c.Core.QueueAttack(star, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 4, 4)
		}
	}
}

func (c *char) stellarAttack(name string, conduct, swirl float64, fallback attacks.AttackTag) info.AttackInfo {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: name, AttackTag: fallback, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Mult: conduct}
	if c.Core.StarReactions.SuperconductActive {
		ai.AttackTag = attacks.AttackTagReactionStarSuperconduct
	}
	if c.Core.StarReactions.DiffusionActive {
		ai.AttackTag, ai.Mult = attacks.AttackTagReactionStarDiffusionCryo, swirl
	}
	return ai
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("odette-particle-icd") {
		c.AddStatus("odette-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Cryo, c.ParticleDelay)
	}
}
