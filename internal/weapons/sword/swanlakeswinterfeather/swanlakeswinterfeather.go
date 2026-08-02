package swanlakeswinterfeather

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.SwanlakesWinterFeather, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
	atk    []float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	preludeKey   = "swanlakes-winter-feather-prelude"
	preludeICD   = "swanlakes-winter-feather-prelude-icd"
	curtainKey   = "swanlakes-winter-feather-curtain-call"
	energyICDKey = "swanlakes-winter-feather-energy-icd"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{atk: make([]float64, attributes.EndStatType)}
	r := float64(p.Refine)
	energy := 3 + r
	stackATK := 0.06 + 0.02*r
	starCD := 0.30 + 0.10*r

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
		default:
			return
		}
		if char.StatusIsActive(preludeICD) {
			return
		}
		char.AddStatus(preludeICD, int(0.1*60), true)
		if !char.StatModIsActive(preludeKey) {
			w.stacks = 0
		}
		if w.stacks < 3 {
			w.stacks++
		}
		w.atk[attributes.ATKP] = stackATK * float64(w.stacks)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(preludeKey, 8*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return w.atk },
		})
		if w.stacks == 3 {
			char.AddStatus(curtainKey, 8*60, true)
		}
	}, fmt.Sprintf("swanlakes-winter-feather-skill-hit-%v", char.Base.Key.String()))

	// Star reactions use per-character contribution snapshots. Modify only the
	// wielder's contribution so the CRIT DMG does not leak to party members.
	c.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || !char.StatusIsActive(curtainKey) {
			return
		}
		atk.Snapshot.Stats[attributes.CD] += starCD
		w.restoreEnergy(char, energy)
	}, fmt.Sprintf("swanlakes-winter-feather-star-dmg-%v", char.Base.Key.String()))

	onStarReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == char.Index() && char.StatusIsActive(curtainKey) {
			w.restoreEnergy(char, energy)
		}
	}
	c.Events.Subscribe(event.OnStarSuperconduct, onStarReaction, fmt.Sprintf("swanlakes-winter-feather-star-sc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, onStarReaction, fmt.Sprintf("swanlakes-winter-feather-star-diff-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) restoreEnergy(char *character.CharWrapper, amount float64) {
	if char.StatusIsActive(energyICDKey) {
		return
	}
	char.AddStatus(energyICDKey, int(3.5*60), true)
	char.AddEnergy("swanlakes-winter-feather", amount)
}
