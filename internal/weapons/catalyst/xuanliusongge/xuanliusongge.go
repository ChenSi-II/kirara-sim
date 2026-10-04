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

const (
	anthemKey = "xuanliu-songge-anthem"
	songKey   = "xuanliu-songge-stacks"
)

type Weapon struct {
	Index int
	char  *character.CharWrapper
	count int
}

func (w *Weapon) SetIndex(idx int) { w.Index = idx }
func (w *Weapon) Init() error      { return nil }

func (w *Weapon) stacks() int {
	if w.char.StatusDuration(songKey) == 0 {
		w.count = 0
	}
	return w.count
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{char: char}
	r := float64(p.Refine)

	heal := make([]float64, attributes.EndStatType)
	heal[attributes.Heal] = 0.03 + 0.01*r
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xuanliu-songge-heal", -1),
		AffectedStat: attributes.Heal,
		Amount:       func() []float64 { return heal },
	})

	hpPerStack := 0.03 + 0.01*r
	hp := make([]float64, attributes.EndStatType)
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xuanliu-songge-hp", -1),
		AffectedStat: attributes.HPP,
		Amount: func() []float64 {
			mult := 1.0
			if char.StatusDuration(anthemKey) > 0 {
				mult = 1.75
			}
			hp[attributes.HPP] = hpPerStack * float64(w.stacks()) * mult
			return hp
		},
	})

	atkPerThousand := 0.003 + 0.001*r
	atkCap := 0.06 + 0.02*r
	c.Events.Subscribe(event.OnInitialize, func(...any) {
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
					if char.StatusDuration(anthemKey) > 0 {
						mult = 1.75
					}
					m[attributes.ATKP] = perStack * float64(stacks) * mult
					return m
				},
			})
		}
	}, fmt.Sprintf("xuanliu-songge-team-%v", char.Index()))

	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		source := args[0].(*info.HealInfo)
		if source.Caller != char.Index() {
			return
		}
		w.count = min(w.stacks()+1, 3)
		char.AddStatus(songKey, 10*60, true)
	}, fmt.Sprintf("xuanliu-songge-heal-%v", char.Base.Key.String()))
	anthem := func(...any) { char.AddStatus(anthemKey, 5*60, true) }
	c.Events.Subscribe(event.OnFrozen, anthem, fmt.Sprintf("xuanliu-songge-frozen-%v", char.Base.Key.String()))
	c.Events.Subscribe(event.OnStarDiffusion, anthem, fmt.Sprintf("xuanliu-songge-star-diffusion-%v", char.Base.Key.String()))

	return w, nil
}
