package vesna

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

const (
	spiritbladeArmedKey = "vesna-spiritblade-armed"
	radianceKey         = "vesna-radiance-star-diffusion"
	composureKey        = "vesna-composure"
	stepReadyKey        = "vesna-spiritblade-step-ready"
)

type char struct {
	*tmpl.Character
	armedSrc          int
	magic             int
	specialStage      int
	danceCount        int
	composure         int
	composureExpiries []int
	freeDance         bool
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 6
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev != c.Index() {
			return
		}
		c.endSpiritbladeArmament()
		c.composure = 0
		c.composureExpiries = nil
		c.DeleteStatus(composureKey)
	}, "vesna-swap")
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(stepReadyKey) {
		return true, action.NoFailure
	}
	if a == action.ActionSkill && c.StatusIsActive(spiritbladeArmedKey) &&
		(c.magic > 0 || (c.specialStage == 2 && c.freeDance)) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) endSpiritbladeArmament() {
	c.DeleteStatus(spiritbladeArmedKey)
	c.magic = 0
	c.specialStage = 0
	c.danceCount = 0
	c.freeDance = false
}

func (c *char) addMagic(amount int) {
	if !c.StatusIsActive(spiritbladeArmedKey) {
		return
	}
	c.magic = min(c.magic+amount, 5)
}

func (c *char) addComposure() {
	if c.Base.Ascension < 1 {
		return
	}
	if c.activeComposure() == 6 {
		c.composureExpiries = c.composureExpiries[1:]
	}
	c.composureExpiries = append(c.composureExpiries, c.Core.F+20*60)
	c.composure = len(c.composureExpiries)
	c.AddStatus(composureKey, 20*60, true)
}

func (c *char) activeComposure() int {
	active := c.composureExpiries[:0]
	for _, expiry := range c.composureExpiries {
		if expiry > c.Core.F {
			active = append(active, expiry)
		}
	}
	c.composureExpiries = active
	c.composure = len(active)
	return c.composure
}

func (c *char) spiritbladeBonus() float64 {
	if c.Base.Ascension < 1 {
		return 1
	}
	return 1 + 0.1*float64(c.activeComposure())
}
