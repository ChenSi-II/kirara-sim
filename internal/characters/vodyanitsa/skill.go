package vodyanitsa

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

const microphoneKey = "vodyanitsa-microphone-summon"

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Water Nymph Overture", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Hydro, Durability: 25, UseHP: true, Mult: skill[c.TalentLvlSkill()]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 28, 28)

	c.skillSrc = c.Core.F
	src := c.skillSrc
	// A2 grants these two linked resources whenever the microphone state starts.
	if c.Base.Ascension >= 4 {
		c.soloStacks = 17
		c.concertStacks = 10
	} else {
		c.soloStacks = 0
		c.concertStacks = 0
	}
	dur := 16 * 60
	if c.Base.Cons >= 2 {
		dur += 9 * 60
	}
	c.AddStatus(microphoneKey, dur, true)
	for delay := 3 * 60; delay < dur; delay += 3 * 60 {
		c.QueueCharTask(c.microphoneAttack(src), delay)
	}
	for delay := 90; delay < dur; delay += 90 {
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
		c.microphoneC2Buff()
	}
}

func (c *char) microphoneC2Buff() {
	if c.Base.Cons < 2 {
		return
	}
	star := c.StatusIsActive("vodyanitsa-flowing-vortex")
	if star {
		target := c.Core.Player.Active()
		if c.Base.Cons >= 6 {
			target = -1
		}
		expiry := c.Core.F + 5*60
		refreshed := false
		for i := range c.c2StarBuffs {
			if c.c2StarBuffs[i].target == target {
				c.c2StarBuffs[i].expiry = expiry
				refreshed = true
				break
			}
		}
		if !refreshed {
			c.c2StarBuffs = append(c.c2StarBuffs, c2StarBuff{expiry: expiry, target: target})
		}
	}
	for _, target := range c.Core.Player.Chars() {
		if c.Base.Cons < 6 && target.Index() != c.Core.Player.Active() {
			continue
		}
		bonus := make([]float64, attributes.EndStatType)
		target.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag("vodyanitsa-c2", 5*60),
			Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
				// Star reaction contributions are modified through OnStarReactionAttack;
				// AttackMods are not applied while those contributions are calculated.
				if !star && (atk.Info.Element == attributes.Hydro || atk.Info.Element == attributes.Cryo) {
					bonus[attributes.CD] = .50
					return bonus
				}
				bonus[attributes.CD] = 0
				return bonus
			},
		})
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
		active := c.Core.Player.ActiveChar()
		healScale := 1.0
		if c.Base.Cons >= 4 && active.CurrentHPRatio() < .40 {
			healScale = 1.5
		} else if c.Base.Cons >= 4 && c.c4HPStacks < 3 {
			c.c4HPStacks++
			m := make([]float64, attributes.EndStatType)
			m[attributes.HPP] = .20
			c.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag(fmt.Sprintf("vodyanitsa-c4-%d", c.Core.F), 6*60), AffectedStat: attributes.HPP, Amount: func() []float64 { return m }})
			c.QueueCharTask(func() { c.c4HPStacks = max(0, c.c4HPStacks-1) }, 6*60)
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  c.Core.Player.Active(),
			Message: "Microphone Performance",
			Src:     healScale * (skillHealFlat[lvl] + skillHealPct[lvl]*c.MaxHP()),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}
