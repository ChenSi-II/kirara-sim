package prune

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func TestConversionUsesActualReactionAndQCannotOverwrite(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	e := c.Combat.PrimaryTarget()
	// Apply a small Pyro aura, then let the initial E consume it completely.
	ai := info.AttackInfo{ActorIndex: 1, Abil: "test pyro", Element: attributes.Pyro, AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagNone, Durability: 5, Mult: 1}
	c.QueueAttack(ai, combat.NewSingleTargetHit(e.Key()), 0, 1)
	tickTo(c, 2)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	tickTo(c, 34)
	if e.(info.Reactable).AuraContains(attributes.Pyro) {
		t.Fatal("fixture did not consume the aura")
	}
	if !ch.StatusIsActive(conversionKey) || ch.converted != attributes.Pyro {
		t.Fatal("E lost conversion when aura was consumed")
	}
	ch.AddStatus(bellKey, 600, true)
	c.Events.Emit(event.OnSwirlHydro, e, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Abil: "Witchlure Bell"}})
	tickTo(c, 83)
	if ch.converted != attributes.Pyro {
		t.Fatal("Q changed E's stored conversion")
	}
}

func TestHammerMissDoesNotGrantBenefitsOrBounce(t *testing.T) {
	c, ch, hits := setupTiming(t, 4)
	ch.Energy = 0
	ch.AddStatus(bellKey, 600, true)
	target := c.Combat.PrimaryTarget()
	// Move the victim between attack creation and collision, not after a
	// zero-delay attack (which the engine executes synchronously).
	c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		if args[0].(*info.AttackEvent).Info.Abil == "Banehunter Oathhammer" {
			target.SetPos(info.Point{X: 100})
		}
	}, "miss-hammer")
	ch.oathhammer(attributes.Cryo, target)
	tickTo(c, 80)
	if ch.Energy != 0 || ch.c2Stacks != 0 || c.Player.ByIndex(1).StatusIsActive("prune-tolling-rally") {
		t.Fatal("miss granted hit benefits")
	}
	if len(hits["Banehunter Oathhammer Bounce"]) != 0 {
		t.Fatal("miss spawned a bounce")
	}
}

func TestOneHammerBouncesOnceAcrossMultipleTargets(t *testing.T) {
	c, ch, _ := setupTiming(t, 4)
	c.Combat.AddEnemy(enemy.New(c, info.EnemyProfile{Level: 90, Pos: info.Coord{R: 1}}))
	applications := 0
	c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		if args[0].(*info.AttackEvent).Info.Abil == "Banehunter Oathhammer Bounce" {
			applications++
		}
	}, "test-bounces")
	ch.oathhammer(attributes.Cryo, c.Combat.PrimaryTarget())
	tickTo(c, 150)
	if applications != 1 {
		t.Fatalf("one hammer spawned %d bounce attacks", applications)
	}
}

func TestHexereiAndC6RequireRallyAndCorrectRecipient(t *testing.T) {
	c, ch, _ := setupTiming(t, 6)
	ally := c.Player.ByIndex(1)
	ally.IsHexerei = true
	c.Player.SetActive(1)
	if !ch.IsHexerei || c.Player.GetHexereiCount() != 2 {
		t.Fatal("Prune missing Hexerei identity")
	}
	selfP, allyP := ch.Stat(attributes.ATKP), ally.Stat(attributes.ATKP)
	selfFlat, allyFlat := ch.Stat(attributes.ATK), ally.Stat(attributes.ATK)
	hit := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}}
	c.Events.Emit(event.OnSwirlCryo, c.Combat.PrimaryTarget(), hit)
	if ch.Stat(attributes.ATKP) != selfP {
		t.Fatal("reaction without Rally buff triggered Hexerei")
	}
	ch.tollingRally()
	c.Events.Emit(event.OnSwirlCryo, c.Combat.PrimaryTarget(), hit)
	for _, check := range [][2]float64{{ch.Stat(attributes.ATKP) - selfP, .60}, {ally.Stat(attributes.ATKP) - allyP, .30}, {ch.Stat(attributes.ATK) - selfFlat, 350}, {ally.Stat(attributes.ATK) - allyFlat, 350}} {
		if math.Abs(check[0]-check[1]) > 1e-9 {
			t.Fatalf("buff=%v, want %v", check[0], check[1])
		}
	}
	tickTo(c, 301)
	if ch.Stat(attributes.ATKP) != selfP || ally.Stat(attributes.ATK) != allyFlat {
		t.Fatal("reaction buffs failed to expire")
	}
}

func TestRallyDoesNotBuffUnlistedDamageSources(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	buff := make([]float64, attributes.EndStatType)
	buff[attributes.ATK] = 5000
	ch.AddStatMod(character.StatMod{Base: modifier.NewBase("test-high-atk", -1), AffectedStat: attributes.ATK, Amount: func() []float64 { return buff }})
	ch.tollingRally()
	ally := c.Player.ByIndex(1)
	normal := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagNormal}}
	ally.ApplyAttackMods(normal, c.Combat.PrimaryTarget())
	if math.Abs(normal.Snapshot.Stats[attributes.DmgP]-.5) > 1e-9 {
		t.Fatal("eligible attack did not receive capped Rally")
	}
	a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagWeaponSkill}}
	ally.ApplyAttackMods(a, c.Combat.PrimaryTarget())
	if a.Snapshot.Stats[attributes.DmgP] != 0 {
		t.Fatal("Rally buffed weapon proc damage")
	}
}
