package escapewheel

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
	core.RegisterWeaponFunc(keys.EscapeWheel, NewWeapon)
}

type Weapon struct {
	Index int

	char      *character.CharWrapper
	cycle     int
	cycleSrc  int
	statBuff  []float64
	atkBonus  float64
	emBonus   float64
	starBonus float64
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	w.scheduleCycle(10 * 60)
	return nil
}

const escapeWheelFreezeICDKey = "escape-wheel-freeze-icd"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	r := float64(p.Refine)
	w := &Weapon{
		char:      char,
		statBuff:  make([]float64, attributes.EndStatType),
		atkBonus:  0.15 + 0.05*r,
		emBonus:   90 + 30*r,
		starBonus: 0.24 + 0.08*r,
	}

	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("escape-wheel-stat", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() []float64 {
			clear(w.statBuff)
			switch w.cycle {
			case 0:
				w.statBuff[attributes.ATKP] = w.atkBonus
			case 1:
				w.statBuff[attributes.EM] = w.emBonus
			}
			return w.statBuff
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("escape-wheel-star-reaction", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if w.cycle == 2 && attacks.AttackTagIsStar(ai.AttackTag) {
				return w.starBonus
			}
			return 0
		},
	})

	freezeCurrent := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(escapeWheelFreezeICDKey) {
			return
		}
		char.AddStatus(escapeWheelFreezeICDKey, 12*60, true)
		// Preserve the current mode for 12 seconds from the reaction. Incrementing
		// the source invalidates the previously queued 10-second transition.
		w.cycleSrc++
		w.scheduleCycle(12 * 60)
	}
	c.Events.Subscribe(event.OnStarSuperconduct, freezeCurrent, fmt.Sprintf("escape-wheel-star-sc-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, freezeCurrent, fmt.Sprintf("escape-wheel-star-diff-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) scheduleCycle(delay int) {
	src := w.cycleSrc
	w.char.QueueCharTask(func() {
		if src != w.cycleSrc {
			return
		}
		w.cycle = (w.cycle + 1) % 3
		w.cycleSrc++
		w.scheduleCycle(10 * 60)
	}, delay)
}
