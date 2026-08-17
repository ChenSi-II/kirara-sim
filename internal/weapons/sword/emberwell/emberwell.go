package emberwell

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

func atkBonus(refine int) float64 { return 0.12 + 0.04*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = atkBonus(p.Refine)

	common.SubscribeOwnerReactions(c, char, "emberwell", func(*info.AttackEvent) {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("emberwell-atk", 12*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return m },
		})
	})

	// Stellar Glimmer is not yet represented by an engine event or attack tag.
	return w, nil
}
