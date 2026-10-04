package klee

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	c1Chance   float64
	sparks     int
	burstSrc   int
	burstEnded bool
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	hex, ok := p.Params["hexerei"]
	c.IsHexerei = !ok || hex != 0

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	c.SetNumCharges(action.ActionSkill, 2)

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.onExitField()
	c.hexInit()
	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge {
		if c.StatusIsActive(a1SparkKey) {
			return 0
		}
		return 50
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) ResetNormalCounter() {
	if c.hexActive() && c.Core.Status.Duration("kleeq") > 0 {
		return
	}
	c.Character.ResetNormalCounter()
}
