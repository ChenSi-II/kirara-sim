package keyoftranscendence

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func init() {
	core.RegisterWeaponFunc(keys.KeyOfTranscendence, NewWeapon)
}

type Weapon struct {
	Index  int
	stacks int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

const (
	transcendenceKey    = "key-of-transcendence-stacks"
	transcendenceICDKey = "key-of-transcendence-stack-icd"
)

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	r := float64(p.Refine)

	onFieldATK := make([]float64, attributes.EndStatType)
	onFieldATK[attributes.ATKP] = 0.12 + 0.04*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("key-of-transcendence-atk", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			if c.Player.Active() != char.Index() {
				return nil
			}
			return onFieldATK
		},
	})

	stackBonus := 0.15 + 0.05*r
	char.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("key-of-transcendence-star-superconduct", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag != attacks.AttackTagReactionStarSuperconduct || !char.StatusIsActive(transcendenceKey) {
				return 0
			}
			return stackBonus * float64(w.stacks)
		},
	})

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() || c.Player.Active() != char.Index() {
			return
		}
		if atk.Info.AttackTag != attacks.AttackTagExtra || char.StatusIsActive(transcendenceICDKey) {
			return
		}
		if !char.StatusIsActive(transcendenceKey) {
			w.stacks = 0
		}
		if w.stacks < 3 {
			w.stacks++
		}
		char.AddStatus(transcendenceICDKey, int(0.2*60), true)
		char.AddStatus(transcendenceKey, 5*60, true)
	}, fmt.Sprintf("key-of-transcendence-charge-%v", char.Base.Key.String()))

	return w, nil
}
