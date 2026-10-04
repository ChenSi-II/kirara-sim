package characters

import (
	"math"
	"strings"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/albedo"
	_ "github.com/genshinsim/gcsim/internal/characters/beidou"
	_ "github.com/genshinsim/gcsim/internal/characters/cyno"
	_ "github.com/genshinsim/gcsim/internal/characters/diona"
	_ "github.com/genshinsim/gcsim/internal/characters/fischl"
	_ "github.com/genshinsim/gcsim/internal/characters/klee"
	_ "github.com/genshinsim/gcsim/internal/characters/mizuki"
	_ "github.com/genshinsim/gcsim/internal/characters/mona"
	_ "github.com/genshinsim/gcsim/internal/characters/qiqi"
	_ "github.com/genshinsim/gcsim/internal/characters/razor"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/characters/venti"
	_ "github.com/genshinsim/gcsim/internal/characters/wriothesley"
	_ "github.com/genshinsim/gcsim/internal/characters/yaemiko"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/gcs/validation"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func enhancedCore(t *testing.T, cons int, params map[string]int, team ...keys.Char) (*core.Core, *enemy.Enemy, *character.CharWrapper) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	e := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{Y: 1, R: 1}})
	c.Combat.AddEnemy(e)
	for i, key := range team {
		p := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
		p.Base.Ascension = 6
		if i == 0 {
			p.Base.Cons = cons
			if params != nil {
				p.Params = params
			}
		}
		p.Talents = info.TalentProfile{Attack: 10, Skill: 10, Burst: 10}
		if _, err := c.AddChar(p); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = e.Key()
	return c, e, c.Player.ByIndex(0)
}

func enhancedAdvance(c *core.Core, frames int) {
	for range frames {
		c.F++
		_ = c.Tick()
		c.Events.Emit(event.OnTick)
	}
}

func closeEnh(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-7 {
		t.Fatalf("got %.9f, want %.9f", got, want)
	}
}
func enhancedReaction(c *core.Core, e *enemy.Enemy, ev event.Event, actor int) {
	c.Events.Emit(ev, e, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor}})
}
func enhancedHit(c *core.Core, e *enemy.Enemy, actor int, tag attacks.AttackTag) {
	ai := info.AttackInfo{ActorIndex: actor, Abil: "Test hit", AttackTag: tag, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Physical, Mult: 1}
	c.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 0, 1)
	enhancedAdvance(c, 1)
}

func TestEnhancedBeidouHold(t *testing.T) {
	if err := validation.ValidateCharParamKeys(keys.Beidou, action.ActionSkill, []string{"hold", "hold_frames", "counter"}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name          string
		enabled, hold int
		wantStacks    int
		refund        bool
	}{{"two stacks", 1, 96, 2, true}, {"one stack", 1, 48, 1, true}, {"before threshold", 1, 47, 0, false}, {"legacy", 0, 96, 0, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, ch := enhancedCore(t, 0, map[string]int{"enhanced": tc.enabled, "start_energy": 0}, keys.Beidou)
			var hit info.AttackInfo
			c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
				a := args[1].(*info.AttackEvent)
				if strings.HasPrefix(a.Info.Abil, "Tidecaller") {
					hit = a.Info
				}
			}, "test")
			_, err := ch.Skill(map[string]int{"hold": 1, "hold_frames": tc.hold})
			if err != nil {
				t.Fatal(err)
			}
			enhancedAdvance(c, 23+tc.hold*tc.enabled)
			if tc.wantStacks > 0 && !strings.Contains(hit.Abil, "Level") {
				t.Fatal("hold did not accumulate counter")
			}
			if ch.StatusIsActive("beidou-enhanced-hold-icd") != tc.refund {
				t.Fatal("incorrect refund trigger")
			}
			if tc.refund {
				closeEnh(t, ch.Energy, float64(8*tc.wantStacks))
			}
			if tc.wantStacks == 2 {
				ready, _ := ch.ActionReady(action.ActionSkill, nil)
				if !ready {
					t.Fatal("two-stack refund must clear skill cooldown")
				}
			}
		})
	}
}

