package lohen

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	joy, will, etchedUses int
	evilsbane             bool
	c6ExtendReady         bool
}

const c3MomentumKey = "lohen-c3-momentum"

func (c *char) skillLevel() int {
	lvl := c.TalentLvlSkill()
	if c.StatusIsActive(c3MomentumKey) {
		lvl = min(lvl+1, len(skillParam[0])-1)
	}
	return lvl
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(masterstrokeKey) && c.joy >= 100 && c.etchedUses < c.maxEtchedUses() {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 5
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
