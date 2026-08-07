package artifact15048

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
	setKey2       = "artifact-15048-2pc"
	setKey4       = "artifact-15048-4pc"
	holderBuffKey = setKey4 + "-holder"
	teamBuffKey   = setKey4 + "-team-reaction-dmg"
)

func init() {
	core.RegisterSetFunc(keys.Artifact15048, NewSet)
}

type Set struct {
	char  *character.CharWrapper
	core  *core.Core
	Index int
	Count int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }

func isSupportedStarReaction(tag attacks.AttackTag) bool {
	switch tag {
	case attacks.AttackTagReactionStarSuperconduct,
		attacks.AttackTagReactionStarDiffusionAnemo,
		attacks.AttackTagReactionStarDiffusionCryo:
		return true
	default:
		return false
	}
}

func (s *Set) activate() {
	atk := make([]float64, attributes.EndStatType)
	atk[attributes.ATKP] = 0.12
	s.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(holderBuffKey, 12*60),
		AffectedStat: attributes.ATKP,
		Amount:       func() []float64 { return atk },
	})
}

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

	triggeredReaction := func(args ...any) {
		ae, ok := args[1].(*info.AttackEvent)
		if ok && ae.Info.ActorIndex == s.char.Index() {
			s.activate()
		}
	}
	s.core.Events.Subscribe(event.OnStarSuperconduct, triggeredReaction, setKey4+"-superconduct-"+s.char.Base.Key.String())
	s.core.Events.Subscribe(event.OnStarDiffusion, triggeredReaction, setKey4+"-diffusion-"+s.char.Base.Key.String())

	// Direct Star reaction damage is represented by a per-character
	// contribution event, distinct from the character that triggered it.
	s.core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		ae, ok := args[1].(*info.AttackEvent)
		if ok && ae.Info.ActorIndex == s.char.Index() && isSupportedStarReaction(ae.Info.AttackTag) {
			s.activate()
		}
	}, setKey4+"-direct-"+s.char.Base.Key.String())

	// The description explicitly makes this effect unique. Every holder uses
	// the same modifier key and the condition only asks whether any holder buff
	// is active, so multiple sets refresh rather than stack the team bonus.
	for _, char := range s.core.Player.Chars() {
		char.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(teamBuffKey, -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !isSupportedStarReaction(ai.AttackTag) {
					return 0
				}
				for _, holder := range s.core.Player.Chars() {
					if holder.StatModIsActive(holderBuffKey) {
						return 0.5
					}
				}
				return 0
			},
		})
	}

	return nil
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	return &Set{core: c, char: char, Count: count}, nil
}
