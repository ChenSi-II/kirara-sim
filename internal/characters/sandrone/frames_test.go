package sandrone

import (
	"fmt"
	"reflect"
	"testing"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
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
	p := testhelper.DefaultProfile(keys.Sandrone, testhelper.TestWeaponKey)
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

func TestBurstHitsContinueAfterAnimation(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.AnimationLength != 99 {
		t.Fatal("wrong burst length")
	}
	advanceTo(t, c, 99)
	if len(hits) != 0 {
		t.Fatal("burst damage landed before the recorded time")
	}
	c.Player.SetActive(1)
	advanceTo(t, c, 180)
	for i, frame := range []int{130, 139, 151} {
		checkFrames(t, hits[fmt.Sprintf("Prismatic Bombardment %d", i+1)], []int{frame})
	}
	checkFrames(t, hits["Convective Inhibition Ray"], []int{180})
}

func TestResolutionC1Recording(t *testing.T) {
	c, ch, hits := setupTiming(t, 1)
	a, err := ch.ChargeAttack(map[string]int{"duration": 500})
	if err != nil {
		t.Fatal(err)
	}
	if a.AnimationLength != 500 {
		t.Fatal("charge does not occupy its hold duration")
	}
	advanceTo(t, c, 401)
	if ch.powerOverdrive {
		t.Fatal("C1 entered overdrive before ray six")
	}
	advanceTo(t, c, 402)
	if !ch.powerOverdrive || ch.resolutionRays != 6 {
		t.Fatal("C1 must enter overdrive on ray six")
	}
	advanceTo(t, c, 600)
	checkFrames(t, hits["Faggio Resolution Sweep"], []int{49, 71, 93, 109, 129, 149, 169, 189, 211, 229, 250, 271, 289, 308, 329, 349, 370, 389, 409})
	checkFrames(t, hits["Faggio Condensing Ray"], []int{109, 169, 225, 285, 342, 402})
	// C1 Z4 timing is extrapolated from C0's transition-relative offsets,
	// not directly measured in the C1 recording.
	checkFrames(t, hits["Faggio Power Overdrive Ray"], []int{422, 447, 470, 493})
	if ch.resolutionChannel || ch.StatusIsActive("sandrone-resolution") {
		t.Fatal("held action survived its requested duration")
	}
}

func TestResolutionStopsWhenInterrupted(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	a, err := ch.ChargeAttack(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 120)
	a.OnRemoved(action.SwapState)
	c.Player.SetActive(1)
	advanceTo(t, c, 500)
	checkFrames(t, hits["Faggio Resolution Sweep"], []int{49, 70, 91, 109})
	checkFrames(t, hits["Faggio Condensing Ray"], []int{111})
}

func TestSkillChargeRecording(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 37)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 220}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 247)
	if ch.powerOverdrive {
		t.Fatal("E->CA overheated before third ray")
	}
	advanceTo(t, c, 248)
	if !ch.powerOverdrive {
		t.Fatal("E->CA must enter overdrive on third ray")
	}
	advanceTo(t, c, 270)
	checkFrames(t, hits["Prism Shot 1"], []int{36})
	checkFrames(t, hits["Prism Shot 2"], []int{48})
	checkFrames(t, hits["Faggio Resolution Sweep"], []int{69, 88, 108, 129, 150, 171, 191, 212, 233, 251})
	checkFrames(t, hits["Faggio Condensing Ray"], []int{133, 191, 248})
	// 4.133s is Z4 in the standalone C0 column, but Z2/Z3 (the third
	// condensing ray) in the E -> CA column. Do not share that mode boundary.
	if len(hits["Faggio Power Overdrive Ray"]) != 0 {
		t.Fatal("standalone overdrive timing leaked into the E -> CA recording")
	}
}

func TestSkillChargeFollowupWindow(t *testing.T) {
	for _, tc := range []struct {
		name       string
		delay      int
		firstSweep int
		firstRay   int
	}{
		{"immediate", 37, 32, 96},
		{"one-frame-scheduling", 38, 32, 96},
		{"outside-window", 39, 49, 111},
		{"long-wait", 240, 49, 111},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, hits := setupTiming(t, 0)
			// Start E away from frame zero to ensure the follow-up is based
			// on the stored E origin, not the global frame or LastAction alone.
			advanceTo(t, c, 100)
			if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
				t.Fatal(err)
			}
			startCharge := 100 + tc.delay
			advanceTo(t, c, startCharge)
			if c.Player.IsAnimationLocked(action.ActionCharge) {
				t.Fatal("charge should be ready after E")
			}
			if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 120}); err != nil {
				t.Fatal(err)
			}
			advanceTo(t, c, startCharge+121)
			sweeps := hits["Faggio Resolution Sweep"]
			rays := hits["Faggio Condensing Ray"]
			if len(sweeps) == 0 || sweeps[0] != startCharge+tc.firstSweep {
				t.Fatalf("sweeps = %v", sweeps)
			}
			checkFrames(t, rays, []int{startCharge + tc.firstRay})
		})
	}
}

