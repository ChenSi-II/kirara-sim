package frostbreath

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const icdKey = "frostbreath-icd"

var cryoHydroReactionEvents = []event.Event{
	event.OnSuperconduct,
	event.OnMelt,
	event.OnVaporize,
	event.OnFrozen,
	event.OnElectroCharged,
	event.OnSwirlHydro,
	event.OnSwirlCryo,
	event.OnCrystallizeHydro,
	event.OnCrystallizeCryo,
	event.OnBloom,
	event.OnHyperbloom,
	event.OnBurgeon,
	event.OnLunarCharged,
	event.OnLunarBloom,
	event.OnLunarCrystallize,
}

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func atkBonus(refine int) float64   { return 0.15 + 0.05*float64(refine) }
func teamEnergy(refine int) float64 { return 4.5 + 1.5*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus(p.Refine)

	proc := func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 16*60, true)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("frostbreath-atk", 15*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return m },
		})
		for _, teammate := range c.Player.Chars() {
			if teammate.Index() != char.Index() {
				teammate.AddEnergy("frostbreath", teamEnergy(p.Refine))
			}
		}
	}
	for _, reaction := range cryoHydroReactionEvents {
		c.Events.Subscribe(reaction, proc, fmt.Sprintf(
			"frostbreath-%d-%s",
			reaction,
			char.Base.Key.String(),
		))
	}
	return w, nil
}
