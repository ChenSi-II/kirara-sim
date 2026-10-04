package lohen

import (
	"fmt"
	"reflect"
	"testing"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"

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
	p := testhelper.DefaultProfile(keys.Lohen, testhelper.TestWeaponKey)
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

func checkFrames(t *testing.T, got, want []int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hit frames = %v, want %v", got, want)
	}
}

func TestMasterstrokeNormalRecording(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	e, err := ch.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, e.Frames(action.ActionAttack))
	for range 5 {
		a, err := ch.Attack(nil)
		if err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, c.F+a.Frames(action.ActionAttack))
	}
	if c.F != 207 {
		t.Fatalf("E -> N5 end = %d, want 207", c.F)
	}
	want := [][]int{{33}, {55}, {81, 93, 105}, {119}, {155, 177}}
	for stage, times := range want {
		for hit, frame := range times {
			checkFrames(t, hits[fmt.Sprintf("Normal %d-%d", stage+1, hit+1)], []int{frame})
		}
	}
}

func TestMasterstrokeN1ChargeRecording(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	e, err := ch.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, e.Frames(action.ActionAttack))
	a, err := ch.Attack(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, c.F+a.Frames(action.ActionCharge))
	z, err := ch.ChargeAttack(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, c.F+z.Frames(action.ActionAttack))
	checkFrames(t, hits["Normal 1-1"], []int{33})
	checkFrames(t, hits["Charged Attack"], []int{58, 66})
	if c.F != 69 {
		t.Fatalf("N1C duration = %d, want 48", c.F-21)
	}
}

func TestBurstHitsContinueAfterSwapWindow(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.Frames(action.ActionSwap) != 134 {
		t.Fatal("wrong burst swap frame")
	}
	advanceTo(t, c, q.Frames(action.ActionSwap))
	c.Player.SetActive(1)
	advanceTo(t, c, 180)
	for hit, frame := range []int{106, 131, 137, 140, 147, 171} {
		checkFrames(t, hits[fmt.Sprintf("Manifest Judgment %d", hit+1)], []int{frame})
	}
}

func TestEtchedRecording(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	a, err := ch.etchedIntoBoneAndSoul()
	if err != nil {
		t.Fatal(err)
	}
	if a.AnimationLength != 70 {
		t.Fatal("wrong special E length")
	}
	advanceTo(t, c, 70)
	for hit, frame := range []int{27, 34, 45, 50} {
		checkFrames(t, hits[fmt.Sprintf("Etched Into Bone and Soul %d", hit+1)], []int{frame})
	}
}

func TestEvilsbaneFollowsNormalHit(t *testing.T) {
	c, ch, hits := setupTiming(t, 2)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	ch.evilsbane = true
	advanceTo(t, c, 21)
	if _, err := ch.Attack(nil); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 40)
	checkFrames(t, hits["Normal 1-1"], []int{33})
	checkFrames(t, hits["Evilsbane Blade"], []int{35})
}