func TestCondensingRayIsBlunt(t *testing.T) {
	for _, stellar := range []bool{false, true} {
		c, _, _ := setupTiming(t, 0)
		c.StarReactions.SuperconductActive = stellar
		rayHits := 0
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			atk := args[1].(*info.AttackEvent)
			if atk.Info.Abil != "Faggio Condensing Ray" {
				return
			}
			rayHits++
			if atk.Info.StrikeType != attacks.StrikeTypeBlunt {
				t.Fatal("Z2/Z3 must be blunt")
			}
		}, "check-ray-strike-type")
		if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 120}); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, 120)
		if rayHits != 1 {
			t.Fatalf("observed %d rays, expected one", rayHits)
		}
	}
}

func TestC0ChargeMatchesOverdriveRecording(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 540}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 227)
	if ch.powerOverdrive {
		t.Fatal("overdrive before the third ray")
	}
	advanceTo(t, c, 228)
	if !ch.powerOverdrive || ch.resolutionRays != 3 {
		t.Fatal("C0 must enter overdrive on the third ray at 228f")
	}
	advanceTo(t, c, 600)
	checkFrames(t, hits["Faggio Resolution Sweep"], []int{49, 70, 91, 109, 128, 148, 168, 190, 209, 229})
	checkFrames(t, hits["Faggio Condensing Ray"], []int{111, 168, 228})
	checkFrames(t, hits["Faggio Power Overdrive Ray"], []int{248, 273, 296, 319, 343, 368, 392, 415, 439, 463, 486, 510, 535})
}

func TestRayCrossingThresholdSwitchesImmediately(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	ch.resolutionPower = 80
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 160}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 111)
	if !ch.powerOverdrive {
		t.Fatal("ray did not trigger overdrive when reaching 100")
	}
	advanceTo(t, c, 160)
	checkFrames(t, hits["Faggio Condensing Ray"], []int{111})
	checkFrames(t, hits["Faggio Power Overdrive Ray"], []int{131, 156})
}

func TestSkillChargeLoopDoesNotInheritOldChannel(t *testing.T) {
	for _, cons := range []int{0, 1} {
		c, ch, hits := setupTiming(t, cons)
		for _, origin := range []int{0, 240, 480} {
			advanceTo(t, c, origin)
			if c.Player.ReadyCheck(action.ActionSkill, keys.Sandrone, nil) != nil {
				t.Fatalf("C%d next E not ready at %d", cons, origin)
			}
			if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
				t.Fatal(err)
			}
			advanceTo(t, c, origin+37)
			if c.Player.ReadyCheck(action.ActionCharge, keys.Sandrone, nil) != nil {
				t.Fatal("CA not ready after E")
			}
			if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
				t.Fatal(err)
			}
			advanceTo(t, c, origin+239)
			if !c.Player.CanQueueNextAction() {
				t.Fatal("default channel blocked queuing next E")
			}
			if c.Player.ReadyCheck(action.ActionSkill, keys.Sandrone, nil) == nil {
				t.Fatal("E usable before cooldown")
			}
		}
		advanceTo(t, c, 720)
		checkFrames(t, hits["Prism Shot 1"], []int{36, 276, 516})
		checkFrames(t, hits["Prism Shot 2"], []int{48, 288, 528})
		// The first two loops retain nine sweeps and two rays. Residual power
		// can overheat C0 in the third loop; don't reset power to force repeats.
		// In particular the old channel's third ray at E+248 must vanish.
		wantSweeps, wantRays := []int{}, []int{}
		for _, origin := range []int{0, 240, 480} {
			for _, f := range []int{69, 88, 108, 129, 150, 171, 191, 212, 233} {
				if cons == 0 && origin == 480 && f > 191 {
					continue
				}
				wantSweeps = append(wantSweeps, origin+f)
			}
			for _, f := range []int{133, 191} {
				wantRays = append(wantRays, origin+f)
			}
		}
		checkFrames(t, hits["Faggio Resolution Sweep"], wantSweeps)
		checkFrames(t, hits["Faggio Condensing Ray"], wantRays)
		if cons == 0 {
			checkFrames(t, hits["Faggio Power Overdrive Ray"], []int{691, 716})
		} else if len(hits["Faggio Power Overdrive Ray"]) != 0 {
			t.Fatal("C1 half-rate incorrectly overheated")
		}
		if err := c.Player.ReadyCheck(action.ActionSkill, keys.Sandrone, nil); err != nil {
			t.Fatal("next E blocked after third loop", err)
		}
		if ch.tacticStacks == 0 {
			t.Fatal("repair/decay did not accumulate tactics across loop")
		}
	}
}

func TestSkillRepairDoesNotRequireStellarBuff(t *testing.T) {
	_, ch, _ := setupTiming(t, 0)
	ch.resolutionPower = 90
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if ch.resolutionPower != 40 {
		t.Fatal("E repair incorrectly requires Stellar status")
	}
	ch.tacticStacks, ch.tacticPowerRemoved = 0, 0
	ch.reduceResolutionPower(5)
	ch.reduceResolutionPower(5)
	if ch.tacticStacks != 1 {
		t.Fatal("partial power removal did not accumulate")
	}
}
