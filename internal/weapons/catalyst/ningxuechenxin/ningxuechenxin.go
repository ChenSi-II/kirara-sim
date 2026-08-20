package ningxuechenxin

import (
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct {
	Index  int
	c      *core.Core
	char   *character.CharWrapper
	refine int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	cryoCount := 0
	electroCount := 0
	for _, ch := range w.c.Player.Chars() {
		switch ch.Base.Element {
		case attributes.Cryo:
			cryoCount++
		case attributes.Electro:
			electroCount++
		}
	}
	starCount := min(cryoCount+electroCount, 4)
	r := float64(w.refine)
	m := make([]float64, attributes.EndStatType)
	w.char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("ningxue-chenxin", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() []float64 {
			m[attributes.EM] = 0
			m[attributes.ATKP] = 0
			if starDomainActive(w.c) {
				m[attributes.EM] = (15 + 5*r) * float64(starCount)
			} else {
				m[attributes.EM] = (18 + 6*r) * float64(cryoCount)
				m[attributes.ATKP] = (0.036 + 0.012*r) * float64(electroCount)
			}
			return m
		},
	})
	w.char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("ningxue-chenxin-star-reaction", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !starDomainActive(w.c) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				return (0.045 + 0.015*r) * float64(starCount)
			default:
				return 0
			}
		},
	})
	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	return &Weapon{c: c, char: char, refine: p.Refine}, nil
}

func starDomainActive(c *core.Core) bool {
	return c.StarReactions.SuperconductActive || c.StarReactions.DiffusionActive
}