func TestEnhancedCynoSharedDuration(t *testing.T) {
	for _, burst := range []bool{false, true} {
		c, _, ch := enhancedCore(t, 2, nil, keys.Cyno, keys.Qiqi)
		c.StarReactions.SuperconductActive = true
		duration := 360
		if burst {
			_, _ = ch.Burst(nil)
			duration = 712
		} else {
			_, _ = ch.Skill(nil)
		}
		if ch.StatusDuration("cyno-q") != duration || ch.StatusDuration("cyno-shared-sunrise") != duration {
			t.Fatal("sunrise must match pactsworn")
		}
		enhancedAdvance(c, 100)
		_, _ = ch.Skill(nil)
		if burst {
			duration += 240
		}
		if ch.StatusDuration("cyno-q") != duration-100 || ch.StatusDuration("cyno-shared-sunrise") != duration-100 {
			t.Fatal("skill extension mismatched or E state extended")
		}
		enhancedAdvance(c, 50)
		c.Player.SetActive(1)
		c.Events.Emit(event.OnCharacterSwap, 0, 1)
		other := c.Player.ByIndex(1)
		if ch.StatusIsActive("cyno-q") || ch.StatusIsActive("cyno-shared-sunrise") {
			t.Fatal("Cyno retained state after swap")
		}
		if other.StatusDuration("cyno-shared-sunrise") != duration-150 {
			t.Fatal("swap reset remaining duration")
		}
		closeEnh(t, other.Stat(attributes.EM), 300)
		enhancedAdvance(c, duration-150+1)
		if other.StatusIsActive("cyno-shared-sunrise") {
			t.Fatal("transferred sunrise outlived state")
		}
	}
}

func TestEnhancedYaeRetainsSakura(t *testing.T) {
	for _, enabled := range []int{0, 1} {
		c, _, ch := enhancedCore(t, 2, map[string]int{"enhanced": enabled}, keys.YaeMiko, keys.Qiqi)
		for range 3 {
			_, _ = ch.Skill(nil)
			enhancedAdvance(c, 40)
		}
		closeEnh(t, ch.Stat(attributes.EM), 100+200*float64(enabled))
		_, _ = ch.Burst(nil)
		count := ch.Tags["totems"]
		if enabled == 1 && count != 3 {
			t.Fatal("enhanced burst consumed sakura")
		}
		if enabled == 0 && count != 0 {
			t.Fatal("legacy burst failed to consume sakura")
		}
		if enabled == 1 {
			enhancedAdvance(c, 1000)
			if ch.Tags["totems"] != 3 {
				t.Fatal("sakura missed 10 second extension")
			}
			enhancedAdvance(c, 600)
			if ch.Tags["totems"] != 0 {
				t.Fatal("sakura failed to expire")
			}
		}
	}
}

func TestEnhancedQiqiReactionsAndCharges(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Qiqi, keys.YumemizukiMizuki)
	_, _ = ch.Skill(nil)
	if ch.StatusDuration("qiqi-e") != 901 {
		t.Fatalf("unexpected E duration/status: %v", ch.StatusDuration("qiqi-e"))
	}
	star := info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(star), 0)
	c.StarReactions.SuperconductActive = true
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(star), .5)
	enhancedReaction(c, e, event.OnStarDiffusion, 1)
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionCryo}), 0)
	c.StarReactions.SuperconductActive = false
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionCryo}), .5)
	_, _ = ch.Burst(nil)
	c.Player.SetActive(1)
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	for i := range 5 {
		a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagReactionStarSuperconduct, Mult: 1}}
		c.Events.Emit(event.OnEnemyHit, e, a)
		want := 0.0
		if i < 4 {
			want = 6 * ch.TotalAtk()
		}
		closeEnh(t, a.Info.FlatDmg, want)
	}
	enhancedAdvance(c, 481)
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionCryo}), 0)
}

