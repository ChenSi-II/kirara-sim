package sandrone

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	skillStart         int
	resolutionSrc      int
	resolutionPower    float64
	resolutionRays     int
	resolutionRayGoal  int
	tacticStacks       int
	tacticExpiry       int
	tacticPowerRemoved float64
	resolutionTick     int
	powerOverdrive     bool
	resolutionChannel  bool
	sweepTailUntil     int
}

func (c *char) reduceResolutionPower(amount float64) {
	removed := min(amount, c.resolutionPower)
	c.resolutionPower -= removed
	if c.powerOverdrive && c.resolutionPower < 50 {
		c.powerOverdrive = false
	}
	if c.Base.Ascension >= 1 {
		// Small decay ticks must accumulate across the loop, not lose each
		// remainder to integer division before reaching ten removed power.
		c.tacticPowerRemoved += removed
		stacks := int((c.tacticPowerRemoved + 1e-9) / 10)
		if c.Core.F >= c.tacticExpiry {
			c.tacticStacks = 0
		}
		c.tacticStacks = min(10, c.tacticStacks+stacks)
		if stacks > 0 {
			// A shared 60s refresh timer prevents indefinite retention.
			// Independent-vs-shared layer refresh remains unmeasured.
			c.tacticExpiry = c.Core.F + 60*60
			expires := c.tacticExpiry
			c.QueueCharTask(func() {
				if c.tacticExpiry == expires {
					c.tacticStacks = 0
				}
			}, 60*60)
		}
		c.tacticPowerRemoved = max(0, c.tacticPowerRemoved-float64(stacks)*10)
	}
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 3
	c.NormalCon = 3
	c.BurstCon = 5
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.InitStellarRadiance()
	c.initAscensions()
	c.initConstellations()
	return nil
}
