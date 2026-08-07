package reactable

import (
	"fmt"
	"slices"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

type starContribution struct {
	dmg     float64
	isCrit  bool
	charInd int
	ae      info.AttackEvent
}

// doStarReactionAttack follows the Lunar Crystallize contribution model:
// every team member calculates an independent contribution, the results are
// sorted after crit, then weighted 60%/30%/5%/5%. The reaction multiplier is
// supplied separately, so it does not affect the weights.
func doStarReactionAttack(
	c *core.Core,
	target info.Target,
	owner int,
	rt info.ReactionType,
	tag attacks.AttackTag,
	ele attributes.Element,
	mult float64,
	durability info.Durability,
	pattern info.AttackPattern,
) {
	contributions := make([]starContribution, 0, len(c.Player.Chars()))
	ai := info.AttackInfo{
		DamageSrc:        target.Key(),
		Abil:             string(rt),
		AttackTag:        tag,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          ele,
		Mult:             mult,
		Durability:       durability,
		IgnoreDefPercent: 1,
	}

	for charInd, char := range c.Player.Chars() {
		ai.ActorIndex = charInd
		snap := char.Snapshot(&ai)
		ae := info.AttackEvent{
			Info:        ai,
			Pattern:     pattern,
			SourceFrame: c.F,
			Snapshot:    snap,
		}

		c.Events.Emit(event.OnStarReactionAttack, target, &ae)
		em := ae.Snapshot.Stats[attributes.EM]
		dmg := combat.CalcSpecialReactionDmg(char.Base.Level, char.ReactBonus(ae.Info), ae.Info, em)
		isCrit := false
		if c.Rand.Float64() <= ae.Snapshot.Stats[attributes.CR] {
			dmg *= 1 + ae.Snapshot.Stats[attributes.CD]
			isCrit = true
		}
		contributions = append(contributions, starContribution{dmg: dmg, isCrit: isCrit, charInd: charInd, ae: ae})
	}

	if len(contributions) == 0 {
		return
	}

	slices.SortStableFunc(contributions, func(i, j starContribution) int {
		switch diff := j.dmg - i.dmg; {
		case diff < 0:
			return -1
		case diff > 0:
			return 1
		default:
			return 0
		}
	})

	for i := range contributions {
		contr := &contributions[i]
		c.Combat.Log.NewEvent(fmt.Sprintf("%s contributor %d", rt, i+1), glog.LogElementEvent, contr.charInd).
			Write("target", target.Key()).
			Write("damage", &contr.dmg).
			Write("crit", &contr.isCrit).
			Write("mult", reactionContributorMult[i]).
			Write("cr", &contr.ae.Snapshot.Stats[attributes.CR]).
			Write("cd", &contr.ae.Snapshot.Stats[attributes.CD]).
			Write("em", &contr.ae.Snapshot.Stats[attributes.EM]).
			Write("base_damage_bonus", &contr.ae.Info.BaseDmgBonus).
			Write("react_bonus", c.Player.Chars()[contr.charInd].ReactBonus(contr.ae.Info)).
			Write("elevation", &contr.ae.Info.Elevation)
		ai.FlatDmg += contr.dmg * reactionContributorMult[i]
	}

	ai.ActorIndex = owner
	// Mult has already been folded into each contribution and must not be
	// evaluated again by the normal attack formula.
	ai.Mult = 0
	snap := info.Snapshot{}
	if contributions[0].isCrit {
		snap.Stats[attributes.CR] = 1
	}
	c.QueueAttackWithSnap(ai, snap, pattern, 1)
}
