package mitya

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
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
	p := testhelper.DefaultProfile(keys.Mitya, testhelper.TestWeaponKey)
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

func TestBeaconGenerationAndExpiry(t *testing.T) {
	for _, hold := range []bool{false, true} {
		t.Run(map[bool]string{false: "steady", true: "overload"}[hold], func(t *testing.T) {
			c, ch, _ := setupTiming(t, 0)
			p := map[string]int{}
			if hold {
				p["hold"] = 1
			}
			ch.Skill(p)
			c.Events.Emit(event.OnStarSuperconduct, nil, &info.AttackEvent{})
			advanceTo(t, c, 89)
			if ch.beaconCount() != 0 {
				t.Fatal("beacon too early")
			}
			advanceTo(t, c, 90)
			if ch.beaconCount() != 1 || !c.StarReactions.SuperconductActive {
				t.Fatal("first beacon/domain missing")
			}
			advanceTo(t, c, 360)
			want := 2
			if hold {
				want = 4
			}
			if ch.beaconCount() != want {
				t.Fatalf("beacons %d want %d", ch.beaconCount(), want)
			}
			advanceTo(t, c, 961)
			if ch.beaconCount() != 0 || c.StarReactions.SuperconductActive {
				t.Fatal("beacons/domain failed to expire")
			}
		})
	}
}
func TestCapRefreshOverflowAndReplacement(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	ch.Skill(nil)
	for i := 0; i < 4; i++ {
		ch.addBeacon()
	}
	advanceTo(t, c, 100)
	ch.addBeacon()
	advanceTo(t, c, 101)
	if ch.beaconCount() != 4 || len(hits["Steady Core Beacon"]) != 1 {
		t.Fatal("overflow should consume only the incoming beacon")
	}
	advanceTo(t, c, 601)
	if ch.beaconCount() != 4 {
		t.Fatal("overflow must refresh all beacon lifetimes")
	}
	ch.ResetActionCooldown(action.ActionSkill)
	ch.Skill(map[string]int{"hold": 1})
	advanceTo(t, c, 602)
	if ch.beaconCount() != 0 || len(hits["Steady Core Detonation"]) != 1 {
		t.Fatal("replacing steady core must detonate all carried beacons")
	}
	advanceTo(t, c, 721)
	if len(hits["Steady Core"]) != 5 {
		t.Fatalf("old core kept attacking: %v", hits["Steady Core"])
	}
}
func TestOverloadChargeSpendsAndInterrupts(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	ch.Skill(map[string]int{"hold": 1})
	ch.addBeacon()
	ch.addBeacon()
	a, err := ch.ChargeAttack(map[string]int{"duration": 300})
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 90)
	if ch.beaconCount() != 1 || len(hits["Gradation Beacon"]) != 1 {
		t.Fatal("special charge did not spend once")
	}
	a.OnRemoved(0)
	advanceTo(t, c, 301)
	if ch.beaconCount() != 1 || len(hits["Gradation Beacon"]) != 1 {
		t.Fatal("interrupted charge kept consuming")
	}
}
func TestCoreBuffsFollowActiveCharacter(t *testing.T) {
	c, ch, _ := setupTiming(t, 2)
	ally := c.Player.ByIndex(1)
	baseSelf := ch.Stat(attributes.EM)
	baseAlly := ally.Stat(attributes.EM)
	baseCD := ch.Stat(attributes.CD)
	ch.Skill(map[string]int{"hold": 1})
	if ch.Stat(attributes.EM)-baseSelf != 200 || ally.Stat(attributes.EM) != baseAlly {
		t.Fatal("overload EM applies to the active character only")
	}
	if ch.Stat(attributes.CD)-baseCD < .499 {
		t.Fatal("C1 crit damage missing")
	}
	c.Player.SetActive(1)
	if ch.Stat(attributes.EM) != baseSelf || ally.Stat(attributes.EM)-baseAlly != 200 {
		t.Fatal("buff did not follow swap")
	}
}
func TestGenerationRefreshDoesNotMultiplyStreams(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	for i := 0; i < 10; i++ {
		c.Events.Emit(event.OnStarSuperconduct, nil, &info.AttackEvent{})
	}
	advanceTo(t, c, 180)
	if ch.beaconCount() != 2 {
		t.Fatalf("overlapping generation: %d", ch.beaconCount())
	}
}

func TestStarBaseBonusAndC6UseSeparateFactors(t *testing.T) {
	c, ch, _ := setupTiming(t, 6)
	a := &info.AttackEvent{Info: info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}}
	c.Events.Emit(event.OnApplyAttack, a)
	want := min(.14, ch.Stat(attributes.EM)*.00028)
	if math.Abs(a.Info.BaseDmgBonus-want) > 1e-9 || a.Info.Elevation != .2 {
		t.Fatalf("wrong factors: %+v", a.Info)
	}
	if ch.ReactBonus(a.Info) != 0 {
		t.Fatal("base bonus incorrectly added to reaction-bonus bucket")
	}
}
func TestCoreC4TriggersEveryFiveIncludingExpiry(t *testing.T) {
	c, ch, hits := setupTiming(t, 4)
	ch.Skill(nil)
	for i := 0; i < 4; i++ {
		ch.addBeacon()
	}
	ch.addBeacon()
	ch.addBeacon()
	advanceTo(t, c, 1)
	if len(hits["Core Fission C4"]) != 0 {
		t.Fatal("early C4")
	}
	// Five overflow consumptions in total; initial C1 beacon means two above.
	for i := 0; i < 3; i++ {
		ch.addBeacon()
	}
	advanceTo(t, c, 2)
	if len(hits["Core Fission C4"]) != 1 {
		t.Fatalf("C4 missing: %v", hits)
	}
	// Refresh all carried beacons shortly before natural core expiry.
	advanceTo(t, c, 1100)
	for i := 0; i < 5; i++ {
		ch.addBeacon()
	}
	advanceTo(t, c, 1201)
	if len(hits["Steady Core Detonation"]) != 1 || len(hits["Core Fission C4"]) != 2 {
		t.Fatalf("expiry failed to consume/C4: %v", hits)
	}
}
