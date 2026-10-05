package sandrone

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
)

func TestCountedChargeStopsWithoutFollowingAction(t *testing.T) {
	for _, tc := range []struct {
		cons, count, lastFrame int
	}{
		{0, 1, 111}, {0, 2, 168}, {0, 3, 228}, {1, 6, 402}, {6, 6, 402},
	} {
		c, ch, hits := setupTiming(t, tc.cons)
		if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"rays": tc.count}); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, tc.lastFrame-1)
		if !c.Player.IsAnimationLocked(action.ActionBurst) {
			t.Fatal("next action unlocked before final ray")
		}
		advanceTo(t, c, tc.lastFrame)
		if c.Player.IsAnimationLocked(action.ActionBurst) || ch.resolutionChannel {
			t.Fatal("final ray did not release charge")
		}
		advanceTo(t, c, tc.lastFrame+200)
		if got := len(hits["Faggio Condensing Ray"]); got != tc.count {
			t.Fatalf("C%d requested %d rays, got %d", tc.cons, tc.count, got)
		}
		if len(hits["Faggio Power Overdrive Ray"]) != 0 {
			t.Fatal("counted charge continued firing in overdrive")
		}
	}
}

func TestCountedChargeCountsEmissionsNotTargets(t *testing.T) {
	for _, miss := range []bool{false, true} {
		c, ch, hits := setupTiming(t, 0)
		second := enemy.New(c, info.EnemyProfile{Level: 90, Pos: info.Coord{R: 1}})
		c.Combat.AddEnemy(second)
		fired := 0
		c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
			atk := args[0].(*info.AttackEvent)
			if atk.Info.Abil != "Faggio Condensing Ray" {
				return
			}
			fired++
			if miss {
				// Move both targets after the attack pattern is constructed.
				pos := info.Point{X: float64(100 * fired)}
				c.Combat.PrimaryTarget().SetPos(pos)
				second.SetPos(pos)
			}
		}, "test-count-ray-emissions")
		if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"射线": 3}); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, 450)
		wantHits := 6
		if miss {
			wantHits = 0
		}
		if fired != 3 || ch.resolutionChannel || len(hits["Faggio Condensing Ray"]) != wantHits {
			t.Fatalf("miss=%v fired=%d hits=%v", miss, fired, hits["Faggio Condensing Ray"])
		}
	}
}

func TestCountedChargeReleasesOnEarlyOverdrive(t *testing.T) {
	for _, tc := range []struct {
		power     float64
		end, rays int
	}{{80, 111, 1}, {99, 60, 0}} {
		c, ch, hits := setupTiming(t, 0)
		ch.resolutionPower = tc.power
		if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"rays": 3}); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, tc.end)
		if ch.resolutionChannel || !ch.powerOverdrive || c.Player.IsAnimationLocked(action.ActionSkill) {
			t.Fatal("early overdrive did not release the counted action")
		}
		advanceTo(t, c, 450)
		if len(hits["Faggio Condensing Ray"]) != tc.rays || len(hits["Faggio Power Overdrive Ray"]) != 0 {
			t.Fatalf("residual power=%v fired fictitious/extra rays: %v", tc.power, hits)
		}
	}
}

func TestCountedChargeInvalidRequestsDoNotStartChannel(t *testing.T) {
	for _, params := range []map[string]int{
		{"rays": 0}, {"射线": -1}, {"rays": 4},
		{"rays": 3, "射线": 2}, {"rays": 3, "duration": 200},
	} {
		_, ch, _ := setupTiming(t, 0)
		if _, err := ch.ChargeAttack(params); err == nil {
			t.Fatalf("invalid params accepted: %v", params)
		}
		if ch.resolutionChannel || ch.resolutionSrc != 0 {
			t.Fatal("invalid parameters mutated channel state")
		}
	}
	_, ch, _ := setupTiming(t, 0)
	ch.powerOverdrive, ch.resolutionPower = true, 80
	if _, err := ch.ChargeAttack(map[string]int{"rays": 3}); err == nil {
		t.Fatal("overdrive accepted condensing ray request")
	}
}

func TestCountedChargeStopsWhileNextSkillIsOnCooldown(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 37)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"rays": 1}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 133)
	if ch.resolutionChannel || c.Player.IsAnimationLocked(action.ActionSkill) {
		t.Fatal("first ray failed to end charge")
	}
	if c.Player.ReadyCheck(action.ActionSkill, keys.Sandrone, nil) == nil {
		t.Fatal("ray count bypassed E cooldown")
	}
	advanceTo(t, c, 240)
	if err := c.Player.ReadyCheck(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	checkFrames(t, hits["Faggio Condensing Ray"], []int{133})
}
