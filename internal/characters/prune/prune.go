package prune

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	converted attributes.Element
	bellSrc   int
	c2Stacks  int
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(conversionKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	hex, ok := p.Params["hexerei"]
	c.IsHexerei = !ok || hex != 0
	c.EnergyMax = 70
	c.NormalHitNum = 3
	c.SkillCon = 5
	c.BurstCon = 3
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	c.initReactionBuffs()
	return nil
}
