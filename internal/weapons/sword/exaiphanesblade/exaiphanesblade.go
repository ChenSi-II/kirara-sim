package exaiphanesblade

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const icdKey = "exaiphanes-blade-icd"

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

var (
	atkByRefine    = [...]float64{0, 0.16, 0.20, 0.24, 0.32, 0.40}
	energyByRefine = [...]float64{0, 3, 3, 5, 5, 5}
)

func isTraveler(key keys.Char) bool {
	switch key {
	case keys.AetherAnemo, keys.AetherDendro, keys.AetherElectro, keys.AetherGeo,
		keys.AetherHydro, keys.AetherPyro, keys.LumineAnemo, keys.LumineDendro,
		keys.LumineElectro, keys.LumineGeo, keys.LumineHydro, keys.LuminePyro:
		return true
	default:
		return false
	}
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	if !isTraveler(char.Base.Key) {
		return w, nil
	}
	r := min(max(p.Refine, 1), 5)

	if r >= 2 {
		// The game account's resonated-element count is not part of a character
		// profile, so expose it as an explicit weapon parameter.
		resonated := min(max(p.Params["resonated_elements"], 0), 7)
		m := make([]float64, attributes.EndStatType)
		m[attributes.CD] = 0.06 * float64(resonated)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("exaiphanes-blade-resonance-cd", -1),
			AffectedStat: attributes.CD,
			Amount:       func() []float64 { return m },
		})
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkByRefine[r]
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 5*60, true)
		char.AddEnergy("exaiphanes-blade", energyByRefine[r])
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("exaiphanes-blade-atk", 8*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return m },
		})
	}, fmt.Sprintf("exaiphanes-blade-%s", char.Base.Key.String()))

	return w, nil
}
