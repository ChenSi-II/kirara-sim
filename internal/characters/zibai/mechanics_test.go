package zibai

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func TestStrideResetsLunarPhaseRecoveryCooldown(t *testing.T) {
	c, ch := rotationCore(t)
	c.Player.ByIndex(1).Moonsign = 1
	c.Events.Emit(event.OnLunarCrystallize)
	if ch.phase != 0 {
		t.Fatal("gained phase outside the stance")
	}
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	c.Events.Emit(event.OnLunarCrystallize)
	if ch.phase != 35 {
		t.Fatalf("first reaction phase=%v", ch.phase)
	}
	c.Events.Emit(event.OnLunarCrystallize)
	if ch.phase != 35 {
		t.Fatal("reaction cooldown missing")
	}
	ch.phase = 70
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	c.Events.Emit(event.OnLunarCrystallize)
	if ch.phase != 35 {
		t.Fatalf("stride did not reset cooldown: %v", ch.phase)
	}
}

func TestBurstExtensionKeepsNaturalPhaseRecovery(t *testing.T) {
	c, ch := rotationCore(t)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := ch.Burst(nil); err != nil {
		t.Fatal(err)
	}
	for c.F < 901 {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	ch.phase = 0
	for c.F < 960 {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	if ch.phase != 10 {
		t.Fatalf("extended stance stopped natural recovery: %v", ch.phase)
	}
	for c.F < 1004 {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	if ch.phase != 0 || ch.StatusIsActive(lunarPhaseKey) {
		t.Fatal("expired stance retained phase")
	}
}

func TestC6Requires70AndReplacesPreviousElevation(t *testing.T) {
	_, ch := rotationCore(t)
	ch.Base.Cons = 6
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	ch.phase = 15
	if ready, _ := ch.ActionReady(action.ActionSkill, nil); ready {
		t.Fatal("C6 bypassed 70 phase requirement")
	}
	ch.phase = 100
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if math.Abs(ch.c6Elevation-.48) > 1e-9 {
		t.Fatal("wrong 100-phase elevation")
	}
	ch.phase = 70
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if ch.c6Elevation != 0 {
		t.Fatal("70-phase stride retained old elevation")
	}
}

func TestC6LunarReactionIsNotElevatedTwice(t *testing.T) {
	c, ch := rotationCore(t)
	ch.Base.Cons = 6
	ch.c6Elevation = .48
	c.Events.Subscribe(event.OnLunarReactionAttack, func(args ...any) {
		args[1].(*info.AttackEvent).Snapshot.Stats[attributes.CR] = -1
	}, "test-c6-no-crit")
	damage := 0.0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
			return
		}
		if a.Info.Elevation != 0 {
			t.Fatal("C6 applied again to the aggregate reaction packet")
		}
		damage += args[2].(float64)
	}, "test-c6-aggregate-damage")
	apply := func(contributor int) float64 {
		damage = 0
		contributors := [info.MaxChars]bool{}
		contributors[contributor] = true
		// Use Zibai as reaction owner even for the ally-only contribution:
		// ownership must not spread her elevation to another contributor.
		reactable.DoLCrAttackWithContrib(contributors, c.Combat.PrimaryTarget(), c, 0)
		end := c.F + 60
		for c.F < end {
			c.F++
			if err := c.Tick(); err != nil {
				t.Fatal(err)
			}
		}
		return damage
	}
	selfBase, allyBase := apply(0), apply(1)
	ch.AddStatus("zibai-c6-elevation", 180, true)
	selfBoosted, allyBoosted := apply(0), apply(1)
	if selfBase <= 0 || math.Abs(selfBoosted/selfBase-1.48) > 1e-9 {
		t.Fatalf("self reaction damage base=%v boosted=%v", selfBase, selfBoosted)
	}
	if math.Abs(allyBoosted-allyBase) > 1e-9 {
		t.Fatalf("C6 changed ally contribution: base=%v boosted=%v", allyBase, allyBoosted)
	}
}

func TestC4RetainsNormalComboThroughPlayerExec(t *testing.T) {
	c, ch := rotationCore(t)
	ch.Base.Cons = 4
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	ch.NormalCounter = 2 // next normal is N3
	if err := c.Player.Exec(action.ActionBurst, keys.Zibai, nil); err != nil {
		t.Fatal(err)
	}
	if ch.NormalCounter != 2 {
		t.Fatal("Q reset C4's normal chain")
	}
	ch.DeleteStatus(lunarPhaseKey)
	ch.ResetNormalCounter()
	if ch.NormalCounter != 0 {
		t.Fatal("normal chain persisted outside stance")
	}
}

func TestC4RequiresStrideHitAndC1UsesReactionBucket(t *testing.T) {
	for _, miss := range []bool{false, true} {
		c, ch := rotationCore(t)
		ch.Base.Cons = 4
		// NewChar ran at C0; install the tested constellation hooks explicitly.
		ch.initConstellations()
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		seen := false
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.Abil != "Spirit Steed's Stride 2" {
				return
			}
			seen = true
			if math.Abs(ch.ReactBonus(a.Info)-2.5) > 1e-9 {
				t.Fatalf("first stride ReactBonus=%v", ch.ReactBonus(a.Info))
			}
			if a.Info.BaseDmgBonus > .14+1e-9 {
				t.Fatal("C1 still in the base bonus bucket")
			}
		}, "test-c1")
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		if ch.scattermoon {
			t.Fatal("C4 granted at cast rather than hit")
		}
		if miss {
			c.Combat.PrimaryTarget().SetPos(info.Point{X: 100})
		}
		for c.F < 35 {
			c.F++
			if err := c.Tick(); err != nil {
				t.Fatal(err)
			}
		}
		if ch.scattermoon == miss || seen == miss {
			t.Fatalf("miss=%v scattermoon=%v hit=%v", miss, ch.scattermoon, seen)
		}
		other := info.AttackInfo{ActorIndex: 0, Element: attributes.Geo, AttackTag: attacks.AttackTagDirectLunarCrystallize}
		if math.Abs(ch.ReactBonus(other)-.3) > 1e-9 {
			t.Fatal("C1 leaked to other lunar attacks")
		}
	}
}
