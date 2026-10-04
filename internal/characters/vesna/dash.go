package vesna

import "github.com/genshinsim/gcsim/pkg/core/action"

// Image 2 measures the armed dash at 0.350s (21 frames). Other states and
// dash-to-jump cancels retain the template defaults; see PLACEHOLDER_FRAMES.md.
func (c *char) Dash(p map[string]int) (action.Info, error) {
	ai, err := c.Character.Dash(p)
	if err != nil || !c.StatusIsActive(spiritbladeArmedKey) {
		return ai, err
	}
	baseFrames := ai.Frames
	ai.Frames = func(next action.Action) int {
		if next == action.ActionJump {
			return baseFrames(next)
		}
		return 21
	}
	ai.AnimationLength = 21
	return ai, nil
}
