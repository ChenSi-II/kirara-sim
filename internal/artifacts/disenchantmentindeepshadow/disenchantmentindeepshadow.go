package disenchantmentindeepshadow

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	setKey2                     = "disenchantment-in-deep-shadow-2pc"
	setKey4                     = "disenchantment-in-deep-shadow-4pc"
	starSuperconductAffectedKey = setKey4 + "-star-superconduct-affected"
	superconductShredKey        = "superconduct-phys-shred"
)

func init() {
	core.RegisterSetFunc(keys.DisenchantmentInDeepShadow, NewSet)
}

type Set struct {
	char  *character.CharWrapper
	core  *core.Core
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }

func (s *Set) Init() error {
	if s.Count < 2 {
		return nil
	}

	atk := make([]float64, attributes.EndStatType)
	atk[attributes.ATKP] = 0.18
	s.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(setKey2, -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			return atk
		},
	})

	if s.Count < 4 {
		return nil
	}

	// Star Superconduct does not use the physical RES modifier that marks an
	// ordinary Superconduct target, so keep an equivalent target status here.
	s.core.Events.Subscribe(event.OnStarSuperconduct, func(args ...any) {
		target, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}
		target.AddStatus(starSuperconductAffectedKey, 12*60, true)
	}, setKey4+"-mark-"+s.char.Base.Key.String())

	s.char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(setKey4+"-reaction-dmg", -1),
		Amount: func(ai info.AttackInfo) float64 {
			switch ai.AttackTag {
			case attacks.AttackTagSuperconductDamage:
				return 0.8
			case attacks.AttackTagReactionStarSuperconduct:
				return 0.4
			default:
				return 0
			}
		},
	})

	crit := make([]float64, attributes.EndStatType)
	crit[attributes.CR] = 0.16
	s.char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(setKey4+"-crit", -1),
		Amount: func(_ *info.AttackEvent, target info.Target) []float64 {
			e, ok := target.(*enemy.Enemy)
			if !ok {
				return nil
			}
			if e.ResistModIsActive(superconductShredKey) || e.StatusIsActive(starSuperconductAffectedKey) {
				return crit
			}
			return nil
		},
	})

	// Star reaction contributions are calculated before the consolidated hit,
	// so AttackMods cannot affect their snapshots through the normal path.
	s.core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		target, ok := args[0].(*enemy.Enemy)
		if !ok || (!target.ResistModIsActive(superconductShredKey) && !target.StatusIsActive(starSuperconductAffectedKey)) {
			return
		}
		ae, ok := args[1].(*info.AttackEvent)
		if !ok || ae.Info.ActorIndex != s.char.Index() {
			return
		}
		ae.Snapshot.Stats[attributes.CR] += 0.16
	}, setKey4+"-star-crit-"+s.char.Base.Key.String())

	return nil
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	return &Set{core: c, char: char, Count: count}, nil
}
