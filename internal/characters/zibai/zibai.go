package zibai

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	phase         float64
	strides       int
	phaseSrc      int
	scattermoon   bool
	c1FirstStride bool
	c6Elevation   float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 4
	c.SkillCon = 3
	c.BurstCon = 5
	c.Moonsign = 1
	w.Character = c
	return nil
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(lunarPhaseKey) {
		if c.phase < 70 {
			return false, action.SkillCD
		}
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	return nil
}
