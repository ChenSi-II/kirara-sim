package characters

import (
	"strings"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/illuga"
	_ "github.com/genshinsim/gcsim/internal/characters/jahoda"
	_ "github.com/genshinsim/gcsim/internal/characters/nefer"
	_ "github.com/genshinsim/gcsim/internal/characters/odette"
	_ "github.com/genshinsim/gcsim/internal/characters/prune"
	_ "github.com/genshinsim/gcsim/internal/characters/sandrone"
	_ "github.com/genshinsim/gcsim/internal/characters/vesna"
	_ "github.com/genshinsim/gcsim/internal/characters/zibai"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// Expectations transcribed from the C3/C5 talent descriptions, in A/E/Q order.
func TestAuditConstellationTalentLevels(t *testing.T) {
	for _, tc := range []struct {
		key    keys.Char
		c3, c5 [3]int
	}{
		{keys.Sandrone, [3]int{12, 9, 9}, [3]int{12, 9, 12}},
		{keys.Prune, [3]int{9, 9, 12}, [3]int{9, 12, 12}},
		{keys.Illuga, [3]int{9, 9, 12}, [3]int{9, 12, 12}},
		{keys.Jahoda, [3]int{9, 9, 12}, [3]int{9, 12, 12}},
	} {
		for _, cons := range []int{3, 5} {
			_, _, ch := enhancedCore(t, cons, nil, tc.key)
			want := tc.c3
			if cons == 5 {
				want = tc.c5
			}
			got := [3]int{ch.TalentLvlAttack(), ch.TalentLvlSkill(), ch.TalentLvlBurst()}
			if got != want {
				t.Fatalf("%s C%d: %v, want %v", tc.key, cons, got, want)
			}
		}
	}
}

func TestAuditSandroneC4RequiresOwnDamage(t *testing.T) {
	c, e, ch := enhancedCore(t, 4, nil, keys.Sandrone, keys.Razor)
	count := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Prismatic Resonance Cannon (C4)" {
			count++
		}
	}, "audit-c4")
	enhancedReaction(c, e, event.OnStarSuperconduct, ch.Index())
	enhancedAdvance(c, 2)
	if count != 0 {
		t.Fatal("reaction event alone triggered C4")
	}
	queue := func(actor int) {
		ai := info.AttackInfo{ActorIndex: actor, Abil: "audit star hit", AttackTag: attacks.AttackTagReactionStarSuperconduct, ICDTag: attacks.ICDTagNone, Element: attributes.Cryo, Mult: 1}
		c.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 0, 1)
		enhancedAdvance(c, 3)
	}
	queue(1)
	if count != 0 {
		t.Fatal("teammate damage triggered C4")
	}
	queue(0)
	if count != 1 {
		t.Fatalf("own talent damage triggered %d C4 hits, want 1", count)
	}
	queue(0)
	if count != 1 {
		t.Fatal("C4 ignored its cooldown")
	}
}

func TestAuditStarBaseBonusAppliesToContributors(t *testing.T) {
	for _, key := range []keys.Char{keys.Sandrone, keys.Vesna, keys.Odette} {
		c, e, ch := enhancedCore(t, 0, nil, key)
		ai := info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagReactionStarDiffusionCryo}
		atk := &info.AttackEvent{Info: ai}
		c.Events.Emit(event.OnStarReactionAttack, e, atk)
		closeEnh(t, atk.Info.BaseDmgBonus, min(ch.TotalAtk()/100*.007, .14))
		atk.Info.IsStarDiffusionReaction = true
		before := atk.Info.BaseDmgBonus
		c.Events.Emit(event.OnEnemyHit, e, atk)
		closeEnh(t, atk.Info.BaseDmgBonus, before)
	}
}

func TestAuditOdetteSplendorRecipientsAndC6(t *testing.T) {
	for _, cons := range []int{0, 6} {
		c, e, ch := enhancedCore(t, cons, nil, keys.Odette, keys.Razor, keys.Diona)
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		c.Player.SetActive(1)
		enhancedAdvance(c, 61)
		ai := info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}
		want := .15
		if cons == 6 {
			want = .30
		}
		for _, idx := range []int{1, 2} {
			closeEnh(t, c.Player.ByIndex(idx).ReactBonus(ai), want)
		}
		if cons == 6 {
			atk := &info.AttackEvent{Info: ai}
			c.Events.Emit(event.OnEnemyHit, e, atk)
			closeEnh(t, atk.Info.Elevation, .45)
			enhancedAdvance(c, 180)
			closeEnh(t, c.Player.ByIndex(2).ReactBonus(ai), 1.2) // Four grants of two layers.
		}
		ch.DeleteStatus("odette-solo-dance-double")
		closeEnh(t, c.Player.ByIndex(2).ReactBonus(ai), 0)
	}
}

