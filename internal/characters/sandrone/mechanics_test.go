package sandrone

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

func TestOverdriveAllowsRepressWithoutResettingPower(t *testing.T) {
	c, ch, hits := setupTiming(t, 0)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 250}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 251)
	if ch.resolutionChannel || !ch.powerOverdrive {
		t.Fatal("release must stop shooting and retain overdrive")
	}
	power := ch.resolutionPower
	if err := c.Player.ReadyCheck(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal("overdrive prevented re-pressing charge", err)
	}
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 60}); err != nil {
		t.Fatal(err)
	}
	if ch.resolutionPower != power || !ch.powerOverdrive || ch.StatusIsActive("sandrone-resolution") {
		t.Fatal("re-press reset power or illegally entered Resolution")
	}
	advanceTo(t, c, 360)
	checkFrames(t, hits["Faggio Condensing Ray"], []int{111, 168, 228})
	// Restart startup is provisionally 20f. This checks cancellation of the
	// old firing loop, not an independent measurement of that startup.
	checkFrames(t, hits["Faggio Power Overdrive Ray"], []int{248, 271, 296})
	if ch.resolutionChannel {
		t.Fatal("restarted channel outlived its hold")
	}
}

func TestTacticsExpireAndCannotBuffLaterBurst(t *testing.T) {
	for _, expired := range []bool{false, true} {
		c, ch, _ := setupTiming(t, 0)
		c.StarReactions.SuperconductActive = true
		ch.resolutionPower = 100
		ch.reduceResolutionPower(20)
		if ch.tacticStacks != 2 {
			t.Fatal("20 removed power must grant two stacks")
		}
		advanceTo(t, c, 3599)
		if ch.tacticStacks != 2 {
			t.Fatal("tactics expired before 60 seconds")
		}
		if expired {
			advanceTo(t, c, 3600)
			if ch.tacticStacks != 0 {
				t.Fatal("tactics outlived 60 seconds")
			}
		}
		mult := 0.0
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.Abil == "Convective Inhibition Ray" {
				mult = a.Info.Mult
			}
		}, "test-tactic-beam")
		if _, err := ch.Burst(nil); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, c.F+181)
		want := burst[2][ch.TalentLvlBurst()]
		if !expired {
			want *= 1.2
		}
		if math.Abs(mult-want) > 1e-9 || ch.tacticStacks != 0 {
			t.Fatalf("expired=%v mult=%v want=%v stacks=%v", expired, mult, want, ch.tacticStacks)
		}
	}
}

func TestERepairsTheEntirePowerBar(t *testing.T) {
	_, ch, _ := setupTiming(t, 0)
	ch.resolutionPower = 73
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if ch.resolutionPower != 0 {
		t.Fatalf("E repaired %v power; want the whole bar", ch.resolutionPower)
	}
}

func TestC6ClusterRunsUntilBeforeSixthRay(t *testing.T) {
	c, _, hits := setupTiming(t, 6)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, map[string]int{"duration": 500}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 500)
	clusters := hits["Faggio Cluster Condensing Ray"]
	if len(clusters) != 4 {
		t.Fatalf("cluster segments = %v, want four", clusters)
	}
	// C1's sixth normal ray is at 402f in the confirmed recording. The
	// cluster offsets are provisional, but all segments must precede it.
	if clusters[len(clusters)-1] >= 402 {
		t.Fatalf("cluster outlived sixth ray: %v", clusters)
	}
}

func TestIntrinsicRadianceHasOwnLifetimeAndSuperconductPriority(t *testing.T) {
	for _, song := range []bool{false, true} {
		c, ch, _ := setupTiming(t, 0)
		c.StarReactions.DiffusionActive = true
		if ch.DiffusionRadiance() {
			t.Fatal("a global vortex flag granted character Radiance")
		}
		if song {
			ally := c.Player.ByIndex(1)
			ally.Base.Key = keys.Vodyanitsa
			ally.Base.Ascension = 1
			ally.AddStatus("vodyanitsa-microphone-summon", 300, true)
		}
		c.Events.Emit(event.OnStarDiffusion, c.Combat.PrimaryTarget(), &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}})
		if !ch.DiffusionRadiance() {
			t.Fatal("reaction did not grant Radiance")
		}
		c.StarReactions.SuperconductActive = true
		if !ch.SuperconductRadiance() || ch.DiffusionRadiance() {
			t.Fatal("Superconduct must take priority")
		}
		beamTag := attacks.AttackTagNone
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.Abil == "Convective Inhibition Ray" {
				beamTag = a.Info.AttackTag
			}
		}, "test-radiance-beam")
		if _, err := ch.Burst(nil); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, c, 181)
		if beamTag != attacks.AttackTagReactionStarSuperconduct {
			t.Fatal("Q overwrote Superconduct with Diffusion")
		}
		c.StarReactions.SuperconductActive = false
		advanceTo(t, c, 481)
		if ch.DiffusionRadiance() != song {
			t.Fatal("wrong Radiance lifetime at 8 seconds")
		}
		advanceTo(t, c, 721)
		if ch.DiffusionRadiance() {
			t.Fatal("Radiance outlived the 12-second extension")
		}
	}
}
