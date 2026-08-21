package nefer

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	phantasmUses   int
	seeds, veils   int
	a2DewRemainder float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 4
	c.SkillCon = 3
	c.BurstCon = 5
	c.Moonsign = 1
	c.SetNumCharges(action.ActionSkill, 2)
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	return nil
}
