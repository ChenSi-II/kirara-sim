package xinzhi

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	skillWindowKey = "xinzhi-skill-window"
	stackKey       = "xinzhi-stacks"
	stackICDKey    = "xinzhi-stack-icd"
)

type Weapon struct {
	Index  int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	m := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xinzhi", -1),
		AffectedStat: attributes.NoStat,
		Amount: func() []float64 {
			m[attributes.ATKP] = 0
			m[attributes.EM] = 0
			if !char.StatusIsActive(stackKey) {
				w.stacks = 0
				return m
			}
			if starDomainActive(c) {
				m[attributes.ATKP] = (0.045 + 0.015*r) * float64(w.stacks)
			} else {
				m[attributes.ATKP] = (0.03 + 0.01*r) * float64(w.stacks)
				m[attributes.EM] = (15 + 5*r) * float64(w.stacks)
			}
			return m
		},
	})
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("xinzhi-star-reaction", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !starDomainActive(c) || !char.StatusIsActive(stackKey) {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
				return (0.06 + 0.02*r) * float64(w.stacks)
			default:
				return 0
			}
		},
	})

	c.Events.Subscribe(event.OnSkill, func(...any) {
		if c.Player.Active() == char.Index() {
			char.AddStatus(skillWindowKey, 12*60, true)
		}
	}, fmt.Sprintf("xinzhi-skill-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || !char.StatusIsActive(skillWindowKey) || char.StatusIsActive(stackICDKey) {
			return
		}
		if !char.StatusIsActive(stackKey) {
			w.stacks = 0
		}
		w.stacks = min(w.stacks+1, 3)
		char.AddStatus(stackKey, 6*60, true)
		char.AddStatus(stackICDKey, 60, true)
	}, fmt.Sprintf("xinzhi-hit-%v", char.Base.Key.String()))

	return w, nil
}

func starDomainActive(c *core.Core) bool {
	return c.StarReactions.SuperconductActive || c.StarReactions.DiffusionActive
}
