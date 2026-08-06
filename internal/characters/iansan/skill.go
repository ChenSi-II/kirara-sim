package iansan

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const nightsoulKey = "iansan-nightsoul"

func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Thunderbolt Rush", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Electro, Durability: 25, Mult: skill[0][lvl]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2.5), 22, 22, c.skillParticle)
	c.restoreNightsoul(int(skillParam[1][lvl]))
	c.AddStatus(nightsoulKey, 20*60, true)
	c.a1Buff()
	c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	f := frames.InitAbilSlice(48)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 48, CanQueueAfter: 30, State: action.SkillState}, nil
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("iansan-particle-icd") {
		c.AddStatus("iansan-particle-icd", 4*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 4, attributes.Electro, c.ParticleDelay)
	}
}

func (c *char) restoreNightsoul(amount int) {
	before := c.nightsoul
	if c.Base.Cons >= 4 && c.surging > 0 {
		amount += 4
		c.surging--
	}
	overflow := max(before+amount-54, 0)
	c.nightsoul = min(54, c.nightsoul+amount)
	if c.Base.Ascension >= 4 && c.nightsoul > before && !c.StatusIsActive("iansan-a4-heal-icd") {
		c.AddStatus("iansan-a4-heal-icd", 168, true)
		c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: c.Core.Player.Active(), Message: "Kinetic Energy Gradient Test", Src: .6 * c.TotalAtk(), Bonus: c.Stat(attributes.Heal)})
	}
	if c.Base.Cons >= 6 && overflow > 0 {
		c.c6Buff()
	}
}
