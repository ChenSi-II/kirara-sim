package prune

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
	p := testhelper.DefaultProfile(keys.Prune, testhelper.TestWeaponKey)
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

func TestImageBurstAndDelayedPassive(t *testing.T) {
	c, ch, hits := setupTiming(t, 6)
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.Frames(action.ActionSwap) != 75 {
		t.Fatal("Q must swap at 75f")
	}
	tickTo(c, 175)
	// A bell-triggered swirl must produce its hammer 48f later; a C4
	// ricochet follows the hammer by 64f, including while Prune is off field.
	c.Events.Emit(event.OnSwirlCryo, c.Combat.PrimaryTarget(), &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Abil: "Witchlure Bell"}})
	c.Player.SetActive(1)
	tickTo(c, 1030)
	assertFrames(t, hits["The Bell Tolls!"], []int{44})
	assertFrames(t, hits["Witchlure Bell"], []int{175, 292, 409, 526, 646, 764, 877, 994})
	assertFrames(t, hits["Banehunter Oathhammer"], []int{223})
	assertFrames(t, hits["Banehunter Oathhammer Bounce"], []int{287})
}

func TestImageConvertedSkill(t *testing.T) {
	c, ch, hits := setupTiming(t, 4)
	ch.converted = attributes.Cryo
	ch.AddStatus(conversionKey, 360, false)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 110)
	assertFrames(t, hits["Hexhunter Chime"], []int{41})
	assertFrames(t, hits["Witch-tribution Ricochet"], []int{105})
}
