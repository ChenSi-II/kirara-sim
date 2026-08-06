package echoesoftheheart

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

func emBonus(refine int) float64 { return 45 + 15*float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = emBonus(p.Refine)

	common.SubscribeOwnerReactions(c, char, "echoes-of-the-heart", func(*info.AttackEvent) {
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("echoes-of-the-heart-em", 12*60),
			AffectedStat: attributes.EM,
			Amount:       func() []float64 { return m },
		})
	})

	// Stellar Glimmer is not yet represented by an engine event or attack tag.
	return w, nil
}
