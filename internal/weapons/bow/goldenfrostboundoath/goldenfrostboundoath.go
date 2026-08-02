package goldenfrostboundoath

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/weapons/common"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.GoldenFrostboundOath, NewWeapon)
}

type Weapon struct {
	*common.NoEffect
	core *core.Core
	char *character.CharWrapper
}

func (w *Weapon) moondriftNearby() bool {
	moondrifts, _ := w.core.Constructs.ConstructsByType(construct.GeoConstructLunarCrystallize)
	playerPos := w.core.Combat.Player().Pos()
	for _, moondrift := range moondrifts {
		if playerPos.Distance(moondrift.Pos()) < 20 {
			return true
		}
	}
	return false
}

func lunarCrystallize(tag attacks.AttackTag) bool {
	return tag == attacks.AttackTagReactionLunarCrystallize || tag == attacks.AttackTagDirectLunarCrystallize
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		NoEffect: common.NewNoEffect(base),
		core:     c,
		char:     char,
	}
	r := float64(p.Refine)

	const buffKey = "golden-frostbound-oath-repayment"
	selfBonus := 0.3 + 0.1*r
	teamBonus := selfBonus * 0.5

	selfStats := make([]float64, attributes.EndStatType)
	selfStats[attributes.DEFP] = 0.12 + 0.04*r
	char.AddStatMod(character.StatMod{
		Base: modifier.NewBase("golden-frostbound-oath-self", -1),
		Amount: func() []float64 {
			selfStats[attributes.GeoP] = 0
			if char.StatusIsActive(buffKey) {
				selfStats[attributes.GeoP] = selfBonus
			}
			return selfStats
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("golden-frostbound-oath-self-lcr", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if char.StatusIsActive(buffKey) && lunarCrystallize(ai.AttackTag) {
				return selfBonus
			}
			return 0
		},
	})

	teamStats := make([]float64, attributes.EndStatType)
	for _, other := range c.Player.Chars() {
		if other.Index() == char.Index() {
			continue
		}
		other.AddStatMod(character.StatMod{
			Base: modifier.NewBase(fmt.Sprintf("golden-frostbound-oath-team-geo-%v", char.Index()), -1),
			Amount: func() []float64 {
				teamStats[attributes.GeoP] = 0
				if char.StatusIsActive(buffKey) && w.moondriftNearby() {
					teamStats[attributes.GeoP] = teamBonus
				}
				return teamStats
			},
		})
		other.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(fmt.Sprintf("golden-frostbound-oath-team-lcr-%v", char.Index()), -1),
			Amount: func(ai info.AttackInfo) float64 {
				if char.StatusIsActive(buffKey) && w.moondriftNearby() && lunarCrystallize(ai.AttackTag) {
					return teamBonus
				}
				return 0
			},
		})
	}

	proc := func() {
		char.AddStatus(buffKey, 6*60, true)
	}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagDirectLunarCrystallize:
			proc()
		}
	}, fmt.Sprintf("golden-frostbound-oath-hit-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		// OnLunarReactionAttack is a pre-damage contribution hook. Delay the
		// status by one frame so the hit that grants the effect is not buffed.
		char.QueueCharTask(proc, 1)
	}, fmt.Sprintf("golden-frostbound-oath-lcr-%v", char.Base.Key.String()))
	return w, nil
}
