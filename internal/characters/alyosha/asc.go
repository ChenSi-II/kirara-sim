package alyosha

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("alyosha-a4", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.ActorIndex != c.Index() || (atk.Info.AttackTag != attacks.AttackTagElementalArt && atk.Info.AttackTag != attacks.AttackTagElementalBurst) {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.DmgP] = min(c.Stat(attributes.ER)*.35, .70)
			return out
		},
	})
}
