package nefer

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func (c *char) initAscensions() {
	c.Core.Flags.Custom[reactable.LunarBloomEnableKey] = 1
	c.Core.Events.Subscribe(event.OnDendroCore, func(...any) {
		if c.StatusIsActive(seedWindowKey) {
			c.seeds++
		}
	}, "nefer-seeds-of-deceit")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagDirectLunarBloom {
			atk.Info.BaseDmgBonus += min(c.Stat(attributes.EM)*.000175, .14)
		}
	}, "nefer-lunar-bloom-base")
}
