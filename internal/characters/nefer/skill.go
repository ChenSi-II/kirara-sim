package nefer

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	shadowDanceKey = "nefer-shadow-dance"
	seedWindowKey  = "nefer-seed-of-deceit-window"
)

func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Dance of a Thousand Nights", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, Durability: 25, Mult: skillParam[0][lvl], FlatDmg: skillParam[1][lvl] * c.Stat(attributes.EM)}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 26, 26, c.skillParticle)
	c.phantasmUses = 0
	c.AddStatus(shadowDanceKey, int(skillParam[3][lvl]*60), true)
	if c.Base.Ascension >= 1 {
		c.AddStatus(seedWindowKey, 15*60, true)
	}
	if c.Base.Cons >= 2 {
		c.addVeils(2)
	}
	if c.Base.Cons >= 4 {
		c.QueueCharTask(c.c4ResTick(c.Core.F), 1)
	}
	c.SetCD(action.ActionSkill, int(skillParam[11][lvl]*60))
	f := frames.InitAbilSlice(52)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 52, CanQueueAfter: 36, State: action.SkillState}, nil
}

func (c *char) phantasmPerformance() (action.Info, error) {
	lvl := c.TalentLvlSkill()
	bonus := 1 + .08*float64(c.veils)
	c.phantasmUses++
	c.Core.Player.ConsumeDew(1)
	c.absorbSeeds()
	for i := 0; i < 2; i++ {
		atkIdx, emIdx := 4+i*2, 5+i*2
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Phantasm Performance Nefer %d", i+1), AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, Durability: 25, Mult: skillParam[atkIdx][lvl] * bonus, FlatDmg: skillParam[emIdx][lvl] * c.Stat(attributes.EM) * bonus}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 18+i*12, 18+i*12)
	}
	for i := 0; i < 3; i++ {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: fmt.Sprintf("Phantasm Performance Shade %d", i+1), AttackTag: attacks.AttackTagDirectLunarBloom, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, IgnoreDefPercent: 1, Mult: skillParam[8+i][lvl] * bonus}
		if c.Base.Cons >= 6 && i == 1 {
			ai.Mult = 0
			ai.FlatDmg = .85 * c.Stat(attributes.EM)
			ai.Elevation = .15
		}
		if c.Base.Cons >= 1 {
			ai.FlatDmg += .6 * c.Stat(attributes.EM) * bonus
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 24+i*10, 24+i*10)
	}
	if c.Base.Cons >= 6 {
		finish := info.AttackInfo{ActorIndex: c.Index(), Abil: "Phantasm Performance (C6) Finale", AttackTag: attacks.AttackTagDirectLunarBloom, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro, IgnoreDefPercent: 1, FlatDmg: 1.2 * c.Stat(attributes.EM), Elevation: .15}
		c.Core.QueueAttack(finish, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 60, 60)
	}
	if c.phantasmUses >= 3 {
		c.DeleteStatus(shadowDanceKey)
	}
	f := frames.InitAbilSlice(76)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 76, CanQueueAfter: 62, State: action.ChargeAttackState}, nil
}

func (c *char) absorbSeeds() {
	if c.seeds == 0 {
		return
	}
	c.addVeils(c.seeds)
	c.seeds = 0
}

func (c *char) addVeils(n int) {
	limit := 3
	if c.Base.Cons >= 2 {
		limit = 5
	}
	c.veils = min(limit, c.veils+n)
	if c.veils == limit {
		amount := 100.
		if c.Base.Cons >= 2 {
			amount = 200
		}
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.EM] = amount
		c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("nefer-veil-em", 8*60), AffectedStat: attributes.EM, Amount: func() []float64 { return buff }})
	}
}

func (c *char) c4ResTick(src int) func() {
	return func() {
		if c.Base.Cons < 4 || !c.StatusIsActive(shadowDanceKey) {
			return
		}
		for _, target := range c.Core.Combat.Enemies() {
			if e, ok := target.(*enemy.Enemy); ok {
				e.AddResistMod(info.ResistMod{Base: modifier.NewBase("nefer-c4-dendro-res", 90), Ele: attributes.Dendro, Value: -.20})
			}
		}
		c.QueueCharTask(c.c4ResTick(src), 60)
	}
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("nefer-particle-icd") {
		c.AddStatus("nefer-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Dendro, c.ParticleDelay)
	}
}
