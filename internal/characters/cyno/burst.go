package cyno

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

var burstFrames []int

const (
	burstKey = "cyno-q"
)

func init() {
	burstFrames = frames.InitAbilSlice(86) // Q -> J
	burstFrames[action.ActionAttack] = 84
	burstFrames[action.ActionSkill] = 84
	burstFrames[action.ActionDash] = 84
	burstFrames[action.ActionSwap] = 83
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.enterPactsworn(712, false)
	c.SetCD(action.ActionBurst, 1200)
	c.ConsumeEnergy(3)

	if c.Base.Cons >= 1 {
		c.c1()
	}
	c.c6Init()

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) tryBurstPPSlide(hitmark int) {
	duration := c.StatusDuration(burstKey)
	if 0 < duration && duration < hitmark {
		c.extendPactsworn(hitmark - duration + 1)
		c.Core.Log.NewEvent("pp slide activated", glog.LogCharacterEvent, c.Index()).
			Write("expiry", c.StatusExpiry(burstKey))
		src := c.burstSrc
		c.QueueCharTask(func() {
			c.onBurstExpiry(src)
		}, hitmark-duration+3) // 3f because burst expires on 2f
	}
}

func (c *char) onExitField() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if !c.StatusIsActive(burstKey) {
			return
		}
		prev := args[0].(int)
		if prev == c.Index() {
			c.transferSunrise(args[1].(int))
			c.DeleteStatus(burstKey)
			c.onBurstExpiry(c.burstSrc)
		}
	}, "cyno-burst-clear")
}

func (c *char) onBurstExpiry(burstSrc int) {
	if burstSrc != c.burstSrc {
		return
	}
	if c.StatusIsActive(burstKey) {
		return
	}
	c.DeleteStatus(c6Key)
	c.c6Stacks = 0
	c.DeleteStatus(a1Key)
}
