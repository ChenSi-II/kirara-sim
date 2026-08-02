package disasterandremorse

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.DisasterAndRemorse, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{NoEffect: common.NewNoEffect(base)}
	r := float64(p.Refine)

	const (
		roadKey          = "calamitous-remorse-road"
		noMercyKey       = "calamitous-remorse-no-mercy"
		noRemedyKey      = "calamitous-remorse-no-remedy"
		procICDKey       = "calamitous-remorse-proc-icd"
		noMercyExtendICD = "calamitous-remorse-no-mercy-extend-icd"
		noRemedyExtICD   = "calamitous-remorse-no-remedy-extend-icd"
	)

	// TODO: The 75% increase granted by Magical Mystery is not applied because
	// gcsim has no team-level Magical Mystery state yet.
	dmgBonus := 0.3 + 0.1*r
	m := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("calamitous-remorse-dmg", -1),
		Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
			active := false
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra:
				active = char.StatusIsActive(noMercyKey)
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
				active = char.StatusIsActive(noRemedyKey)
			}
			if !active {
				return nil
			}
			m[attributes.DmgP] = dmgBonus
			return m
		},
	})

	clear := func() {
		char.DeleteStatus(roadKey)
		char.DeleteStatus(noMercyKey)
		char.DeleteStatus(noRemedyKey)
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(procICDKey) {
			return
		}
		char.AddStatus(procICDKey, 18*60, true)
		char.AddStatus(roadKey, 17*60, true)
		char.AddStatus(noMercyKey, 3*60, true)
		char.AddStatus(noRemedyKey, 3*60, true)
		expiry := char.StatusExpiry(roadKey)
		char.QueueCharTask(func() {
			if char.StatusExpiry(roadKey) == expiry {
				clear()
			}
		}, 17*60)
	}, fmt.Sprintf("calamitous-remorse-skill-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || !char.StatusIsActive(roadKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal, attacks.AttackTagExtra:
			if char.StatusIsActive(noRemedyExtICD) {
				return
			}
			char.AddStatus(noRemedyExtICD, 0.1*60, true)
			char.ExtendStatus(noRemedyKey, 60)
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
			if char.StatusIsActive(noMercyExtendICD) {
				return
			}
			char.AddStatus(noMercyExtendICD, 0.1*60, true)
			char.ExtendStatus(noMercyKey, 60)
		}
	}, fmt.Sprintf("calamitous-remorse-hit-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if args[0].(int) == char.Index() {
			clear()
		}
	}, fmt.Sprintf("calamitous-remorse-swap-%v", char.Base.Key.String()))
	return w, nil
}
