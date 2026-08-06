package songofthevigil

import (
	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

const energyICDKey = "song-of-the-vigil-energy-icd"

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func energyRestore(refine int) float64 { return 3 + float64(refine) }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	common.SubscribeOwnerReactions(c, char, "song-of-the-vigil", func(*info.AttackEvent) {
		if char.StatusIsActive(energyICDKey) {
			return
		}
		char.AddStatus(energyICDKey, 9*60, true)
		char.AddEnergy("song-of-the-vigil", energyRestore(p.Refine))
	})

	// Stellar Glimmer is not yet represented by an engine event.
	return w, nil
}
