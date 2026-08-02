package flameforgedinsight

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

func init() {
	core.RegisterWeaponFunc(keys.FlameForgedInsight, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const flameForgedInsightICDKey = "flame-forged-insight-icd"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	energy := 9 + 3*r
	em := make([]float64, attributes.EndStatType)
	em[attributes.EM] = 45 + 15*r

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(flameForgedInsightICDKey) {
			return
		}
		char.AddStatus(flameForgedInsightICDKey, 15*60, true)
		char.AddEnergy("flame-forged-insight", energy)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("flame-forged-insight-em", 15*60),
			AffectedStat: attributes.EM,
			Amount:       func() []float64 { return em },
		})
	}

	reactions := []event.Event{
		event.OnElectroCharged,
		event.OnLunarCharged,
		event.OnBloom,
		event.OnLunarBloom,
		event.OnCrystallizeHydro,
		event.OnCrystallizeCryo,
		event.OnCrystallizeElectro,
		event.OnCrystallizePyro,
		event.OnLunarCrystallize,
	}
	for _, evt := range reactions {
		c.Events.Subscribe(evt, onReaction, fmt.Sprintf("flame-forged-insight-%v-%v", evt, char.Base.Key.String()))
	}

	return w, nil
}
