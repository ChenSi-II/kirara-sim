package ateaspoonoftranscendence

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func atkBonus(refine int) float64 { return 0.21 + 0.07*float64(refine) }

func NewWeapon(_ *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus(p.Refine)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("a-teaspoon-of-transcendence-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount:       func() []float64 { return m },
	})

	// Stellar-Conduct and Stellar Swirl have no engine events or reaction
	// attack tags yet, so their charged-attack stack bonus cannot be applied.
	return w, nil
}
