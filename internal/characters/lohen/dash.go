package lohen

import "github.com/genshinsim/gcsim/pkg/core/action"

func (c *char) Dash(p map[string]int) (action.Info, error) {
	ai, err := c.Character.Dash(p)
	if err != nil {
		return ai, err
	}
	// Image 8: dodge lasts 0.400s. Earlier individual cancels are unknown.
	// See PLACEHOLDER_FRAMES.md.
	ai.Frames = func(action.Action) int { return 24 }
	ai.AnimationLength = 24
	ai.CanQueueAfter = 24
	return ai, nil
}
