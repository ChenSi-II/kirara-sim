package vodyanitsa

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) initAscensions() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		if !c.StatusIsActive(microphoneKey) {
			return
		}
		atk, ok := args[1].(*info.AttackEvent)
		if !ok {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
		default:
			return
		}
		target, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}
		target.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag("vodyanitsa-flowing-vortex-anemo-res", 6*60),
			Ele:   attributes.Anemo,
			Value: -0.30,
		})
	}, "vodyanitsa-a1-flowing-vortex")
}
