package frostboundoath

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.FrostboundOath, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = 90 + 30*r
	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag("frostbound-oath-em", 12*60),
			Amount: func() []float64 {
				return m
			},
		})
	}, fmt.Sprintf("frostbound-oath-skill-%v", char.Base.Key.String()))
	return w, nil
}
