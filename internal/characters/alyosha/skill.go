package alyosha

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

const hunterMarkKey = "alyosha-hunter-mark"

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"] != 0
	index, hitmark := 0, 24
	if hold {
		index, hitmark = 1, 42
	}
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{
		ActorIndex: c.Index(), Abil: "Thunderbolt Strike", AttackTag: attacks.AttackTagElementalArt,
		ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypePierce, Element: attributes.Electro, Durability: 25,
		Mult: skill[index][lvl],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), hitmark, hitmark, c.skillHit)
	c.SetCD(action.ActionSkill, int(skillParam[2][lvl]*60))
	f := frames.InitAbilSlice(hitmark + 24)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: hitmark + 24, CanQueueAfter: hitmark + 8, State: action.SkillState}, nil
}

func (c *char) skillHit(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	c.markSrc = c.Core.F
	c.AddStatus(hunterMarkKey, int(skillParam[3][c.TalentLvlSkill()]*60), true)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Electro, c.ParticleDelay)
}

func (c *char) activateMark() {
	if !c.StatusIsActive(hunterMarkKey) {
		return
	}
	c.DeleteStatus(hunterMarkKey)
	dur := int(skillParam[5][c.TalentLvlSkill()] * 60)
	amount := skillParam[4][c.TalentLvlSkill()]
	if c.Base.Cons >= 6 {
		amount *= 2
	}
	for _, ch := range c.Core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.ATKP] = amount
		if c.Base.Cons >= 6 {
			buff[attributes.EM] = 100
		}
		ch.AddStatMod(character.StatMod{
			Base:   modifier.NewBaseWithHitlag("alyosha-hunters-precision", dur),
			Amount: func() []float64 { return buff },
		})
	}
}
