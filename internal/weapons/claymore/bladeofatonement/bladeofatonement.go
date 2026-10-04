package bladeofatonement

import (
	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func emBonus(refine int) float64 { return 48 + 16*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = emBonus(p.Refine)

	common.SubscribeOwnerReactions(c, char, "blade-of-atonement", func(*info.AttackEvent) {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("blade-of-atonement-em", 12*60),
			AffectedStat: attributes.EM,
			Amount:       func() []float64 { return m },
		})
	})

	star := make([]float64, attributes.EndStatType)
	star[attributes.ATKP] = 0.12 + 0.04*float64(p.Refine)
	common.SubscribeOwnerStarReactions(c, char, "blade-of-atonement-star", func(*info.AttackEvent) {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("blade-of-atonement-atk", 12*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return star },
		})
	})
	return w, nil
}
