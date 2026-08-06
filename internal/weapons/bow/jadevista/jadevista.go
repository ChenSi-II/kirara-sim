package jadevista

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct {
	Index  int
	core   *core.Core
	char   *character.CharWrapper
	refine int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	sameElement := 0
	differentElement := 0
	for _, other := range w.core.Player.Chars() {
		if other.Index() == w.char.Index() {
			continue
		}
		if other.Base.Element == w.char.Base.Element {
			sameElement++
		} else {
			differentElement++
		}
	}

	// The passive has a combined three-stack cap and explicitly gives the
	// same-element EM stacks priority.
	if sameElement > 3 {
		sameElement = 3
	}
	remaining := 3 - sameElement
	if differentElement > remaining {
		differentElement = remaining
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = float64(sameElement) * (48 + 16*float64(w.refine))
	m[attributes.ATKP] = float64(differentElement) * (0.09 + 0.03*float64(w.refine))
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("jade-vista", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() []float64 {
			return m
		},
	})

	return nil
}

// For every other party member of the same Elemental Type, EM is increased by
// 64/80/96/112/128. For every other party member of a different Elemental
// Type, ATK is increased by 12/15/18/21/24%. The effects have a combined
// three-stack cap, with EM stacks applied first.
func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{
		core:   c,
		char:   char,
		refine: p.Refine,
	}, nil
}