func TestEnhancedDionaExtraPaws(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Diona)
	count := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Icy Paw (Stellar)" {
			count++
		}
	}, "test")
	_, _ = ch.Skill(nil)
	enhancedReaction(c, e, event.OnSuperconduct, 0)
	enhancedReaction(c, e, event.OnStarDiffusion, 0)
	enhancedAdvance(c, 20)
	if count != 3 {
		t.Fatalf("want three extra paws, got %d", count)
	}
	enhancedAdvance(c, 191)
	enhancedReaction(c, e, event.OnStarDiffusion, 0)
	enhancedAdvance(c, 20)
	if count != 6 {
		t.Fatalf("extra paws failed after ICD, got %d", count)
	}
	enhancedAdvance(c, 1200)
	enhancedReaction(c, e, event.OnSuperconduct, 0)
	enhancedAdvance(c, 20)
	if count != 6 {
		t.Fatal("extra paws survived skill window")
	}
}

func TestEnhancedWriothesleyRadiantAttacks(t *testing.T) {
	c, _, ch := enhancedCore(t, 6, nil, keys.Wriothesley)
	c.StarReactions.SuperconductActive = true
	_, _ = ch.Skill(nil)
	enhancedAdvance(c, 30)
	var normals []info.AttackInfo
	var icicles int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if strings.HasPrefix(a.Info.Abil, "Normal") {
			if strings.Contains(a.Info.Abil, "C6") {
				icicles++
			} else {
				normals = append(normals, a.Info)
			}
		}
	}, "test")
	for range 5 {
		_, _ = ch.Attack(nil)
		enhancedAdvance(c, 60)
	}
	if len(normals) != 6 {
		t.Fatalf("normal hit count %d", len(normals))
	}
	for i, a := range normals {
		want := attacks.AttackTagNormal
		if i == 2 || i == 5 {
			want = attacks.AttackTagReactionStarSuperconduct
		}
		if a.AttackTag != want {
			t.Fatalf("hit %d has tag %v", i, a.AttackTag)
		}
	}
	if icicles != 1 {
		t.Fatalf("expected N5 icicle, got %d", icicles)
	}
	ch.SetHPByRatio(.1)
	if ch.ActionStam(action.ActionCharge, nil) != 0 {
		t.Fatal("radiant charge must be free")
	}
	before := ch.CurrentHPRatio()
	_, _ = ch.ChargeAttack(nil)
	enhancedAdvance(c, 20)
	closeEnh(t, ch.CurrentHPRatio()-before, .3)
	duration := ch.StatusDuration("wriothesley-e")
	_, _ = ch.ChargeAttack(nil)
	enhancedAdvance(c, 20)
	closeEnh(t, ch.CurrentHPRatio()-before, .3)
	if ch.StatusDuration("wriothesley-e") != duration-20 {
		t.Fatal("charge extended E twice")
	}
	closeEnh(t, ch.DamageReduction(0), .25)
}

func TestEnhancedMizukiEmpowerAndExpiry(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.YumemizukiMizuki, keys.Qiqi)
	baseEM := ch.Stat(attributes.EM)
	_, _ = ch.Skill(map[string]int{"travel": 0})
	closeEnh(t, ch.Stat(attributes.EM), baseEM*1.1)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 100+baseEM*.1)
	extra := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Dreamdrifter (Stellar)" {
			extra++
			closeEnh(t, a.Info.Mult, 10)
		}
	}, "test")
	enhancedReaction(c, e, event.OnStarDiffusion, 0)
	enhancedAdvance(c, 100)
	if extra != 1 {
		t.Fatalf("expected one empowered cloud, got %d", extra)
	}
	if !e.ResistModIsActive("mizuki-c2-res-anemo") { // Reapply via a hit to inspect the live aura.
		enhancedHit(c, e, 0, attacks.AttackTagNormal)
	}
	if !e.ResistModIsActive("mizuki-c2-res-anemo") {
		t.Fatal("dream resistance shred did not apply")
	}
	c.Player.SetActive(1)
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	enhancedHit(c, e, 1, attacks.AttackTagNormal)
	if e.ResistModIsActive("mizuki-c2-res-anemo") {
		t.Fatal("shred persisted after leaving dream")
	}
	closeEnh(t, ch.Stat(attributes.EM), baseEM)
}

