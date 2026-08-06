package common

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

// SubscribeOwnerReactions subscribes to every elemental-reaction event and
// filters the callback to reactions attributed to the weapon holder.
func SubscribeOwnerReactions(
	c *core.Core,
	char *character.CharWrapper,
	key string,
	callback func(*info.AttackEvent),
) {
	for reaction := event.ReactionEventStartDelim + 1; reaction < event.ReactionEventEndDelim; reaction++ {
		reaction := reaction
		c.Events.Subscribe(reaction, func(args ...any) {
			if len(args) < 2 {
				return
			}
			atk, ok := args[1].(*info.AttackEvent)
			if !ok || atk.Info.ActorIndex != char.Index() {
				return
			}
			callback(atk)
		}, fmt.Sprintf("%s-%d-%s", key, reaction, char.Base.Key.String()))
	}
}
