package linnea

import (
	"reflect"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/construct"
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
	p := testhelper.DefaultProfile(keys.Linnea, testhelper.TestWeaponKey)
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

type testMoondrift struct{}

func (testMoondrift) OnDestruct()                      {}
func (testMoondrift) Key() int                         { return 123 }
func (testMoondrift) Type() construct.GeoConstructType { return construct.GeoConstructLunarCrystallize }
func (testMoondrift) Expiry() int                      { return 9999 }
func (testMoondrift) IsLimited() bool                  { return true }
func (testMoondrift) Count() int                       { return 1 }
func (testMoondrift) Direction() info.Point            { return info.Point{} }
func (testMoondrift) Pos() info.Point                  { return info.Point{} }

func TestImageSingleAndContinuousTapSchedules(t *testing.T) {
	for _, taps := range []int{1, 5} {
		c, ch, hits := setupTiming(t, 0)
		e, err := ch.Skill(map[string]int{"taps": taps})
		if err != nil {
			t.Fatal(err)
		}
		wantSwap := 27
		if taps == 5 {
			wantSwap = 72
		}
		if e.Frames(action.ActionSwap) != wantSwap || e.CanQueueAfter > wantSwap {
			t.Fatal("wrong tap swap endpoint")
		}
		tickTo(c, wantSwap)
		c.Player.SetActive(1)
		tickTo(c, 620)
		if taps == 1 {
			assertFrames(t, hits["Lumi Pound-Pound Pummeler"], []int{118, 141, 263, 285, 410, 434, 557, 580})
			if len(hits["Million Ton Crush"]) != 0 {
				t.Fatal("single tap should not trigger E3")
			}
		} else {
			assertFrames(t, hits["Million Ton Crush"], []int{115})
			assertFrames(t, hits["Lumi Pound-Pound Pummeler"], []int{252, 273, 579, 601})
		}
	}
}

func TestImageMoonConstructBranches(t *testing.T) {
	for _, initial := range []bool{false, true} {
		c, ch, hits := setupTiming(t, 0)
		if initial {
			c.Constructs.New(testMoondrift{}, false)
		}
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		if !initial {
			tickTo(c, 200)
			c.Constructs.New(testMoondrift{}, false)
		}
		tickTo(c, 1000)
		if initial {
			assertFrames(t, hits["Heavy Overdrive Hammer"], []int{138, 459, 780})
			assertFrames(t, hits["Lumi Pound-Pound Pummeler"], []int{200, 224, 340, 361, 521, 545, 661, 682, 842, 866, 980})
		} else {
			assertFrames(t, hits["Heavy Overdrive Hammer"], []int{230, 553, 868})
			// The two no-construct recordings disagree by one frame on N1;
			// the initial no-construct profile uses the 118f first-column hit.
			assertFrames(t, hits["Lumi Pound-Pound Pummeler"], []int{118, 141, 292, 315, 418, 441, 613, 637, 757, 779, 932, 958})
		}
	}
}