func TestEnhancedRazorOverflowRequiresWolf(t *testing.T) {
	c, _, ch := enhancedCore(t, 6, nil, keys.Razor, keys.Sucrose)
	surges := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Surge of Lightning" {
			surges++
		}
	}, "test")
	for range 2 {
		ch.ResetActionCooldown(action.ActionSkill)
		_, _ = ch.Skill(nil)
		enhancedAdvance(c, 100)
	}
	if surges != 0 {
		t.Fatal("sigil overflow triggered without wolf")
	}
	baseCD := ch.Stat(attributes.CD)
	_, _ = ch.Burst(nil)
	enhancedAdvance(c, 100)
	closeEnh(t, ch.Stat(attributes.CD), baseCD+.5)
	for range 2 {
		ch.ResetActionCooldown(action.ActionSkill)
		_, _ = ch.Skill(nil)
		enhancedAdvance(c, 100)
	}
	if surges != 1 {
		t.Fatalf("expected one wolf overflow attack, got %d", surges)
	}
}

func TestEnhancedVentiHurricaneAndExtension(t *testing.T) {
	for _, hex := range []int{0, 1} {
		c, e, ch := enhancedCore(t, 0, map[string]int{"hexerei": hex}, keys.Venti, keys.Sucrose)
		ticks, arrows := 0, 0
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.Abil == "Wind's Grand Ode" {
				ticks++
			}
			if a.Info.Abil == "Hurricane Arrow" {
				arrows++
			}
		}, "test")
		_, _ = ch.Burst(nil)
		enhancedAdvance(c, 110)
		enhancedReaction(c, e, event.OnSwirlCryo, 0)
		if ch.StatusIsActive("venti-hexerei-dmg") != (hex == 1) {
			t.Fatal("incorrect hex buff gate")
		}
		for range 3 {
			_, _ = ch.Attack(nil)
			enhancedAdvance(c, 50)
		}
		enhancedAdvance(c, 500)
		wantTicks := 20
		if hex == 1 {
			wantTicks = 25
		}
		if ticks != wantTicks {
			t.Fatalf("hex=%d: ticks=%d want=%d", hex, ticks, wantTicks)
		}
		if hex == 1 && arrows != 4 {
			t.Fatalf("hurricane arrows=%d", arrows)
		}
	}
}

func TestEnhancedKleeSparkChargesAndNaturalExplosion(t *testing.T) {
	c, e, ch := enhancedCore(t, 0, nil, keys.Klee, keys.Sucrose)
	e.SetPos(info.Point{Y: 100})
	for range 2 {
		ch.ResetActionCooldown(action.ActionSkill)
		_, _ = ch.Skill(nil)
		enhancedAdvance(c, 100)
	}
	_, _ = ch.Burst(nil)
	for range 3 {
		if ch.ActionStam(action.ActionCharge, nil) != 0 {
			t.Fatal("missing guaranteed spark")
		}
		_, _ = ch.ChargeAttack(nil)
		enhancedAdvance(c, 100)
	}
	if ch.StatusIsActive("a1-spark") {
		t.Fatal("three sparks did not consume independently")
	}
	c, _, ch = enhancedCore(t, 4, nil, keys.Klee)
	explosions := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Sparkly Explosion (C4)" {
			explosions++
			closeEnh(t, a.Snapshot.Stats[attributes.DmgP], 1)
		}
	}, "test")
	_, _ = ch.Burst(nil)
	enhancedAdvance(c, 900)
	if explosions != 1 {
		t.Fatalf("natural expiry explosions=%d", explosions)
	}
	c.Events.Emit(event.OnCharacterSwap, 0, 0)
	enhancedAdvance(c, 5)
	if explosions != 1 {
		t.Fatal("burst exploded twice")
	}
}

