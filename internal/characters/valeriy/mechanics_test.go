package valeriy

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"math"
	"testing"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

type timingAlly struct{ *tmpl.Character }

func (c *timingAlly) Init() error { return nil }

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func setupTiming(t *testing.T, cons int) (*core.Core, *char, map[string][]int) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
	c.Combat.AddEnemy(target)
	p := testhelper.DefaultProfile(keys.Valeriy, testhelper.TestWeaponKey)
	p.Base.Cons = cons
	p.Base.Ascension = 6
	if _, err := c.AddChar(p); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)); err != nil {
		t.Fatal(err)
	}
	ally := c.Player.ByIndex(1)
	ally.Character = &timingAlly{tmpl.NewWithWrapper(c, ally)}
	ally.SkillCon, ally.BurstCon = 3, 5
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = target.Key()
	hits := make(map[string][]int)
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == 0 {
			hits[atk.Info.Abil] = append(hits[atk.Info.Abil], c.F)
		}
	}, "record-timing")
	return c, c.Player.ByIndex(0).Character.(*char), hits
}

func advanceTo(t *testing.T, c *core.Core, frame int) {
	t.Helper()
	for c.F < frame {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSkillPotentialShieldAndC1(t *testing.T) {
	for _, cons := range []int{0, 1, 4} {
		c, ch, _ := setupTiming(t, cons)
		ch.Skill(nil)
		want := 40.0
		if cons >= 1 {
			want = 100
		}
		if cons >= 4 {
			want = 130
		}
		if ch.potential != want {
			t.Fatalf("C%d potential %v want %v", cons, ch.potential, want)
		}
		ch.ChargeAttack(nil)
		advanceTo(t, c, 31)
		if ch.orders != 15 || ch.potential != 0 {
			t.Fatal("special charge must spend all potential and grant orders")
		}
		advanceTo(t, c, 601)
		ch.Skill(nil)
		want = 40
		if cons >= 4 {
			want = 52
		}
		if ch.potential != want {
			t.Fatalf("first-skill bonus repeated: %v", ch.potential)
		}
	}
}
func TestBurstGainOnlyActiveOtherAndCap(t *testing.T) {
	c, ch, _ := setupTiming(t, 4)
	ch.Burst(nil)
	hit := func(actor int, ele attributes.Element) {
		c.Events.Emit(event.OnEnemyHit, c.Combat.PrimaryTarget(), &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor, Element: ele}})
	}
	hit(0, attributes.Electro)
	if ch.potential != 0 {
		t.Fatal("own hit generated potential")
	}
	c.Player.SetActive(1)
	hit(0, attributes.Electro)
	hit(1, attributes.Cryo)
	if ch.potential != 0 {
		t.Fatal("off-field/non-electro hit generated potential")
	}
	for i := 0; i < 30; i++ {
		hit(1, attributes.Electro)
	}
	if ch.potential != 104 || ch.burstGained != 104 {
		t.Fatalf("C4 cap: %v/%v", ch.potential, ch.burstGained)
	}
}
func TestOrdersFilterAndPerEnemyConsumption(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	ch.Skill(nil)
	ch.ChargeAttack(nil)
	advanceTo(t, c, 31)
	c.Player.SetActive(1)
	hit := func(actor int, ele attributes.Element, tag attacks.AttackTag) *info.AttackEvent {
		a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor, Element: ele, AttackTag: tag}}
		c.Events.Emit(event.OnEnemyHit, c.Combat.PrimaryTarget(), a)
		return a
	}
	hit(0, attributes.Electro, attacks.AttackTagNormal)
	hit(1, attributes.Cryo, attacks.AttackTagNormal)
	if ch.orders != 15 {
		t.Fatal("ineligible hit consumed orders")
	}
	for i := 0; i < 2; i++ {
		a := hit(1, attributes.Electro, attacks.AttackTagNormal)
		if a.Info.FlatDmg <= 0 {
			t.Fatal("missing flat bonus")
		}
	}
	if ch.orders != 13 {
		t.Fatal("each target must consume one order")
	}
	c.StarReactions.SuperconductActive = true
	a := hit(1, attributes.Electro, attacks.AttackTagNormal)
	if a.Info.FlatDmg != 0 || ch.orders != 13 {
		t.Fatal("ordinary damage consumed stellar orders")
	}
	a = hit(1, attributes.Electro, attacks.AttackTagReactionStarSuperconduct)
	if a.Info.FlatDmg <= 0 || ch.orders != 12 {
		t.Fatal("stellar bonus missing")
	}
}
func TestShieldOverflowCap(t *testing.T) {
	c, ch, _ := setupTiming(t, 4)
	ch.Skill(nil)
	s := c.Player.Shields.Get(shield.ValeriySkill)
	hp := s.CurrentHP()
	ch.gainPotential(13)
	ch.gainPotential(13)
	ch.gainPotential(13)
	want := hp + 20*.08*ch.TotalAtk()
	if math.Abs(s.CurrentHP()-want) > 1e-6 {
		t.Fatalf("shield HP %v want %v", s.CurrentHP(), want)
	}
}
