package alyosha

import (
	"reflect"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

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
	p := testhelper.DefaultProfile(keys.Alyosha, testhelper.TestWeaponKey)
	p.Base.Cons = cons
	p.Base.Ascension = 6
	if _, err := c.AddChar(p); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = target.Key()
	hits := make(map[string][]int)
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == 0 {
			hits[atk.Info.Abil] = append(hits[atk.Info.Abil], c.F)
		}
	}, "test-timing")
	return c, c.Player.ByIndex(0).Character.(*char), hits
}

func tickTo(c *core.Core, frame int) {
	for c.F < frame {
		c.F++
		c.Tick()
	}
}

func assertFrames(t *testing.T, got, want []int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hit frames = %v; want %v", got, want)
	}
}

func TestImageSkillAndComboCancels(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	e, err := ch.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.AnimationLength != 38 || e.Frames(action.ActionBurst) != 30 || e.Frames(action.ActionSwap) != 38 {
		t.Fatalf("unexpected E animation/cancel: %+v", e)
	}
	tickTo(c, 30)
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.AnimationLength != 57 || q.Frames(action.ActionSkill) != 49 {
		t.Fatalf("unexpected Q animation/cancel: %+v", q)
	}
	assertFrames(t, hits["Thunderbolt Strike"], []int{26})
	if e.Frames(action.ActionBurst)+q.AnimationLength != 87 || q.Frames(action.ActionSkill)+e.AnimationLength != 87 {
		t.Fatal("EQ and QE should each take the image's 1.450 seconds")
	}
}

func TestImageBurstProjectilesSurviveSwap(t *testing.T) {
	for _, cons := range []int{0, 2} {
		t.Run(string(rune('0'+cons)), func(t *testing.T) {
			c, ch, hits := setupTiming(t, cons)
			if _, err := ch.Burst(nil); err != nil {
				t.Fatal(err)
			}
			tickTo(c, 57)
			c.Player.SetActive(1)
			tickTo(c, 1300)
			field := []int{82, 199, 317, 434, 552, 668, 785}
			tugarin := []int{128, 245, 361, 479, 596, 715, 829}
			if cons >= 2 {
				field = append(field, 903, 1023, 1137)
				tugarin = append(tugarin, 946, 1068, 1181)
			}
			assertFrames(t, hits["Fulgurite Hunting Field"], field)
			assertFrames(t, hits["Tugarin"], tugarin)
		})
	}
}
