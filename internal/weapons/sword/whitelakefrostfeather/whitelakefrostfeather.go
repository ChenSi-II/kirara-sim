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
		}
	}, fmt.Sprintf("whitelake-frostfeather-%s", char.Base.Key.String()))

	// Stellar Glimmer CRIT DMG and its energy trigger are blocked on missing
	// engine reaction events and attack tags. The independent ATK stacks are
	// fully modeled and can trigger while the holder is off-field.
	return w, nil
}