func TestEnhancedFischlReactionBuffs(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Fischl, keys.Venti, keys.Qiqi)
	_, _ = ch.Skill(nil)
	enhancedAdvance(c, 100)
	c.Player.SetActive(1)
	enhancedReaction(c, e, event.OnOverload, 1)
	enhancedReaction(c, e, event.OnLunarCharged, 1)
	closeEnh(t, ch.Stat(attributes.EM), 190)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 190)
	closeEnh(t, c.Player.ByIndex(2).Stat(attributes.EM), 100)
	if err := c.Player.Exec(action.ActionAttack, keys.Venti, nil); err != nil {
		t.Fatal(err)
	}
	enhancedAdvance(c, 50)
	closeEnh(t, ch.Stat(attributes.EM), 280)
	enhancedAdvance(c, 600)
	closeEnh(t, ch.Stat(attributes.EM), 100)
}

func TestEnhancedAlbedoSilverAndOffFieldC2(t *testing.T) {
	c, e, ch := enhancedCore(t, 2, nil, keys.Albedo, keys.Sucrose)
	baseDEF := ch.Stat(attributes.DEFP)
	_, _ = ch.Skill(nil)
	enhancedAdvance(c, 30)
	closeEnh(t, ch.Stat(attributes.DEFP), baseDEF+.5)
	c.Player.SetActive(1)
	blossoms := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Fatal Blossom (C2)" {
			blossoms++
			closeEnh(t, a.Info.Mult, 3)
			if !a.Info.UseDef {
				t.Fatal("C2 must use DEF")
			}
		}
	}, "test")
	for range 4 {
		enhancedHit(c, e, 1, attacks.AttackTagNormal)
		enhancedAdvance(c, 121)
	}
	if blossoms != 3 {
		t.Fatalf("expected three off-field flowers, got %d", blossoms)
	}
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 225)
	if !ch.StatusIsActive("albedo-hex-silver") {
		t.Fatal("silver creation did not grant damage buff")
	}
}

func TestEnhancedMonaOmenExtensionCapAndEMDuration(t *testing.T) {
	c, e, ch := enhancedCore(t, 0, nil, keys.Mona, keys.Sucrose)
	e.AddStatus("omen-debuff", 300, false)
	e.SetTag("omen-debuff", 4)
	for range 6 {
		_, _ = ch.Attack(nil)
		enhancedAdvance(c, 60)
	}
	if got := e.StatusExpiry("omen-debuff"); got != 780 {
		t.Fatalf("omen expiry=%d, want 780 (8 second cap)", got)
	}
	c, _, ch = enhancedCore(t, 2, nil, keys.Mona)
	_, _ = ch.ChargeAttack(nil)
	enhancedAdvance(c, 100)
	if got := ch.StatusDuration("mona-hexerei-c2-em"); got < 620 || got > 720 {
		t.Fatalf("C2 EM duration=%d", got)
	}
}

func TestEnhancedSucroseBuffRecipients(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Sucrose, keys.Klee, keys.Qiqi)
	_, _ = ch.Skill(nil)
	for i := range 3 {
		if !c.Player.ByIndex(i).StatusIsActive("sucrose-hexerei-skill") {
			t.Fatal("skill buff missing")
		}
	}
	_, _ = ch.Burst(nil)
	if !c.Player.ByIndex(1).StatusIsActive("sucrose-hexerei-burst") || c.Player.ByIndex(2).StatusIsActive("sucrose-hexerei-burst") {
		t.Fatal("burst buff recipients incorrect")
	}
	enhancedReaction(c, e, event.OnStarDiffusion, 0)
	closeEnh(t, c.Player.ByIndex(2).Stat(attributes.EM), 150)
	// The absorption box is centered five units in front of Sucrose.
	e.SetPos(info.Point{Y: 5})
	ai := info.AttackInfo{ActorIndex: 1, Abil: "Test pyro application", AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, Element: attributes.Pyro, Durability: 100, Mult: 1}
	c.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 0, 1)
	enhancedAdvance(c, 200)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.PyroP)-.288, .2857142)
	closeEnh(t, c.Player.ByIndex(2).Stat(attributes.PyroP), .2)
	enhancedAdvance(c, 281)
	if c.Player.ByIndex(1).StatusIsActive("sucrose-c6") {
		t.Fatal("C6 buff outlasted burst")
	}
}

