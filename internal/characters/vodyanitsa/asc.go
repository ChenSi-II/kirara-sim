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
		c.Core.Events.Subscribe(event.OnStarDiffusionVortex, func(args ...any) {
			detonated := args[1].(bool)
			if detonated {
				if c.flowingVortex {
					c.shredAnemo()
					c.AddStatus(recentVortexKey, 5*60, true)
				}
				c.flowingVortex = false
				return
			}
			c.flowingVortex = c.StatusIsActive(microphoneKey)
			if c.flowingVortex {
				c.shredAnemo()
			}
		}, "vodyanitsa-a1-flowing-vortex")
	}
	if c.Base.Ascension >= 4 {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			if c.StatusDuration(songStacksKey) == 0 {
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
						out[attributes.DmgP] += .60
					}
					return out
				},
			})
		}
		c.AddStarDamageMod("vodyanitsa-c6-star-elevation", func(atk *info.AttackEvent) {
			if !c.StatusIsActive(microphoneKey) {
				return
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				atk.Info.Elevation += .25
			}
		})
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
	if star != (c.flowingVortex || c.StatusDuration(recentVortexKey) > 0) {
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

// The existing simulator treats the party and its targets as nearby.
func (c *char) shredAnemo() {
	for _, t := range c.Core.Combat.Enemies() {
		target, ok := t.(*enemy.Enemy)
		if !ok {
			continue
		}
		target.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag("vodyanitsa-flowing-vortex-anemo-res", 6*60),
			Ele:   attributes.Anemo,
			Value: -0.35,
		})
	}
}
