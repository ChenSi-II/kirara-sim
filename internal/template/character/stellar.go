package character

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const DiffusionRadianceKey = "stellar-diffusion-radiance"

// Enhanced is opt-out, like the existing hexerei parameter. It controls the
// additional kit; the two Radiance states still require their actual triggers.
func (c *Character) SetEnhanced(p info.CharacterProfile) {
	value, exists := p.Params["enhanced"]
	c.Enhanced = !exists || value != 0
}

func (c *Character) InitDiffusionRadiance(superconductPriority bool) {
	if !c.Enhanced {
		return
	}
	c.superconductPriority = superconductPriority
	c.Core.Events.Subscribe(event.OnStarDiffusion, func(...any) {
		c.AddStatus(DiffusionRadianceKey, 8*60, true)
	}, c.Base.Key.String()+"-diffusion-radiance")
}

func (c *Character) SuperconductRadiance() bool {
	return c.Enhanced && c.Core.StarReactions.SuperconductActive
}
func (c *Character) DiffusionRadiance() bool {
	return c.Enhanced && c.StatusIsActive(DiffusionRadianceKey) && !(c.superconductPriority && c.SuperconductRadiance())
}
func (c *Character) StellarRadiance() bool { return c.SuperconductRadiance() || c.DiffusionRadiance() }

func IsStarDiffusion(tag attacks.AttackTag) bool {
	return tag == attacks.AttackTagReactionStarDiffusionAnemo || tag == attacks.AttackTagReactionStarDiffusionCryo
}

// AddStarDamageMod applies a modifier before calculation to both talent hits
// and individual Star Diffusion contributions. The final diffusion packet
// already contains those contributions and must not receive the modifier twice.
func (c *Character) AddStarDamageMod(key string, apply func(*info.AttackEvent)) {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if attacks.AttackTagIsStar(atk.Info.AttackTag) && !atk.Info.IsStarDiffusionReaction {
			apply(atk)
		}
	}, key)
	c.Core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		apply(args[1].(*info.AttackEvent))
	}, key)
}

func (c *Character) RadiantReaction(tag attacks.AttackTag) bool {
	if c.SuperconductRadiance() && (tag == attacks.AttackTagSuperconductDamage || tag == attacks.AttackTagReactionStarSuperconduct) {
		return true
	}
	return c.DiffusionRadiance() && (tag == attacks.AttackTagSwirlCryo || IsStarDiffusion(tag))
}
