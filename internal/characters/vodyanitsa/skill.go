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

const (
	// SongKey identifies the active 遥久之歌 state for party interactions.
	SongKey         = "vodyanitsa-microphone-summon"
	microphoneKey   = SongKey
	songStacksKey   = "vodyanitsa-song-stacks"
	c2Key           = "vodyanitsa-c2"
	recentVortexKey = "vodyanitsa-recent-flowing-vortex"
)

// Timings are from user image 3; see PLACEHOLDER_FRAMES.md. The skill's
// nominal duration and cooldown continue to come from origin_data.
func (c *char) Skill(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Water Nymph Overture", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Hydro, Durability: 25, UseHP: true, Mult: skill[c.TalentLvlSkill()]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), skillHitmark, skillHitmark, c.microphoneResShred(lvl))

	c.skillSrc = c.Core.F
	src := c.skillSrc
	// A2 grants these two linked resources whenever the microphone state starts.
	if c.Base.Ascension >= 4 {
		c.soloStacks = 25
		c.concertStacks = 10
		c.AddStatus(songStacksKey, 30*60, true)
	} else {
		c.soloStacks = 0
		c.concertStacks = 0
	}
	dur := 16 * 60
	if c.Base.Cons >= 2 {
		dur += 9 * 60
	}
	c.AddStatus(microphoneKey, dur, true)
	if c.Base.Ascension >= 1 && c.Core.StarReactions.DiffusionStacks > 0 {
		c.flowingVortex = true
		c.shredAnemo()
	}
	for _, delay := range microphoneAttackFrames {
		if delay < dur {
			c.QueueCharTask(c.microphoneAttack(src), delay)
		}
	}
	for _, delay := range microphoneHealFrames {
		if delay < dur {
			c.QueueCharTask(c.microphoneHeal(src, lvl), delay)
		}
	}
	c.SetCD(action.ActionSkill, 16*60)
	f := frames.InitAbilSlice(skillLength)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: skillLength, CanQueueAfter: skillLength, State: action.SkillState}, nil
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
	c.c2Star = c.flowingVortex || c.StatusDuration(recentVortexKey) > 0
	c.AddStatus(c2Key, 5*60, true)
	for _, target := range c.Core.Player.Chars() {
		bonus := make([]float64, attributes.EndStatType)
		target.AddAttackMod(character.AttackMod{
			Base: modifier.NewBaseWithHitlag(c2Key, 5*60),
			Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
				if c.StatusDuration(c2Key) == 0 || (c.Base.Cons < 6 && target.Index() != c.Core.Player.Active()) {
					return nil
				}
				// Star reaction contributions are modified through OnStarReactionAttack;
				// AttackMods are not applied while those contributions are calculated.
				if c.c2Star && atk.Info.IsDirectStarDamage() && atk.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
					bonus[attributes.CD] = .60
					return bonus
				}
				if !c.c2Star && !attacks.AttackTagIsStar(atk.Info.AttackTag) && (atk.Info.Element == attributes.Hydro || atk.Info.Element == attributes.Cryo) {
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
		c.c1Buff()
	}
}
