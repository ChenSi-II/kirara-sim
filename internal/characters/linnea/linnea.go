package linnea

import (
	"fmt"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type lumiForm int

const (
	lumiSuper lumiForm = iota
	lumiStandard
)

type char struct {
	*tmpl.Character
	lumiSrc            int
	lumiForm           lumiForm
	lumiFeed           int
	fieldCatalogStacks int
	fieldCatalogSrc    int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w)}
	c.EnergyMax = 60
	c.NormalHitNum = 3
	c.SkillCon = 3
	c.BurstCon = 5
	c.Moonsign = 1
	w.Character = c
	return nil
}

func (c *char) Init() error {
	c.initAscensions()
	c.initConstellations()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lumi":
		return c.StatusIsActive(lumiKey), nil
	case "lumi-form":
		if !c.StatusIsActive(lumiKey) {
			return "absent", nil
		}
		return c.lumiForm.String(), nil
	case "field-catalog":
		return c.fieldCatalogStacks, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	if a == action.ActionCharge {
		return nil
	}
	return c.Character.NextQueueItemIsValid(k, a, p)
}

func (f lumiForm) String() string {
	switch f {
	case lumiSuper:
		return "super"
	case lumiStandard:
		return "standard"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}
