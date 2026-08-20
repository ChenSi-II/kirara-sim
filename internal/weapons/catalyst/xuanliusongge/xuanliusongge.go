package xuanliusongge

import (
	"fmt"
	"math"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const anthemKey = "xuanliu-songge-anthem"

var songStackKeys = []string{"xuanliu-songge-stack-0", "xuanliu-songge-stack-1", "xuanliu-songge-stack-2"}

type Weapon struct {
	Index int
	char  *character.CharWrapper
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func (w *Weapon) stacks() int {
	count := 0
	for _, key := range songStackKeys {
		if w.char.StatusIsActive(key) {
			count++
		}
	}
	return count
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{char: char}
	r := float64(p.Refine)

	heal := make([]float64, attributes.EndStatType)
	heal[attributes.Heal] = 0.06 + 0.02*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xuanliu-songge-heal", -1),
		AffectedStat: attributes.Heal,
		Amount:       func() []float64 { return heal },
	})

	hpPerStack := 0.04 + 0.01*r
	hp := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xuanliu-songge-hp", -1),
		AffectedStat: attributes.HPP,
		Amount: func() []float64 {
			mult := 1.0
			if char.StatusIsActive(anthemKey) {
				mult = 1.75
			}
			hp[attributes.HPP] = hpPerStack * float64(w.stacks()) * mult
			return hp
		},
	})

	atkPerThousand := 0.00175 + 0.00075*r
	atkCap := 0.035 + 0.015*r
	for _, ch := range c.Player.Chars() {
		target := ch
		m := make([]float64, attributes.EndStatType)
		target.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(fmt.Sprintf("xuanliu-songge-atk-%v", char.Base.Key.String()), -1),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				m[attributes.ATKP] = 0
				stacks := w.stacks()
				if stacks == 0 || target.Index() != c.Player.Active() {
					return m
				}
				excessThousands := math.Floor(max(char.MaxHP()-40000, 0) / 1000)
				perStack := min(excessThousands*atkPerThousand, atkCap)
				mult := 1.0
				if char.StatusIsActive(anthemKey) {
					mult = 1.75
				}
				m[attributes.ATKP] = perStack * float64(stacks) * mult
				return m
			},
		})
	}

	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		source := args[0].(*info.HealInfo)
		amount := args[2].(float64)
		if source.Caller != char.Index() || amount <= 0 {
			return
		}
		index := 0
		for i, key := range songStackKeys {
			if char.StatusExpiry(key) < char.StatusExpiry(songStackKeys[index]) {
				index = i
			}
		}
		char.AddStatus(songStackKeys[index], 8*60, true)
	}, fmt.Sprintf("xuanliu-songge-heal-%v", char.Base.Key.String()))
	anthem := func(...any) { char.AddStatus(anthemKey, 5*60, true) }
	c.Events.Subscribe(event.OnFrozen, anthem, fmt.Sprintf("xuanliu-songge-frozen-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, anthem, fmt.Sprintf("xuanliu-songge-star-diffusion-%v", char.Base.Key.String()))

	return w, nil
}