func TestEnhancedCynoConstellationReactions(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Cyno, keys.Qiqi)
	c.StarReactions.SuperconductActive = true
	c.Player.ByIndex(1).Energy = 0
	// E's short Pactsworn state does not enable the original Q-only C4 refund.
	_, _ = ch.Skill(nil)
	enhancedReaction(c, e, event.OnStarSuperconduct, 0)
	closeEnh(t, c.Player.ByIndex(1).Energy, 0)
	_, _ = ch.Burst(nil)
	enhancedAdvance(c, 100)
	ch.Energy = 0
	bolts := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Raiment: Just Scales (C6)" {
			bolts++
			if a.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
				t.Fatal("C6 bolt not converted")
			}
			closeEnh(t, a.Info.Mult, 2)
			closeEnh(t, a.Info.FlatDmg, 6*ch.Stat(attributes.EM))
		}
	}, "test")
	_, _ = ch.Attack(nil)
	enhancedAdvance(c, 45)
	if bolts != 1 {
		t.Fatalf("C6 bolt count %d", bolts)
	}
	c.Player.ByIndex(1).Energy = 0
	for range 6 {
		enhancedReaction(c, e, event.OnStarSuperconduct, 0)
	}
	closeEnh(t, c.Player.ByIndex(1).Energy, 15)
	c.Player.SetActive(1)
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	ch.Energy = 0
	enhancedReaction(c, e, event.OnStarSuperconduct, 1)
	enhancedReaction(c, e, event.OnLunarCharged, 1)
	closeEnh(t, ch.Energy, 20)
	for range 7 {
		enhancedHit(c, e, 1, attacks.AttackTagNormal)
		enhancedAdvance(c, 7)
	}
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}), .8)
}

func TestEnhancedYaeEmpoweredTickAndC6(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.YaeMiko, keys.Qiqi)
	c.StarReactions.SuperconductActive = true
	_, _ = ch.Skill(nil)
	enhancedAdvance(c, 35)
	enhancedReaction(c, e, event.OnSuperconduct, 1)
	enhancedReaction(c, e, event.OnStarSuperconduct, 1)
	count := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Sesshou Sakura (Stellar)" {
			count++
			closeEnh(t, a.Info.Mult, 2)
			closeEnh(t, a.Snapshot.Stats[attributes.CD], ch.Stat(attributes.CD)+2)
		}
	}, "test")
	enhancedAdvance(c, 100)
	if count != 1 {
		t.Fatalf("empowered tick count=%d", count)
	}
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}), .5)
	enhancedAdvance(c, 650)
	closeEnh(t, c.Player.ByIndex(1).ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}), 0)
}

func TestEnhancedWriothesleyC2AndC6Multipliers(t *testing.T) {
	for _, radiance := range []bool{false, true} {
		c, _, ch := enhancedCore(t, 6, nil, keys.Wriothesley)
		c.StarReactions.SuperconductActive = radiance
		_, _ = ch.Skill(nil)
		enhancedAdvance(c, 30)
		var first, full float64
		var firstCrit, fullCrit float64
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
			a := args[1].(*info.AttackEvent)
			if a.Info.Abil == "Normal 2 (Enhanced)" {
				if first == 0 {
					first = a.Info.Mult
					firstCrit = a.Snapshot.Stats[attributes.CR]
				} else {
					full = a.Info.Mult
					fullCrit = a.Snapshot.Stats[attributes.CR]
				}
			}
		}, "test")
		for range 8 {
			_, _ = ch.Attack(nil)
			enhancedAdvance(c, 45)
		}
		ratio := 1.25
		if radiance {
			ratio = 1.5
		}
		closeEnh(t, full/first, ratio)
		bonus := 0.0
		if radiance {
			bonus = .1
		}
		closeEnh(t, firstCrit, ch.Stat(attributes.CR)+bonus)
		closeEnh(t, fullCrit, firstCrit)
	}
}

