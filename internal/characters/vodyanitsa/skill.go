package vodyanitsa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const microphoneKey = "vodyanitsa-microphone-summon"

func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Water Nymph Overture", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Hydro, Durability: 25, UseHP: true, Mult: skill[c.TalentLvlSkill()]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 28, 28)

	c.skillSrc = c.Core.F
	src := c.skillSrc
	c.AddStatus(microphoneKey, 16*60, true)
	for delay := 3 * 60; delay < 16*60; delay += 3 * 60 {
		c.QueueCharTask(c.microphoneAttack(src), delay)
	}
	for delay := 90; delay < 16*60; delay += 90 {
		c.QueueCharTask(c.microphoneHeal(src, lvl), delay)
	}
	c.SetCD(action.ActionSkill, 16*60)
	f := frames.InitAbilSlice(58)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 58, CanQueueAfter: 42, State: action.SkillState}, nil
}

func (c *char) microphoneAttack(src int) func() {
	return func() {
		if src != c.skillSrc || !c.StatusIsActive(microphoneKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Microphone Performance", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Hydro, Durability: 25, UseHP: true, Mult: skill[c.TalentLvlSkill()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0, c.microphoneResShred(c.TalentLvlSkill()))
	}
}

func (c *char) microphoneResShred(lvl int) info.AttackCBFunc {
	return func(a info.AttackCB) {
		target, ok := a.Target.(*enemy.Enemy)
		if !ok {
			return
		}
		for _, ele := range []attributes.Element{attributes.Hydro, attributes.Cryo} {
			target.AddResistMod(info.ResistMod{
				Base:  modifier.NewBaseWithHitlag("vodyanitsa-microphone-"+ele.String()+"-res", 6*60),
				Ele:   ele,
				Value: -skillResist[lvl],
			})
		}
	}
}

func (c *char) microphoneHeal(src, lvl int) func() {
	return func() {
		if src != c.skillSrc || !c.StatusIsActive(microphoneKey) {
			return
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  c.Core.Player.Active(),
			Message: "Microphone Performance",
			Src:     skillHealFlat[lvl] + skillHealPct[lvl]*c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}
