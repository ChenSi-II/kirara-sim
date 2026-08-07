package glacierandsnowfield

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
	setKey2 = "glacier-and-snowfield-2pc"
	setKey4 = "glacier-and-snowfield-4pc"
)

func init() {
	core.RegisterSetFunc(keys.GlacierAndSnowfield, NewSet)
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

	cryo := make([]float64, attributes.EndStatType)
	cryo[attributes.CryoP] = 0.15
	s.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(setKey2, -1),
		AffectedStat: attributes.CryoP,
		Amount:       func() []float64 { return cryo },
	})

	if s.Count < 4 {
		return nil
	}

	s.char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(setKey4+"-reaction-dmg", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag == attacks.AttackTagSuperconductDamage {
				return 1
			}
			if ai.Amped && ai.AmpType == info.ReactionTypeMelt {
				return 0.15
			}
			return 0
		},
	})

	burstBuff := make([]float64, attributes.EndStatType)
	burstBuff[attributes.CryoP] = 0.30
	s.core.Events.Subscribe(event.OnBurst, func(args ...any) {
		if s.core.Player.Active() != s.char.Index() {
			return
		}
		s.char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(setKey4+"-burst", 10*60),
			AffectedStat: attributes.CryoP,
			Amount:       func() []float64 { return burstBuff },
		})
	}, setKey4+"-trigger-"+s.char.Base.Key.String())

	return nil
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	return &Set{core: c, char: char, Count: count}, nil
}
