package roufengyouxian

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
	pointICDKey = "roufeng-youxian-point-icd"
	lockoutKey  = "roufeng-youxian-lockout"
	buffKey     = "roufeng-youxian-star-reaction"
)

type Weapon struct {
	Index  int
	points int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)
	er := make([]float64, attributes.EndStatType)
	er[attributes.ER] = 0.15 + 0.05*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("roufeng-youxian-er", -1),
		AffectedStat: attributes.ER,
		Amount:       func() []float64 { return er },
	})

	bonus := 0.18 + 0.06*r
	for _, ch := range c.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(fmt.Sprintf("roufeng-youxian-%v", char.Base.Key.String()), -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !char.StatusIsActive(buffKey) {
					return 0
				}
				switch ai.AttackTag {
				case attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo:
					return bonus
				default:
					return 0
				}
			},
		})
	}

	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || char.StatusIsActive(lockoutKey) || char.StatusIsActive(pointICDKey) {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagElementalArt, attacks.AttackTagElementalArtHold, attacks.AttackTagElementalBurst:
		default:
			return
		}
		char.AddStatus(pointICDKey, 2, true)
		w.points++
		if w.points < 3 {
			return
		}
		w.points = 0
		char.AddStatus(buffKey, 12*60, true)
		char.AddStatus(lockoutKey, 12*60, true)
	}, fmt.Sprintf("roufeng-youxian-hit-%v", char.Base.Key.String()))

	return w, nil
}
