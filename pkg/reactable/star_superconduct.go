package reactable

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	starSuperconductMaxStacks = 12
	starSuperconductInterval  = 4 * 60
	starSuperconductBuffKey   = "star-superconduct-elemental-bonus"
)

func (r *Reactable) addStarSuperconductStack(a *info.AttackEvent) {
	state := &r.core.StarReactions
	if !state.Enabled || !state.SuperconductActive || a.Info.Durability < info.ZeroDur {
		return
	}
	if a.Info.Element != attributes.Cryo && a.Info.Element != attributes.Electro {
		return
	}
	state.SuperconductStacks = min(state.SuperconductStacks+1, starSuperconductMaxStacks)
}

func (r *Reactable) activateStarSuperconductDomain() {
	state := &r.core.StarReactions
	if state.SuperconductActive {
		return
	}
	state.SuperconductActive = true
	state.SuperconductCoefficient = 1
	r.applyStarSuperconductElementalBonus(0)
	r.core.Tasks.Add(r.settleStarSuperconduct, starSuperconductInterval)
}

func (r *Reactable) settleStarSuperconduct() {
	state := &r.core.StarReactions
	stacks := min(state.SuperconductStacks, starSuperconductMaxStacks)
	state.SuperconductStacks = 0
	if stacks == 0 {
		// Zero stacks is the documented exception to 1.4 + 0.05 * stacks.
		state.SuperconductCoefficient = 1
	} else {
		state.SuperconductCoefficient = 1.4 + 0.05*float64(stacks)
	}
	r.applyStarSuperconductElementalBonus(stacks)
	r.core.Log.NewEvent("star superconduct domain settled", glog.LogElementEvent, -1).
		Write("stacks", stacks).
		Write("coefficient", state.SuperconductCoefficient).
		Write("elemental_bonus", 0.28+0.01*float64(stacks))
	r.core.Tasks.Add(r.settleStarSuperconduct, starSuperconductInterval)
}

func (r *Reactable) applyStarSuperconductElementalBonus(stacks int) {
	bonus := 0.28 + 0.01*float64(stacks)
	for _, char := range r.core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		buff[attributes.CryoP] = bonus
		buff[attributes.ElectroP] = bonus
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(starSuperconductBuffKey, starSuperconductInterval),
			AffectedStat: attributes.NoStat,
			Amount: func() []float64 {
				return buff
			},
		})
	}
}

func (r *Reactable) tryStarSuperconduct(a *info.AttackEvent, frozen bool) bool {
	if a.Info.Durability < info.ZeroDur {
		return false
	}

	var reacted bool
	if frozen {
		reacted = r.tryStarSuperconductOnFrozen(a)
	} else {
		reacted = r.tryStarSuperconductOnAura(a)
	}
	if !reacted {
		return false
	}

	a.Reacted = true
	r.core.Events.Emit(event.OnStarSuperconduct, r.self, a)
	r.activateStarSuperconductDomain()
	r.queueStarSuperconduct(a.Info.ActorIndex)
	return true
}

func (r *Reactable) tryStarSuperconductOnFrozen(a *info.AttackEvent) bool {
	if a.Info.Element != attributes.Electro || r.GetAuraDurability(info.ReactionModKeyFrozen) < info.ZeroDur {
		return false
	}
	a.Info.Durability -= r.reduce(attributes.Cryo, a.Info.Durability, 1)
	r.reduce(attributes.Frozen, a.Info.Durability, 1)
	a.Info.Durability = 0
	return true
}

func (r *Reactable) tryStarSuperconductOnAura(a *info.AttackEvent) bool {
	if r.GetAuraDurability(info.ReactionModKeyFrozen) >= info.ZeroDur {
		return false
	}

	var consumed info.Durability
	switch a.Info.Element {
	case attributes.Electro:
		if r.GetAuraDurability(info.ReactionModKeyCryo) < info.ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Cryo, a.Info.Durability, 1)
	case attributes.Cryo:
		if r.GetAuraDurability(info.ReactionModKeyElectro) < info.ZeroDur {
			return false
		}
		consumed = r.reduce(attributes.Electro, a.Info.Durability, 1)
	default:
		return false
	}
	a.Info.Durability = max(a.Info.Durability-consumed, 0)
	return true
}

func (r *Reactable) queueStarSuperconduct(owner int) {
	// Preserve ordinary Superconduct's 0.1s reaction-damage GCD.
	if r.superconductGCD != -1 && r.core.F < r.superconductGCD {
		return
	}
	r.superconductGCD = r.core.F + 0.1*60
	coefficient := r.core.StarReactions.SuperconductCoefficient
	doStarReactionAttack(
		r.core,
		r.self,
		owner,
		info.ReactionTypeStarSuperconduct,
		attacks.AttackTagReactionStarSuperconduct,
		attributes.Cryo,
		coefficient,
		combat.NewCircleHitOnTarget(r.self, nil, 3),
	)
}
