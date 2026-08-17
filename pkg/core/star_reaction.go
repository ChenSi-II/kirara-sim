package core

import "github.com/genshinsim/gcsim/pkg/core/info"

const (
	starReactionAvatar150 int32 = 10000150
	starReactionAvatar133 int32 = 10000133
	starReactionLumineID  int32 = 10000007
	// The data pipeline represents the user's 10000007-5 placeholder as the
	// Cryo Lumine skill-depot/sub-id 705.
	starReactionLumineCryoSubID int32 = 705
)

// StarReactionState is shared across all reactable targets. Both domains are
// team mechanics, so their counters must not live on an individual enemy.
type StarReactionState struct {
	Enabled bool

	SuperconductActive      bool
	SuperconductStacks      int
	SuperconductCoefficient float64

	DiffusionActive bool
	DiffusionStacks int
	DiffusionOwner  int
	DiffusionTarget info.Target
	DiffusionCycle  int
}

func isStarReactionAvatar(id, subID int32) bool {
	switch id {
	case starReactionAvatar150, starReactionAvatar133:
		return true
	case starReactionLumineID:
		return subID == starReactionLumineCryoSubID
	default:
		return false
	}
}

func (c *Core) initStarReactions() {
	for _, char := range c.Player.Chars() {
		data := char.Data()
		if data != nil && isStarReactionAvatar(data.Id, data.SubId) {
			c.StarReactions.Enabled = true
			return
		}
	}
}
