package sandrone

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	skillStart         int
	resolutionSrc      int
	resolutionPower    float64
	resolutionRays     int
	tacticStacks       int
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
		c.tacticStacks = min(10, c.tacticStacks+stacks)
		c.tacticPowerRemoved = max(0, c.tacticPowerRemoved-float64(stacks)*10)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionCharge && c.powerOverdrive {
		return false, action.SkillCD
	}
	return c.Character.ActionReady(a, p)
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
	c.initAscensions()
	c.initConstellations()
	return nil
}
