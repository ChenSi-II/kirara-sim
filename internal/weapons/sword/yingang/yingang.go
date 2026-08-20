package yingang

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

var stackKeys = []string{"yingang-stack-0", "yingang-stack-1"}

type Weapon struct{ Index int }

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{}
	em := 39 + 13*float64(p.Refine)
	m := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("yingang", -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			stacks := 0
			for _, key := range stackKeys {
				if char.StatusIsActive(key) {
					stacks++
				}
			}
			m[attributes.EM] = em * float64(stacks)
			return m
		},
	})
	c.Events.Subscribe(event.OnSkill, func(...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		index := 0
		for i, key := range stackKeys {
			if char.StatusExpiry(key) < char.StatusExpiry(stackKeys[index]) {
				index = i
			}
		}
		char.AddStatus(stackKeys[index], 12*60, true)
	}, fmt.Sprintf("yingang-skill-%v", char.Base.Key.String()))
	return w, nil
}
