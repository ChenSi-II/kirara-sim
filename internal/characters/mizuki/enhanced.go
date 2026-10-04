package mizuki

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) enhancedInit() {
	if !c.Enhanced {
		return
	}
	c.InitDiffusionRadiance(false)
	hook := func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != c.Index() || !c.StatusIsActive(dreamDrifterStateKey) || c.StatusIsActive("mizuki-cloud-empower-icd") {
			return
		}
		c.empoweredCloud = true
		c.AddStatus("mizuki-cloud-empower-icd", 150, false)
	}
	for _, ev := range []event.Event{event.OnSwirlCryo, event.OnSwirlHydro, event.OnSwirlElectro, event.OnSwirlPyro, event.OnStarDiffusion} {
		c.Core.Events.Subscribe(ev, hook, "mizuki-cloud-empower")
	}
	for _, ch := range c.Core.Player.Chars() {
		buff := make([]float64, attributes.EndStatType)
		ch.AddStatMod(character.StatMod{Base: modifier.NewBase("mizuki-enhanced-em", -1), AffectedStat: attributes.EM, Extra: true, Amount: func() []float64 {
			if !c.StatusIsActive(dreamDrifterStateKey) {
				return nil
			}
			buff[attributes.EM] = .1 * c.NonExtraStat(attributes.EM)
			return buff
		}})
	}
	if c.Base.Cons >= 2 {
		c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			e := args[0].(*enemy.Enemy)
			active := c.StatusIsActive(dreamDrifterStateKey) && e.IsWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 12))
			for _, ele := range []attributes.Element{attributes.Pyro, attributes.Hydro, attributes.Cryo, attributes.Electro, attributes.Anemo} {
				key := "mizuki-c2-res-" + ele.String()
				if active {
					e.AddResistMod(info.ResistMod{Base: modifier.NewBase(key, 1), Ele: ele, Value: -.2})
				} else {
					e.DeleteResistMod(key)
				}
			}
		}, "mizuki-c2-res")
	}
	if c.Base.Cons < 6 {
		return
	}
	buff := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{Base: modifier.NewBase("mizuki-c6-self", -1), Extra: true, Amount: func() []float64 {
		em := max(0, c.NonExtraStat(attributes.EM)-500)
		buff[attributes.CR] = min(.2, em*.0004)
		buff[attributes.CD] = min(.8, em*.0016)
		return buff
	}})
	apply := func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if c.StatusIsActive(dreamDrifterStateKey) && tmpl.IsStarDiffusion(a.Info.AttackTag) {
			a.Snapshot.Stats[attributes.CR] += .1
			a.Snapshot.Stats[attributes.CD] += .2
		}
	}
	c.Core.Events.Subscribe(event.OnStarReactionAttack, apply, "mizuki-c6-star")
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.IsDirectStarDamage() {
			apply(args...)
		}
	}, "mizuki-c6-star-direct")
}

func (c *char) empoweredCloudCB(empowered bool) info.AttackCBFunc {
	done := false
	return func(a info.AttackCB) {
		if done || !empowered || a.Target.Type() != info.TargettableEnemy || !c.DiffusionRadiance() {
			return
		}
		done = true
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Dreamdrifter (Stellar)", AttackTag: attacks.AttackTagReactionStarDiffusionAnemo, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Anemo, UseEM: true, Mult: 10}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(a.Target, nil, cloudExplosionRadius), 0, 1)
	}
}

func (c *char) healOther(picker int) {
	var lowest *character.CharWrapper
	for _, ch := range c.Core.Player.Chars() {
		if ch.Index() != picker && ch.CurrentHP() > 0 && (lowest == nil || ch.CurrentHP() < lowest.CurrentHP()) {
			lowest = ch
		}
	}
	if lowest != nil {
		c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: lowest.Index(), Message: "Snack (Enhanced C4)", Src: 2.66 * c.Stat(attributes.EM), Bonus: c.Stat(attributes.Heal)})
	}
}
