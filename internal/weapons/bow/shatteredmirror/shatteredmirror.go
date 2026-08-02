package shatteredmirror

import (
	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.ShatteredMirror, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
	core   *core.Core
	char   *character.CharWrapper
	refine int
}

func (w *Weapon) Init() error {
	stacks := 0
	for _, other := range w.core.Player.Chars() {
		// Unknown is not a real shared region (Traveler and Aloy currently map
		// to it), but the wielder always counts themselves.
		if other.Index() == w.char.Index() || (w.char.CharZone != info.ZoneUnknown && other.CharZone == w.char.CharZone) {
			stacks++
		}
	}
	stacks = min(stacks, 3)
	r := float64(w.refine)
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = (0.06 + 0.02*r) * float64(stacks)
	m[attributes.EM] = (24 + 8*r) * float64(stacks)
	w.char.AddStatMod(character.StatMod{
		Base: modifier.NewBase("shattered-mirror-unity", -1),
		Amount: func() []float64 {
			return m
		},
	})
	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		NoEffect: common.NewNoEffect(base),
		core:     c,
		char:     char,
		refine:   p.Refine,
	}
	return w, nil
}
