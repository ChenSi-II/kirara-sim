package clashofkings

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

const (
	buffStatusKey = "clash-of-kings-buff-status"
	buffStatKey   = "clash-of-kings-buff"
	icdKey        = "clash-of-kings-icd"
)

type Weapon struct {
	Index    int
	extended bool
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func atkBonus(refine int) float64 { return 0.15 + 0.05*float64(refine) }
func emBonus(refine int) float64  { return 75 + 25*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus(p.Refine)
	m[attributes.EM] = emBonus(p.Refine)
	addBuff := func(duration int) {
		char.AddStatus(buffStatusKey, duration, true)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(buffStatKey, duration),
			AffectedStat: attributes.NoStat,
			Amount:       func() []float64 { return m },
		})
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 12*60, true)
		w.extended = false
		addBuff(6 * 60)
	}, fmt.Sprintf("clash-of-kings-skill-%s", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() ||
			atk.Info.AttackTag != attacks.AttackTagExtra ||
			!char.StatusIsActive(buffStatusKey) || w.extended {
			return
		}
		w.extended = true
		addBuff(char.StatusDuration(buffStatusKey) + 6*60)
	}, fmt.Sprintf("clash-of-kings-charge-%s", char.Base.Key.String()))

	return w, nil
}