func TestAuditNeferShadeUsesEMAndC6ReplacesSelfHit(t *testing.T) {
	for _, cons := range []int{0, 6} {
		c, _, ch := enhancedCore(t, cons, nil, keys.Nefer, keys.Illuga)
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		c.Player.AddVerdantDew()
		if got := ch.ActionStam(action.ActionCharge, nil); got != 0 {
			t.Fatalf("phantasm costs %v stamina", got)
		}
		self, shade, lunar := 0, 0, 0
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if !strings.HasPrefix(a.Info.Abil, "Phantasm Performance") {
				return
			}
			if a.Info.AttackTag == attacks.AttackTagDirectLunarBloom {
				lunar++
				if !a.Info.UseEM {
					t.Fatal("lunar talent scales from ATK instead of EM")
				}
				if cons == 6 {
					closeEnh(t, a.Info.Elevation, .15)
				}
			} else {
				self++
			}
			if strings.Contains(a.Info.Abil, "Shade") {
				shade++
			}
		}, "audit-phantasm")
		if _, err := ch.ChargeAttack(nil); err != nil {
			t.Fatal(err)
		}
		enhancedAdvance(c, 80)
		wantSelf, wantLunar := 2, 3
		if cons == 6 {
			wantSelf, wantLunar = 1, 5
		}
		if self != wantSelf || lunar != wantLunar || shade != 3 {
			t.Fatalf("C%d: self=%d lunar=%d shade=%d", cons, self, lunar, shade)
		}
	}
}

func TestAuditZibaiC6IsSelfOnlyAndAppliedOnce(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Zibai, keys.Illuga)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	} // C1 starts at 100, consumes 100.
	seen := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Spirit Steed's Stride 2" {
			seen++
			closeEnh(t, a.Info.Elevation, .48)
		}
	}, "audit-zibai")
	enhancedAdvance(c, 40)
	if seen != 1 {
		t.Fatal("stride failed to hit")
	}
	a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagDirectLunarCrystallize}}
	c.Events.Emit(event.OnEnemyHit, e, a)
	closeEnh(t, a.Info.Elevation, 0)
	ch.DeleteStatus("zibai-lunar-phase-shift")
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(a.Info), 0)
}

func TestAuditIllugaConsumesOnHitAndUsesLunarScale(t *testing.T) {
	c, e, ch := enhancedCore(t, 4, nil, keys.Illuga, keys.Zibai)
	stats := make([]float64, attributes.EndStatType)
	stats[attributes.EM] = 200
	ch.AddStatMod(character.StatMod{Base: modifier.NewBase("audit-em", -1), AffectedStat: attributes.EM, Amount: func() []float64 { return stats }})
	if _, err := ch.Burst(nil); err != nil {
		t.Fatal(err)
	}
	// Merely applying an attack (including one that will miss) cannot spend
	// the 21 charges. No queued attacks are advanced during these assertions.
	for range 30 {
		a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Element: attributes.Geo, AttackTag: attacks.AttackTagDirectLunarCrystallize}}
		c.Events.Emit(event.OnApplyAttack, a)
		closeEnh(t, a.Info.FlatDmg, 0)
	}
	makeHit := func(tag attacks.AttackTag) *info.AttackEvent {
		a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Element: attributes.Geo, AttackTag: tag}}
		c.Events.Emit(event.OnEnemyHit, e, a)
		return a
	}
	ordinary := makeHit(attacks.AttackTagNormal)
	lunar := makeHit(attacks.AttackTagDirectLunarCrystallize)
	if ordinary.Info.FlatDmg <= 0 || lunar.Info.FlatDmg <= ordinary.Info.FlatDmg*5 {
		t.Fatalf("ordinary=%v lunar=%v: lunar must use its separate table", ordinary.Info.FlatDmg, lunar.Info.FlatDmg)
	}
	before := ch.Stat(attributes.DEF)
	for range 19 {
		makeHit(attacks.AttackTagNormal)
	}
	if ch.StatusIsActive("illuga-haunted-night-oriole-song") {
		t.Fatal("21 hits did not exhaust the song")
	}
	closeEnh(t, before-ch.Stat(attributes.DEF), 200)
	closeEnh(t, makeHit(attacks.AttackTagNormal).Info.FlatDmg, 0)
}
