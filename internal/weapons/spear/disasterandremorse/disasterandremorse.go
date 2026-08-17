package disasterandremorse

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	pathKey         = "disaster-and-remorse-path"
	unforgivableKey = "disaster-and-remorse-unforgivable"
	irreparableKey  = "disaster-and-remorse-irreparable"
	triggerICDKey   = "disaster-and-remorse-trigger-icd"
	normalHitICDKey = "disaster-and-remorse-normal-hit-icd"
	skillHitICDKey  = "disaster-and-remorse-skill-hit-icd"
	damageModKey    = "disaster-and-remorse-damage"
)

type Weapon struct {
	Index      int
	generation int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func damageBonus(refine int, hexerei bool) float64 {
	bonus := 0.30 + 0.10*float64(refine)
	if hexerei {
		bonus *= 1.75
	}
	return bonus
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = damageBonus(p.Refine, char.IsHexerei)

	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(damageModKey, -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			switch atk.Info.AttackTag {
			case attacks.AttackTagNormal, attacks.AttackTagExtra:
				if char.StatusIsActive(unforgivableKey) {
					return m
				}
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold,
				attacks.AttackTagElementalBurst:
				if char.StatusIsActive(irreparableKey) {
					return m
				}
			}
			return nil
		},
	})

	clearBuffs := func() {
		char.DeleteStatus(pathKey)
		char.DeleteStatus(unforgivableKey)
		char.DeleteStatus(irreparableKey)
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if c.Player.Active() != char.Index() || char.StatusIsActive(triggerICDKey) {
			return
		}
		char.AddStatus(triggerICDKey, 18*60, true)
		char.AddStatus(pathKey, 17*60, true)
		char.AddStatus(unforgivableKey, 3*60, true)
		char.AddStatus(irreparableKey, 3*60, true)
		w.generation++
		generation := w.generation
		char.QueueCharTask(func() {
			if generation == w.generation {
				clearBuffs()
			}
		}, 17*60)
	}, fmt.Sprintf("disaster-and-remorse-skill-%s", char.Base.Key.String()))

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != char.Index() || !char.StatusIsActive(pathKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagNormal, attacks.AttackTagExtra:
			if char.StatusIsActive(normalHitICDKey) {
				return
			}
			char.AddStatus(normalHitICDKey, 6, true)
			char.ExtendStatus(irreparableKey, 60)
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold,
			attacks.AttackTagElementalBurst:
			if char.StatusIsActive(skillHitICDKey) {
				return
			}
			char.AddStatus(skillHitICDKey, 6, true)
			char.ExtendStatus(unforgivableKey, 60)
		}
	}, fmt.Sprintf("disaster-and-remorse-hit-%s", char.Base.Key.String()))

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if args[0].(int) == char.Index() {
			w.generation++
			clearBuffs()
		}
	}, fmt.Sprintf("disaster-and-remorse-swap-%s", char.Base.Key.String()))

	return w, nil
}
