package hereticsmoltenblade

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	buffKey = "heretics-molten-blade-buff"
	icdKey  = "heretics-molten-blade-icd"
)

type Weapon struct {
	Index      int
	generation int
	bonus      float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func movementBonus(refine int, distance float64) float64 {
	minimum := 0.135 + 0.045*float64(refine)
	maximum := 0.27 + 0.09*float64(refine)
	ratio := min(max(distance/7, 0), 1)
	return minimum + (maximum-minimum)*ratio
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = movementBonus(p.Refine, 0)

	startTracking := func(generation int) {
		distances := make([]float64, 60)
		cursor := 0
		total := 0.0
		elapsed := 0
		last := c.Combat.Player().Pos()
		var tick func()
		tick = func() {
			if generation != w.generation || !char.StatusIsActive(buffKey) {
				return
			}
			current := c.Combat.Player().Pos()
			distance := current.Distance(last)
			last = current
			total += distance - distances[cursor]
			distances[cursor] = distance
			cursor = (cursor + 1) % len(distances)
			elapsed++
			if elapsed%60 == 0 {
				w.bonus = movementBonus(p.Refine, total)
				m[attributes.ATKP] = w.bonus
			}
			char.QueueCharTask(tick, 1)
		}
		char.QueueCharTask(tick, 1)
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 14*60, true)
		char.AddStatus(buffKey, 14*60, true)
		w.bonus = movementBonus(p.Refine, 0)
		m[attributes.ATKP] = w.bonus
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffKey+"-stat", 14*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return m },
		})
		w.generation++
		startTracking(w.generation)
	}, fmt.Sprintf("heretics-molten-blade-skill-%s", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if args[0].(int) == char.Index() {
			w.generation++
			char.DeleteStatus(buffKey)
			char.DeleteStatMod(buffKey + "-stat")
		}
	}, fmt.Sprintf("heretics-molten-blade-swap-%s", char.Base.Key.String()))

	return w, nil
}
