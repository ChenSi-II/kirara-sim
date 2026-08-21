package lohen

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initConstellations() {
	if c.Base.Cons >= 6 {
		c.AddAttackMod(character.AttackMod{Base: modifier.NewBase("lohen-c6-cd", -1), Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if !c.StatusIsActive(masterstrokeKey) || atk.Info.ActorIndex != c.Index() {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.CD] = 1.75
			return out
		}})
	}
}
