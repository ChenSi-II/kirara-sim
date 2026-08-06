package illuga

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func (c *char) Skill(p map[string]int) (action.Info, error) {
	index, multIndex, emIndex, hitmark := 0, 1, 0, 24
	if p["hold"] != 0 {
		index, multIndex, emIndex, hitmark = 1, 3, 2, 42
	}
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Dawnbearing Songbird", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Geo, Durability: 25, UseDef: true, Mult: skillParam[multIndex][lvl], FlatDmg: skillParam[emIndex][lvl] * c.Stat(attributes.EM)}
	_ = index
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), hitmark, hitmark, c.particleCB)
	c.lightkeepersOath()
	c.SetCD(action.ActionSkill, int(skillParam[4][lvl]*60))
	f := frames.InitAbilSlice(hitmark + 24)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: hitmark + 24, CanQueueAfter: hitmark + 8, State: action.SkillState}, nil
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() == info.TargettableEnemy && !c.StatusIsActive("illuga-particle-icd") {
		c.AddStatus("illuga-particle-icd", 5*60, true)
		c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Geo, c.ParticleDelay)
	}
}
