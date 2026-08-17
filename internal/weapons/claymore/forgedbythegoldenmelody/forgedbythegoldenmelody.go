package forgedbythegoldenmelody

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct {
	Index  int
	char   *character.CharWrapper
	refine int
	phase  int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error {
	w.char.QueueCharTask(w.tick, 10*60)
	return nil
}

func atkBonus(refine int) float64 { return 0.135 + 0.045*float64(refine) }
func emBonus(refine int) float64  { return 90 + 30*float64(refine) }

func (w *Weapon) tick() {
	m := make([]float64, attributes.EndStatType)
	switch w.phase {
	case 0:
		m[attributes.ATKP] = atkBonus(w.refine)
	case 1:
		m[attributes.EM] = emBonus(w.refine)
	case 2:
		// Stellar Glimmer reaction DMG has no engine attack tag yet.
	}
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("forged-by-the-golden-melody-movement", 10*60),
		AffectedStat: attributes.NoStat,
		Amount:       func() []float64 { return m },
	})
	w.phase = (w.phase + 1) % 3
	w.char.QueueCharTask(w.tick, 10*60)
}

func NewWeapon(_ *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{
		char:   char,
		refine: p.Refine,
	}, nil
}
