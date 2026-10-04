package lohen

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"strings"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 6 {
		c.AddAttackMod(character.AttackMod{Base: modifier.NewBase("lohen-c6-cd", -1), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.ActorIndex != c.Index() || (!strings.HasPrefix(atk.Info.Abil, "Etched Into Bone and Soul ") && !strings.HasPrefix(atk.Info.Abil, "Manifest Judgment ")) {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.CD] = 1.75
			return out
		}})
	}
}
