package illuga

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	nightingaleStacks int
	consumedStacks    int
	constructStacks   int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 4
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	return nil
}
