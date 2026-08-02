package faescrystalle

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
	core.RegisterWeaponFunc(keys.FaesCrystalle, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)

	const (
		buffKey = "faes-crystalle-buff"
		icdKey  = "faes-crystalle-icd"
	)

	m := make([]float64, attributes.EndStatType)
	m[attributes.ATKP] = 0.15 + 0.05*r
	energy := 4.5 + 1.5*r

	onReaction := func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(icdKey) {
			return
		}
		char.AddStatus(icdKey, 16*60, true)
		char.AddStatMod(character.StatMod{
			Base: modifier.NewBaseWithHitlag(buffKey, 15*60),
			Amount: func() []float64 {
				return m
			},
		})
		for _, other := range c.Player.Chars() {
			if other.Index() == char.Index() {
				continue
			}
			other.AddEnergy("faes-crystalle", energy)
		}
	}

	// Reactions whose inputs directly include Cryo or Hydro. Shatter is omitted
	// because gcsim, like the game, does not normally classify it as an Elemental
	// Reaction for reaction-triggered passives.
	reactions := []event.Event{
		event.OnSuperconduct,
		event.OnMelt,
		event.OnVaporize,
		event.OnFrozen,
		event.OnElectroCharged,
		event.OnSwirlHydro,
		event.OnSwirlCryo,
		event.OnCrystallizeHydro,
		event.OnCrystallizeCryo,
		event.OnBloom,
		event.OnLunarCharged,
		event.OnLunarBloom,
		event.OnLunarCrystallize,
		event.OnStarSuperconduct,
		event.OnStarDiffusion,
	}
	for _, e := range reactions {
		c.Events.Subscribe(e, onReaction, fmt.Sprintf("faes-crystalle-%v-%v", char.Base.Key.String(), e))
	}
	return w, nil
}
