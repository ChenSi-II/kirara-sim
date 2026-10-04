package whitelakefrostfeather

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct {
	Index    int
	core     *core.Core
	expiries []int
	lastProc int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func atkPerStack(refine int) float64 { return 0.06 + 0.02*float64(refine) }

func (w *Weapon) activeStacks() int {
	active := w.expiries[:0]
	for _, expiry := range w.expiries {
		if expiry > w.core.F {
			active = append(active, expiry)
		}
	}
	w.expiries = active
	return len(active)
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{core: c, lastProc: -1000}
	perStack := atkPerStack(p.Refine)
	m := make([]float64, attributes.EndStatType)

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("whitelake-frostfeather-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			m[attributes.ATKP] = float64(w.activeStacks()) * perStack
			return m
		},
	})

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
		default:
			return
		}
		if c.F-w.lastProc < 6 {
			return
		}
		w.lastProc = c.F
		if w.activeStacks() < 3 {
			w.expiries = append(w.expiries, c.F+8*60)
		} else {
			w.expiries = append(w.expiries[1:], c.F+8*60)
		}
	}, fmt.Sprintf("whitelake-frostfeather-%s", char.Base.Key.String()))

	critBonus := 0.35 + 0.15*float64(p.Refine)
	crit := make([]float64, attributes.EndStatType)
	crit[attributes.CD] = critBonus
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("whitelake-frostfeather-star-cd", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			// Combined reactions already evaluated CRIT per contributor.
			if atk.Info.IsDirectStarDamage() && w.activeStacks() == 3 {
				return crit
			}
			return nil
		},
	})
	c.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == char.Index() && attacks.AttackTagIsStar(atk.Info.AttackTag) && w.activeStacks() == 3 {
			atk.Snapshot.Stats[attributes.CD] += critBonus
		}
	}, fmt.Sprintf("whitelake-frostfeather-star-cd-%v", char.Index()))
	energy := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || w.activeStacks() < 3 || char.StatusDuration("whitelake-frostfeather-energy-icd") > 0 {
			return
		}
		char.AddEnergy("whitelake-frostfeather", 3.5+0.5*float64(p.Refine))
		char.AddStatus("whitelake-frostfeather-energy-icd", 210, false)
	}
	c.Events.Subscribe(event.OnStarSuperconduct, energy, fmt.Sprintf("whitelake-frostfeather-conduct-%v", char.Index()))
	c.Events.Subscribe(event.OnStarDiffusion, energy, fmt.Sprintf("whitelake-frostfeather-diffusion-%v", char.Index()))
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if attacks.AttackTagIsStar(args[1].(*info.AttackEvent).Info.AttackTag) {
			energy(args...)
		}
	}, fmt.Sprintf("whitelake-frostfeather-star-damage-%v", char.Index()))
	return w, nil
}
