package vodyanitsa

import (
	"math"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension >= 1 {
		c.Core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
			if !c.StatusIsActive(microphoneKey) {
				return
			}
			atk, ok := args[1].(*info.AttackEvent)
			if !ok {
				return
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
			default:
				return
			}
			c.AddStatus("vodyanitsa-flowing-vortex", 6*60, true)
			target, ok := args[0].(*enemy.Enemy)
			if !ok {
				return
			}
			target.AddResistMod(info.ResistMod{
				Base:  modifier.NewBaseWithHitlag("vodyanitsa-flowing-vortex-anemo-res", 6*60),
				Ele:   attributes.Anemo,
				Value: -0.30,
			})
		}, "vodyanitsa-a1-flowing-vortex")
	}
	if c.Base.Ascension >= 4 {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			if !c.StatusIsActive(microphoneKey) {
				return
			}
			atk, ok := args[1].(*info.AttackEvent)
			if !ok {
				return
			}
			c.applyA2Bonus(atk)
		}, "vodyanitsa-a2-damage")
	}
	if c.Base.Cons >= 6 {
		for _, ch := range c.Core.Player.Chars() {
			target := ch
			target.AddAttackMod(character.AttackMod{
				Base: modifier.NewBase("vodyanitsa-c6", -1),
				Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
					if !c.StatusIsActive(microphoneKey) {
						return nil
					}
					// Star Diffusion damage is a special reaction and does not
					// consume elemental DMG% stats. It is handled below via Elevation.
					if atk.Info.AttackTag == attacks.AttackTagReactionStarDiffusionAnemo || atk.Info.AttackTag == attacks.AttackTagReactionStarDiffusionCryo {
						return nil
					}
					out := make([]float64, attributes.EndStatType)
					if atk.Info.Element == attributes.Hydro || atk.Info.Element == attributes.Cryo {
						out[attributes.DmgP] += .50
					}
					return out
				},
			})
		}
		c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
			atk, ok := args[0].(*info.AttackEvent)
			if !ok || !c.StatusIsActive(microphoneKey) {
				return
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				atk.Info.Elevation += .25
			}
		}, "vodyanitsa-c6-star-elevation")
	}
}

func (c *char) applyA2Bonus(atk *info.AttackEvent) {
	star := attacks.AttackTagIsStar(atk.Info.AttackTag)
	active := atk.Info.ActorIndex == c.Core.Player.Active()
	if active {
		if c.soloStacks <= 0 {
			return
		}
	} else if c.concertStacks <= 0 {
		return
	}

	// Each enemy hit consumes one stack. The bonus is mode-dependent, so a
	// non-Hydro/Cryo hit (or a star hit without the vortex) still consumes a
	// stack but receives no flat-damage addition.
	if active {
		c.soloStacks--
	} else {
		c.concertStacks--
	}
	if star != c.StatusIsActive("vodyanitsa-flowing-vortex") {
		return
	}
	if !star && (atk.Info.AttackTag >= attacks.AttackTagNoneStat || (atk.Info.Element != attributes.Hydro && atk.Info.Element != attributes.Cryo)) {
		return
	}

	bonus := math.Floor(max(c.MaxHP()-40000, 0) / 1000)
	if star {
		bonus = min(bonus*260, 6500.0)
	} else {
		bonus = min(bonus*140, 3500.0)
	}
	atk.Info.FlatDmg += bonus
}
