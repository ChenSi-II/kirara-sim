package linnea

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	burstHealHitmark = 48
	burstHealKey     = "linnea-burst-heal"
)

func (c *char) Burst(map[string]int) (action.Info, error) {
	c.SetCD(action.ActionBurst, int(burstParam[5][c.TalentLvlBurst()]*60))
	c.ConsumeEnergy(60)

	c.Core.Tasks.Add(func() {
		if c.StatusIsActive(lumiKey) {
			c.refreshLumi()
		} else {
			c.summonLumi(lumiSuper)
		}
		c.initialBurstHeal()
		c.startContinuousHealing()
	}, burstHealHitmark)

	// TODO: replace conservative burst frames once verified frame data is available.
	f := frames.InitAbilSlice(90)
	return action.Info{
		Frames:          frames.NewAbilFunc(f),
		AnimationLength: 90,
		CanQueueAfter:   78,
		State:           action.BurstState,
	}, nil
}

func (c *char) initialBurstHeal() {
	lvl := c.TalentLvlBurst()
	src := burstParam[0][lvl] + burstParam[1][lvl]*c.TotalDef(false)
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index(),
		Target:  -1,
		Message: "Survival Guide Initial Heal",
		Src:     src,
		Bonus:   c.Stat(attributes.Heal),
	})
}

func (c *char) startContinuousHealing() {
	src := c.Core.F
	dur := int(burstParam[4][c.TalentLvlBurst()] * 60)
	c.AddStatus(burstHealKey, dur, true)
	// The local talent data defines duration but not interval. Two seconds is
	// retained as an explicit provisional runtime value pending frame evidence.
	for delay := 2 * 60; delay <= dur; delay += 2 * 60 {
		c.QueueCharTask(func() {
			if !c.StatusIsActive(burstHealKey) || src+dur < c.Core.F {
				return
			}
			lvl := c.TalentLvlBurst()
			heal := burstParam[2][lvl] + burstParam[3][lvl]*c.TotalDef(false)
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index(),
				Target:  c.Core.Player.Active(),
				Message: "Survival Guide Continuous Heal",
				Src:     heal,
				Bonus:   c.Stat(attributes.Heal),
			})
		}, delay)
	}
}
