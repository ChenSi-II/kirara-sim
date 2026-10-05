package odette

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func TestA4MultipliesOriginalDamageInsteadOfAddingReactionBonus(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.ATK] = 4000
	buff[attributes.EM] = 1000
	buff[attributes.CR] = -1
	ch.AddStatMod(character.StatMod{Base: modifier.NewBase("test-a4-stats", -1), AffectedStat: attributes.NoStat, Amount: func() []float64 { return buff }})
	ch.AddReactBonusMod(character.ReactBonusMod{Base: modifier.NewBase("test-react", -1), Amount: func(info.AttackInfo) float64 { return .60 }})
	damages := []float64{}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) { damages = append(damages, args[2].(float64)) }, "test-a4-damage")
	ai := info.AttackInfo{ActorIndex: 0, Abil: "test star", Element: attributes.Cryo, AttackTag: attacks.AttackTagReactionStarSuperconduct, ICDTag: attacks.ICDTagNone, Mult: 2, FlatDmg: 100, Elevation: .45}
	ch.Base.Ascension = 0
	c.QueueAttack(ai, combat.NewSingleTargetHit(c.Combat.PrimaryTarget().Key()), 0, 0)
	tickTo(c, 1)
	ch.Base.Ascension = 6
	c.QueueAttack(ai, combat.NewSingleTargetHit(c.Combat.PrimaryTarget().Key()), 0, 0)
	tickTo(c, 2)
	if len(damages) != 2 || math.Abs(damages[1]/damages[0]-1.3) > 1e-9 {
		t.Fatalf("A4 damages=%v, want ratio 1.3", damages)
	}
	if math.Abs(ch.ReactBonus(ai)-.60) > 1e-9 {
		t.Fatal("A4 leaked into additive reaction bonus")
	}
	for _, actor := range []int{0, 1} {
		a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor, AttackTag: attacks.AttackTagReactionStarDiffusionAnemo, IsStarDiffusionReaction: true, Mult: .75}}
		c.Events.Emit(event.OnStarReactionAttack, c.Combat.PrimaryTarget(), a)
		want := .75
		if actor == 0 {
			want *= 1.3
		}
		if math.Abs(a.Info.Mult-want) > 1e-9 {
			t.Fatalf("actor %d contribution=%v", actor, a.Info.Mult)
		}
	}
}

func TestC2ResistanceRequiresDoubleAndCurrentRadiance(t *testing.T) {
	c, ch, _ := setupTiming(t, 2)
	e := c.Combat.PrimaryTarget()
	// The fixture has 0 resistance. A 20% shred changes the resistance
	// multiplier to 1.1, for the appropriate elements only.
	damage := 0.0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) { damage = args[2].(float64) }, "test-c2-damage")
	hit := func(ele attributes.Element) float64 {
		ai := info.AttackInfo{ActorIndex: 1, Abil: "test c2", Element: ele, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, Mult: 1}
		snap := info.Snapshot{CharLvl: 90}
		snap.Stats[attributes.BaseATK] = 1000
		c.QueueAttackWithSnap(ai, snap, combat.NewSingleTargetHit(e.Key()), 0)
		tickTo(c, c.F+1)
		return damage
	}
	base := hit(attributes.Cryo)
	c.StarReactions.SuperconductActive = true
	if got := hit(attributes.Cryo); math.Abs(got-base) > 1e-9 {
		t.Fatal("C2 active without a double")
	}
	ch.AddStatus(doubleKey, 2000, true)
	for _, ele := range []attributes.Element{attributes.Cryo, attributes.Electro, attributes.Anemo} {
		want := base
		if ele != attributes.Anemo {
			want *= 1.1
		}
		if got := hit(ele); math.Abs(got-want) > 1e-8 {
			t.Fatalf("conduct %v damage=%v want=%v", ele, got, want)
		}
	}
	c.StarReactions.SuperconductActive = false
	c.Events.Emit(event.OnStarDiffusion, e, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}})
	for _, ele := range []attributes.Element{attributes.Cryo, attributes.Anemo, attributes.Electro} {
		want := base
		if ele != attributes.Electro {
			want *= 1.1
		}
		if got := hit(ele); math.Abs(got-want) > 1e-8 {
			t.Fatalf("diffusion %v damage=%v want=%v", ele, got, want)
		}
	}
	tickTo(c, c.F+480)
	if got := hit(attributes.Anemo); math.Abs(got-base) > 1e-8 {
		t.Fatal("C2 outlived diffusion Radiance")
	}
}

func TestBurstCannotBypassCodaCooldown(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 46)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 130)
	if _, err := ch.Burst(nil); err != nil {
		t.Fatal(err)
	}
	if ready, _ := ch.ActionReady(action.ActionSkill, nil); ready {
		t.Fatal("Q reopened Coda before its separate cooldown")
	}
	tickTo(c, 1031)
	// A fresh window, after the 15s cooldown, permits another Coda.
	if _, err := ch.Burst(nil); err != nil {
		t.Fatal(err)
	}
	if ready, _ := ch.ActionReady(action.ActionSkill, nil); !ready {
		t.Fatal("Coda remained on cooldown")
	}
}

func TestCodaWindowIsSixSecondsAfterEOrQ(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 359)
	if !ch.StatusIsActive(codaKey) {
		t.Fatal("Coda window ended before six seconds")
	}
	tickTo(c, 361)
	if ch.StatusIsActive(codaKey) {
		t.Fatal("Coda window outlived six seconds")
	}
	if _, err := ch.Burst(nil); err != nil {
		t.Fatal(err)
	}
	if !ch.StatusIsActive(codaKey) {
		t.Fatal("Q did not open a new six-second Coda window")
	}
}
