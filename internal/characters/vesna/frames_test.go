package vesna

import (
	"reflect"
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

func advanceVesnaFrames(c *core.Core, count int) {
	for range count {
		c.F++
		c.Tick()
	}
}

func TestImageTimingAllowsActionsBeforeDelayedHits(t *testing.T) {
	for _, tc := range []struct {
		name string
		cast func(*char) (action.Info, error)
		end  int
		hit  int
	}{
		{"initial E", func(ch *char) (action.Info, error) { return ch.Skill(nil) }, 27, 33},
		{"Q", func(ch *char) (action.Info, error) { return ch.Burst(nil) }, 133, 135},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, ch := setupVesna(t, 0)
			var hits []int
			c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
				if args[1].(*info.AttackEvent).Info.ActorIndex == ch.Index() {
					hits = append(hits, c.F)
				}
			}, "timing-delayed-hit")
			ai, err := tc.cast(ch)
			if err != nil {
				t.Fatal(err)
			}
			advanceVesnaFrames(c, tc.end)
			if ai.AnimationLength != tc.end || ai.Frames(action.ActionSwap) != tc.end || len(hits) != 0 {
				t.Fatalf("animation end %d, swap %d, hits %v; expected action end %d before damage", ai.AnimationLength, ai.Frames(action.ActionSwap), hits, tc.end)
			}
			// The next action must not erase the already scheduled hit.
			if _, err := ch.Dash(nil); err != nil {
				t.Fatal(err)
			}
			advanceVesnaFrames(c, tc.hit-tc.end)
			if !reflect.DeepEqual(hits, []int{tc.hit}) {
				t.Fatalf("damage frames %v, expected [%d]", hits, tc.hit)
			}
		})
	}
}

func TestArmedNormalStringMatchesImageTimeline(t *testing.T) {
	c, ch := setupVesna(t, 0)
	ch.AddStatus(spiritbladeArmedKey, 1000, false)
	var hits []int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if strings.HasPrefix(a.Info.Abil, "Normal ") {
			hits = append(hits, c.F)
		}
	}, "timing-normal-string")
	for range 6 {
		ai, err := ch.Attack(nil)
		if err != nil {
			t.Fatal(err)
		}
		advanceVesnaFrames(c, ai.Frames(action.ActionAttack))
	}
	if want := []int{20, 45, 77, 93, 123, 153, 208}; !reflect.DeepEqual(hits, want) || c.F != 227 {
		t.Fatalf("normal hits %v, end %d; expected %v, end 227", hits, c.F, want)
	}
}

func TestArmedN1ChargePreservesDelayedFeathers(t *testing.T) {
	c, ch := setupVesna(t, 0)
	ch.AddStatus(spiritbladeArmedKey, 1000, false)
	var hits, feathers []int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Spirit Feather" {
			feathers = append(feathers, c.F)
		} else {
			hits = append(hits, c.F)
		}
	}, "timing-n1-charge")
	n1, err := ch.Attack(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceVesnaFrames(c, n1.Frames(action.ActionCharge))
	ca, err := ch.ChargeAttack(nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceVesnaFrames(c, ca.AnimationLength)
	if c.F != 78 {
		t.Fatalf("N1+CA ends at %d, expected 78", c.F)
	}
	advanceVesnaFrames(c, 115-c.F)
	if !reflect.DeepEqual(hits, []int{20, 58}) || !reflect.DeepEqual(feathers, []int{73, 115, 115}) {
		t.Fatalf("direct hits %v, feathers %v", hits, feathers)
	}
}

func TestSpecialSkillChainAndC6CancelFrames(t *testing.T) {
	c, ch := setupVesna(t, 0)
	ch.AddStatus(spiritbladeArmedKey, 1000, false)
	ch.magic = 5
	var hits []int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		hits = append(hits, c.F)
	}, "timing-special-chain")
	for range 3 {
		ai, err := ch.Skill(nil)
		if err != nil {
			t.Fatal(err)
		}
		advanceVesnaFrames(c, ai.Frames(action.ActionSkill))
	}
	if want := []int{13, 35, 66, 94, 103, 112, 121, 141}; !reflect.DeepEqual(hits, want) || c.F != 142 {
		t.Fatalf("EE123 hits %v, end %d; expected %v, end 142", hits, c.F, want)
	}

	_, c6 := setupVesna(t, 6)
	c6.AddStatus(stepReadyKey, 300, false)
	ai, err := c6.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if ai.AnimationLength != 97 || ai.Frames(action.ActionSkill) != 48 || ai.Frames(action.ActionDash) != 76 || ai.Frames(action.ActionSwap) != 97 {
		t.Fatalf("C6 cancels lost their action-specific timing: %+v", ai)
	}
}

