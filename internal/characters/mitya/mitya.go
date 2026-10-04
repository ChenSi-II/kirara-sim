package mitya

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	beacons                  []int
	critStacks               []int
	reactorUntil, reactorSrc int
	overloaded               bool
	consumed                 int
	generationUntil          int
	generationRunning        bool
	channelSrc               int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 3
	c.SkillCon = 3
	c.NormalCon = 5
	w.Character = c
	return nil
}
func (c *char) Init() error         { c.initEffects(); return nil }
func (c *char) reactorActive() bool { return c.Core.F < c.reactorUntil }
func (c *char) beaconCount() int {
	live := c.beacons[:0]
	for _, expires := range c.beacons {
		if expires > c.Core.F {
			live = append(live, expires)
		}
	}
	c.beacons = live
	return len(live)
}
func (c *char) beaconCap() int {
	if c.Base.Cons >= 1 {
		return 5
	}
	return 4
}
func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge && c.reactorActive() && c.overloaded && c.beaconCount() > 0 {
		return 0
	}
	return c.Character.ActionStam(a, p)
}
func (c *char) addBeacon() {
	dur := 360
	if c.Base.Ascension >= 4 {
		dur = 600
	}
	count := c.beaconCount()
	if count >= c.beaconCap() {
		for i := range c.beacons {
			c.beacons[i] = c.Core.F + dur
		}
		if c.reactorActive() && !c.overloaded {
			c.starHit("Steady Core Beacon", skillParams[c.TalentLvlSkill()][3], true, 0, true)
			c.onConsumed(1)
		}
	} else {
		c.beacons = append(c.beacons, c.Core.F+dur)
	}
	if c.Base.Ascension >= 4 {
		live := c.critStacks[:0]
		for _, e := range c.critStacks {
			if e > c.Core.F {
				live = append(live, e)
			}
		}
		c.critStacks = live
		if len(c.critStacks) >= 4 {
			c.critStacks = c.critStacks[1:]
		}
		c.critStacks = append(c.critStacks, c.Core.F+900)
	}
	if c.Base.Cons >= 1 && !c.StatusIsActive("mitya-c1-energy") {
		c.AddEnergy("mitya-c1", 10)
		c.AddStatus("mitya-c1-energy", 480, false)
	}
	c.syncDomain()
	c.QueueCharTask(c.syncDomain, dur)
}
func (c *char) syncDomain() {
	// Beacons replace the prism domain. Nearby geometry is assumed, as elsewhere
	// in the simulator; the domain ends when the last beacon expires/is spent.
	active := c.beaconCount() > 0
	c.Core.StarReactions.SuperconductActive = active
	c.Core.StarReactions.SuperconductCoefficient = 1
	c.refreshShred()
}
func (c *char) onConsumed(n int) {
	c.consumed += n
	if c.Base.Cons >= 4 && c.reactorUntil > 0 && c.Core.F <= c.reactorUntil {
		for c.consumed >= 5 {
			c.consumed -= 5
			c.starHit("Core Fission C4", 4, true, 0, true)
		}
	}
	if c.Base.Cons >= 6 && c.overloaded && c.reactorActive() {
		for i := 0; i < n; i++ {
			if c.Core.Rand.Float64() < .30 {
				c.addBeacon()
			}
		}
	}
}
func (c *char) spendBeacon() bool {
	if c.beaconCount() == 0 {
		return false
	}
	c.beacons = c.beacons[1:]
	c.onConsumed(1)
	c.syncDomain()
	return true
}
func (c *char) finishCore() {
	if !c.overloaded {
		n := c.beaconCount()
		c.beacons = nil
		if n > 0 {
			c.starHit("Steady Core Detonation", skillParams[c.TalentLvlSkill()][4]*float64(n), false, 0, true)
			c.onConsumed(n)
		}
	}
	c.reactorUntil = 0
	c.syncDomain()
}
