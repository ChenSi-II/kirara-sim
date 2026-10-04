package odette

import (
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
	p := testhelper.DefaultProfile(keys.Odette, testhelper.TestWeaponKey)
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

func TestImageCodaUsesRelativeOffsets(t *testing.T) {
	c, ch, hits := setupTiming(t, 1)
	e, err := ch.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Frames(action.ActionSkill) != 46 {
		t.Fatal("E must allow E2 at 46f")
	}
	tickTo(c, 46)
	coda, err := ch.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if coda.AnimationLength != 82 {
		t.Fatal("E2 must last 82f")
	}
	tickTo(c, 130)
	assertFrames(t, hits["Phantom Night Dancers"], []int{28})
	assertFrames(t, hits["Coda at Dawn's Tolling"], []int{62, 72, 79})
	assertFrames(t, hits["Coda Finale"], []int{112})
	assertFrames(t, hits["Coda Finale (C1)"], []int{112})
}

func TestImageBurstHitsAfterSwapAndDoubleSchedule(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.Frames(action.ActionSwap) != 109 {
		t.Fatal("Q must allow swapping at 109f")
	}
	tickTo(c, 109)
	c.Player.SetActive(1)
	tickTo(c, 1280)
	assertFrames(t, hits["Bluebird Slash 1"], []int{119})
	assertFrames(t, hits["Bluebird Slash 2"], []int{137})
	assertFrames(t, hits["Bluebird Slash 3"], []int{142})
	assertFrames(t, hits["Bluebird Finale"], []int{148})
	assertFrames(t, hits["Dance Double Plume"], []int{257, 490, 723, 958, 1191})
	assertFrames(t, hits["Dance Double Wing"], []int{363, 598, 831, 1068})
}

func TestImmediateEEDoubleTimelineAndSwap(t *testing.T) {
	for _, cons := range []int{0, 1} {
		c, ch, hits := setupTiming(t, cons)
		c.StarReactions.SuperconductActive = true
		// Nonzero start makes accidental use of local/absolute times detectable.
		tickTo(c, 100)
		if err := c.Player.Exec(action.ActionSkill, keys.Odette, nil); err != nil {
			t.Fatal(err)
		}
		tickTo(c, 145)
		if c.Player.ReadyCheck(action.ActionSkill, keys.Odette, nil) == nil {
			t.Fatal("E2 unlocked too early")
		}
		tickTo(c, 146)
		if err := c.Player.ReadyCheck(action.ActionSkill, keys.Odette, nil); err != nil {
			t.Fatal(err)
		}
		if err := c.Player.Exec(action.ActionSkill, keys.Odette, nil); err != nil {
			t.Fatal(err)
		}
		tickTo(c, 227)
		if c.Player.ReadyCheck(action.ActionSwap, testhelper.TestCharKey, nil) == nil {
			t.Fatal("EE swap unlocked too early")
		}
		tickTo(c, 228)
		if err := c.Player.ReadyCheck(action.ActionSwap, testhelper.TestCharKey, nil); err != nil {
			t.Fatal(err)
		}
		if err := c.Player.Exec(action.ActionSwap, testhelper.TestCharKey, nil); err != nil {
			t.Fatal(err)
		}
		tickTo(c, 1392)
		assertFrames(t, hits["Phantom Night Dancers"], []int{128})
		assertFrames(t, hits["Coda at Dawn's Tolling"], []int{162, 172, 179})
		assertFrames(t, hits["Coda Finale"], []int{212})
		if cons == 1 {
			assertFrames(t, hits["Coda Finale (C1)"], []int{212})
		} else if len(hits["Coda Finale (C1)"]) != 0 {
			t.Fatal("C1 hit at C0")
		}
		assertFrames(t, hits["Dance Double Plume"], []int{261, 497, 731, 965, 1198})
		assertFrames(t, hits["Dance Double Wing"], []int{371, 607, 842, 1076, 1307})
		assertFrames(t, hits["Dance Double Plume Stellar"], []int{261, 497, 731, 965, 1198})
		assertFrames(t, hits["Dance Double Wing Stellar"], []int{371, 607, 842, 1076, 1307})
		if ch.StatusIsActive(doubleKey) {
			t.Fatal("EE double outlived recorded disappearance")
		}
	}
}

func TestEOnlyStillUsesItsOwnDoubleTimeline(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 1300)
	assertFrames(t, hits["Dance Double Plume"], []int{164, 395, 631, 864, 1097})
	assertFrames(t, hits["Dance Double Wing"], []int{269, 507, 741, 974, 1206})
}
