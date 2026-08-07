package bloodredproof

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	setKey2 = "blood-red-proof-2pc"
	setKey4 = "blood-red-proof-4pc"
	buffKey = setKey4 + "-buff"
)

func init() {
	core.RegisterSetFunc(keys.BloodRedProof, NewSet)
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
		Amount:       func() []float64 { return atk },
	})

	if s.Count < 4 {
		return nil
	}

	crit := make([]float64, attributes.EndStatType)
	crit[attributes.CR] = 0.16
	s.core.Events.Subscribe(event.OnStarDiffusion, func(args ...any) {
		ae, ok := args[1].(*info.AttackEvent)
		if !ok || ae.Info.ActorIndex != s.char.Index() {
			return
		}
		s.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(buffKey, 10*60),
			AffectedStat: attributes.CR,
			Amount:       func() []float64 { return crit },
		})
	}, setKey4+"-trigger-"+s.char.Base.Key.String())

	s.char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(setKey4+"-reaction-dmg", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !s.char.StatModIsActive(buffKey) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				return 0.4
			default:
				return 0
			}
		},
	})

	return nil
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	return &Set{core: c, char: char, Count: count}, nil
}