func TestEnhancedQiqiCoordinatedICD(t *testing.T) {
	c, e, ch := enhancedCore(t, 0, nil, keys.Qiqi, keys.Venti)
	_, _ = ch.Skill(nil)
	c.Player.SetActive(1)
	count := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil == "Herald of Frost (Coordinated)" {
			count++
			closeEnh(t, a.Info.Mult, .432)
		}
	}, "test")
	enhancedHit(c, e, 1, attacks.AttackTagNormal)
	enhancedAdvance(c, 2)
	enhancedHit(c, e, 1, attacks.AttackTagExtra)
	enhancedAdvance(c, 2)
	if count != 1 {
		t.Fatalf("coordinated attack ignored ICD: %d", count)
	}
	enhancedAdvance(c, 133)
	enhancedHit(c, e, 1, attacks.AttackTagReactionStarDiffusionAnemo)
	enhancedAdvance(c, 2)
	if count != 2 {
		t.Fatalf("direct stellar talent failed to trigger herald: %d", count)
	}
}

func TestEnhancedDisabledStellarKit(t *testing.T) {
	for _, key := range []keys.Char{keys.Beidou, keys.Qiqi, keys.Diona, keys.YaeMiko, keys.Cyno, keys.Wriothesley, keys.YumemizukiMizuki} {
		t.Run(key.String(), func(t *testing.T) {
			c, e, ch := enhancedCore(t, 6, map[string]int{"enhanced": 0}, key)
			c.StarReactions.SuperconductActive = true
			enhancedReaction(c, e, event.OnStarDiffusion, 0)
			if ch.StatusIsActive("stellar-diffusion-radiance") {
				t.Fatal("disabled kit gained radiance")
			}
			_, _ = ch.Skill(nil)
			if key == keys.Cyno && ch.StatusIsActive("cyno-q") {
				t.Fatal("legacy E entered Pactsworn")
			}
			_, _ = ch.Burst(nil)
			if key == keys.Qiqi && ch.StatusIsActive("qiqi-c6-stacks") {
				t.Fatal("disabled Qiqi gained charges")
			}
			enhancedAdvance(c, 200)
		})
	}
}

func TestEnhancedBeidouC6TracksActiveAndDomain(t *testing.T) {
	c, e, ch := enhancedCore(t, 6, nil, keys.Beidou, keys.Qiqi)
	c.StarReactions.SuperconductActive = true
	_, _ = ch.Burst(nil)
	closeEnh(t, ch.Stat(attributes.EM), 300)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 100)
	c.Player.SetActive(1)
	closeEnh(t, ch.Stat(attributes.EM), 100)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 300)
	enhancedHit(c, e, 1, attacks.AttackTagNormal)
	if !e.ResistModIsActive("beidou-c6-cryo") {
		t.Fatal("missing Cryo shred")
	}
	c.StarReactions.SuperconductActive = false
	enhancedHit(c, e, 1, attacks.AttackTagNormal)
	if e.ResistModIsActive("beidou-c6-cryo") {
		t.Fatal("Cryo shred persisted without domain")
	}
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.EM), 100)
}

func TestEnhancedWriothesleyC4DoesNotStackLegacyOverheal(t *testing.T) {
	c, _, ch := enhancedCore(t, 4, nil, keys.Wriothesley, keys.Qiqi)
	base := ch.Stat(attributes.AtkSpd)
	c.Player.Heal(info.HealInfo{Caller: 0, Target: 0, Message: "Test overheal", Src: 1000})
	closeEnh(t, ch.Stat(attributes.AtkSpd)-base, .2)
	c.StarReactions.SuperconductActive = true
	closeEnh(t, ch.Stat(attributes.AtkSpd)-base, .2)
	closeEnh(t, c.Player.ByIndex(1).Stat(attributes.AtkSpd), .1)
}