func TestBurstToDanceUsesItsOwnCancelPoint(t *testing.T) {
	c, ch := setupVesna(t, 0)
	ch.AddStatus(spiritbladeArmedKey, 1000, false)
	ch.specialStage = 2
	ch.magic = 5
	var burstHits []int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Spiritblade: Burst" {
			burstHits = append(burstHits, c.F)
		}
	}, "timing-q-dance")
	q, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	if q.Frames(action.ActionSkill) != 131 || q.Frames(action.ActionSwap) != 133 {
		t.Fatalf("Q -> EE3 must become ready before swap: skill %d, swap %d", q.Frames(action.ActionSkill), q.Frames(action.ActionSwap))
	}
	advanceVesnaFrames(c, 131)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	advanceVesnaFrames(c, 4)
	if !reflect.DeepEqual(burstHits, []int{135}) {
		t.Fatalf("Q damage after EE3 starts: %v, expected [135]", burstHits)
	}
}

func TestBurstSkillCancelFollowsActualSkillBranch(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cons      int
		stage     int
		stepReady bool
		cancel    int
		ability   string
		hitmark   int
	}{
		{"C6 Step takes priority", 6, 2, true, 133, "Spiritblade: Step", 18},
		{"C6 Dance without Step", 6, 2, false, 131, "Spiritblade: Dance Blade", 28},
		{"second stage uses Q end", 0, 1, false, 133, "Spiritblade: Fall", 21},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, ch := setupVesna(t, tc.cons)
			ch.AddStatus(spiritbladeArmedKey, 1000, false)
			ch.specialStage = tc.stage
			ch.magic = 5
			if tc.stepReady {
				ch.AddStatus(stepReadyKey, 300, false)
			}
			var hits []int
			c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
				if args[1].(*info.AttackEvent).Info.Abil == tc.ability {
					hits = append(hits, c.F)
				}
			}, "timing-q-skill-branch")
			if err := c.Player.Exec(action.ActionBurst, keys.Vesna, nil); err != nil {
				t.Fatal(err)
			}
			advanceVesnaFrames(c, 131)
			err := c.Player.ReadyCheck(action.ActionSkill, keys.Vesna, nil)
			if (err == nil) != (tc.cancel == 131) {
				t.Fatalf("skill readiness at 131: %v; expected cancel at %d", err, tc.cancel)
			}
			advanceVesnaFrames(c, tc.cancel-131)
			if err := c.Player.ReadyCheck(action.ActionSkill, keys.Vesna, nil); err != nil {
				t.Fatalf("skill unavailable at its cancel frame %d: %v", tc.cancel, err)
			}
			if err := c.Player.Exec(action.ActionSkill, keys.Vesna, nil); err != nil {
				t.Fatal(err)
			}
			advanceVesnaFrames(c, tc.hitmark)
			if !reflect.DeepEqual(hits, []int{tc.cancel + tc.hitmark}) {
				t.Fatalf("actual skill branch %q hit at %v; expected [%d]", tc.ability, hits, tc.cancel+tc.hitmark)
			}
		})
	}
}
