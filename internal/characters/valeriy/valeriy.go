package valeriy

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
)

type char struct {
	*tmpl.Character
	potential, burstGained, shieldOverflow float64
	firstSkill                             bool
	burstUntil, burstSrc                   int
	orders, ordersUntil                    int
	spent                                  float64
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := &char{Character: tmpl.NewWithWrapper(s, w), firstSkill: true}
	c.EnergyMax = 60
	c.NormalHitNum = 3
	c.NormalCon = 3
	c.SkillCon = 5
	w.Character = c
	return nil
}
func (c *char) Init() error { c.initEffects(); return nil }
func (c *char) gainPotential(n float64) {
	cap := 100.0
	if c.Base.Cons >= 4 {
		cap = 130
	}
	// C2 requires the gauge to already be full when the gain is triggered.
	if c.Base.Cons >= 2 && c.potential >= cap {
		if s, ok := c.Core.Player.Shields.Get(shield.ValeriySkill).(*shield.Tmpl); ok {
			overflow := min(n, 20-c.shieldOverflow)
			s.HP += overflow * .08 * c.TotalAtk()
			c.shieldOverflow += overflow
		}
	}
	c.potential = min(cap, c.potential+n)
}
func (c *char) gainFactor() float64 {
	if c.Base.Cons >= 4 {
		return 1.3
	}
	return 1
}
func (c *char) createShield() {
	p := skillParams[c.TalentLvlSkill()]
	c.shieldOverflow = 0
	c.Core.Player.Shields.Add(&shield.Tmpl{ActorIndex: c.Index(), Target: -1, Name: "Thunderstock Shield", Src: c.Core.F, ShieldType: shield.ValeriySkill, Ele: attributes.Electro, HP: p[1]*c.TotalAtk() + p[2], Expires: c.Core.F + int(p[3]*60)})
}
