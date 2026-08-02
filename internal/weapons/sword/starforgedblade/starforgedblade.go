package starforgedblade

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.StarforgedBlade, NewWeapon)
}

type Weapon struct {
	Index int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const starforgedICDKey = "starforged-blade-icd"

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	if char.Base.Key <= keys.NoChar || char.Base.Key >= keys.TravelerDelim {
		return w, nil
	}

	refine := p.Refine - 1
	atkByRefine := [...]float64{0.16, 0.20, 0.24, 0.32, 0.40}
	energyByRefine := [...]float64{3, 3, 5, 5, 5}
	atk := make([]float64, attributes.EndStatType)
	atk[attributes.ATKP] = atkByRefine[refine]

	// TODO: The account's Traveler resonance history is not represented by the
	// simulator. Default to zero rather than inventing completed resonances, but
	// allow explicit modeling through the weapon's resonated_elements parameter.
	resonatedElements := p.Params["resonated_elements"]
	resonatedElements = max(0, min(resonatedElements, 7))
	if p.Refine >= 2 && resonatedElements > 0 {
		cd := make([]float64, attributes.EndStatType)
		cd[attributes.CD] = 0.06 * float64(resonatedElements)
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("starforged-blade-resonance-cd", -1),
			AffectedStat: attributes.CD,
			Amount:       func() []float64 { return cd },
		})
	}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.ActorIndex != char.Index() || char.StatusIsActive(starforgedICDKey) {
			return
		}
		char.AddStatus(starforgedICDKey, 5*60, true)
		char.AddEnergy("starforged-blade", energyByRefine[refine])
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag("starforged-blade-atk", 8*60),
			AffectedStat: attributes.ATKP,
			Amount:       func() []float64 { return atk },
		})
	}, fmt.Sprintf("starforged-blade-hit-%v", char.Base.Key.String()))

	return w, nil
}
