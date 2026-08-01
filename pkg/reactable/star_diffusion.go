package reactable

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	starDiffusionMaxStacks = 6
	starDiffusionInterval  = 270 // 4.5s
)

func (r *Reactable) tryStarDiffusion(a *info.AttackEvent, aura attributes.Element) bool {
	mod := info.ReactionModKeyCryo
	if aura == attributes.Frozen {
		mod = info.ReactionModKeyFrozen
	}
	if a.Info.Durability < info.ZeroDur || r.GetAuraDurability(mod) < info.ZeroDur {
		return false
	}

	rd := r.reduce(aura, a.Info.Durability, 0.5)
	a.Info.Durability = max(a.Info.Durability-rd, 0)
	a.Reacted = true
	if aura == attributes.Frozen {
		r.checkFreeze()
	}
	r.core.Events.Emit(event.OnStarDiffusion, r.self, a)

	state := &r.core.StarReactions
	newDomain := !state.DiffusionActive
	state.DiffusionActive = true
	state.DiffusionStacks = min(state.DiffusionStacks+1, starDiffusionMaxStacks)
	state.DiffusionOwner = a.Info.ActorIndex
	state.DiffusionTarget = r.self

	// Preserve ordinary Cryo Swirl's reaction-damage GCD while still counting
	// every successfully triggered Star Diffusion toward the vortex.
	if r.swirlCryoGCD == -1 || r.core.F >= r.swirlCryoGCD {
		r.swirlCryoGCD = r.core.F + 0.1*60
		doStarReactionAttack(
			r.core,
			r.self,
			a.Info.ActorIndex,
			info.ReactionTypeStarDiffusionAnemo,
			attacks.AttackTagReactionStarDiffusionAnemo,
			attributes.Anemo,
			0.75,
			combat.NewSingleTargetHit(r.self.Key()),
		)
	}

	if newDomain {
		r.startStarDiffusionCycle()
	}
	if state.DiffusionStacks >= starDiffusionMaxStacks {
		r.detonateStarDiffusion()
		r.startStarDiffusionCycle()
	}
	return true
}

func (r *Reactable) startStarDiffusionCycle() {
	state := &r.core.StarReactions
	state.DiffusionCycle++
	src := state.DiffusionCycle
	r.core.Tasks.Add(func() {
		if r.core.StarReactions.DiffusionCycle != src {
			return
		}
		r.detonateStarDiffusion()
		r.startStarDiffusionCycle()
	}, starDiffusionInterval)
}

func (r *Reactable) detonateStarDiffusion() {
	state := &r.core.StarReactions
	stacks := state.DiffusionStacks
	state.DiffusionStacks = 0
	if stacks == 0 || state.DiffusionTarget == nil {
		return
	}

	mult := 2.0
	if stacks >= 3 {
		mult = 3
	}
	target := state.DiffusionTarget
	doStarReactionAttack(
		r.core,
		target,
		state.DiffusionOwner,
		info.ReactionTypeStarDiffusionCryo,
		attacks.AttackTagReactionStarDiffusionCryo,
		attributes.Cryo,
		mult,
		combat.NewSingleTargetHit(target.Key()),
	)
}
