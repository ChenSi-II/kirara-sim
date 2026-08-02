package thedaybreakchronicles

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
	core.RegisterWeaponFunc(keys.TheDaybreakChronicles, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
	core  *core.Core
	char  *character.CharWrapper
	step  float64
	max   float64
	bonus [3]float64
}

const (
	normalBonus = iota
	skillBonus
	burstBonus
	endBonusType
)

func attackType(tag attacks.AttackTag) int {
	switch tag {
	case attacks.AttackTagNormal:
		return normalBonus
	case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold:
		return skillBonus
	case attacks.AttackTagElementalBurst:
		return burstBonus
	default:
		return -1
	}
}

func (w *Weapon) decay() {
	for i := range w.bonus {
		w.bonus[i] = max(0, w.bonus[i]-w.step)
	}
	w.core.Tasks.Add(w.decay, 60)
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	r := float64(p.Refine)
	w := &Weapon{
		NoEffect: common.NewNoEffect(base),
		core:     c,
		char:     char,
		step:     0.075 + 0.025*r,
		max:      0.45 + 0.15*r,
	}

	// TODO: gcsim has no combat-state transition event. Conservatively assume
	// the simulation starts and remains in combat: the three bonuses start at
	// zero, increase from hits, and decay once per second. Consequently, the
	// out-of-combat 3s reset to the maximum cannot currently be represented.
	// TODO: Magical Mystery is also not represented at team level, so a hit
	// increases only its matching category instead of increasing all three.
	m := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("daybreak-chronicles-dmg", -1),
		Amount: func(atk *info.AttackEvent, t info.Target) []float64 {
			typ := attackType(atk.Info.AttackTag)
			if typ < 0 {
				return nil
			}
			m[attributes.DmgP] = w.bonus[typ]
			return m
		},
	})

	icdKeys := [endBonusType]string{
		"daybreak-chronicles-normal-icd",
		"daybreak-chronicles-skill-icd",
		"daybreak-chronicles-burst-icd",
	}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		typ := attackType(atk.Info.AttackTag)
		if typ < 0 || char.StatusIsActive(icdKeys[typ]) {
			return
		}
		char.AddStatus(icdKeys[typ], 0.1*60, true)
		w.bonus[typ] = min(w.max, w.bonus[typ]+w.step)
	}, fmt.Sprintf("daybreak-chronicles-hit-%v", char.Base.Key.String()))
	c.Tasks.Add(w.decay, 60)
	return w, nil
}
