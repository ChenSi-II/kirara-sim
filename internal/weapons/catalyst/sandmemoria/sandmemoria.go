package sandmemoria

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
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
	core.RegisterWeaponFunc(keys.SandMemoria, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
	extended bool
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)

	const (
		buffKey = "sand-memoria-buff"
		icdKey  = "sand-memoria-icd"
	)
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.15 + 0.05*r
	m[attributes.EM] = 75 + 25*r

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 12*60, true)
		w.extended = false
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 6*60),
			Amount: func() []float64 {
				return m
			},
		})
	}, fmt.Sprintf("sand-memoria-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || atk.Info.AttackTag != attacks.AttackTagExtra {
			return
		}
		if w.extended || !char.StatModIsActive(buffKey) {
			return
		}
		w.extended = true
		char.ExtendStatus(buffKey, 6*60)
	}, fmt.Sprintf("sand-memoria-charge-%v", char.Base.Key.String()))
	return w, nil
}
