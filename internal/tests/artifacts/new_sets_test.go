package artifacts_test

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"

	_ "github.com/genshinsim/gcsim/pkg/simulation"
)

func init() {
	core.RegisterCharFunc(keys.TestCharDoNotUse, testhelper.NewChar)
}

func characterProfile(set keys.Set, count int) info.CharacterProfile {
	p := info.CharacterProfile{}
	p.Base.Key = keys.TestCharDoNotUse
	p.Base.Element = attributes.Anemo
	p.Base.Level = 90
	p.Base.MaxLevel = 90
	p.Stats = make([]float64, attributes.EndStatType)
	p.StatsByLabel = make(map[string][]float64)
	p.Params = make(map[string]int)
	p.Sets = make(map[keys.Set]int)
	if count > 0 {
		p.Sets[set] = count
	}
	p.SetParams = make(map[keys.Set]map[string]int)
	p.Weapon.Key = keys.DullBlade
	p.Weapon.Params = make(map[string]int)
	p.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	return p
}

func coreWithSet(t *testing.T, set keys.Set, partySize int) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(characterProfile(set, 4)); err != nil {
		t.Fatal(err)
	}
	for range partySize - 1 {
		if _, err := c.AddChar(characterProfile(keys.NoSet, 0)); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c
}

func newEnemy(c *core.Core) *enemy.Enemy {
	return enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{R: 1},
	})
}

func closeTo(got, want float64) bool {
	return math.Abs(got-want) < 1e-12
}

func TestDisenchantmentInDeepShadow(t *testing.T) {
	c := coreWithSet(t, keys.DisenchantmentInDeepShadow, 1)
	char := c.Player.ByIndex(0)

	if got := char.Stat(attributes.ATKP); got != 0.18 {
		t.Fatalf("2pc ATK%% = %v, want 0.18", got)
	}
	if got := char.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagSuperconductDamage}); got != 0.8 {
		t.Fatalf("Superconduct bonus = %v, want 0.8", got)
	}
	if got := char.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}); got != 0.4 {
		t.Fatalf("Star Superconduct bonus = %v, want 0.4", got)
	}

	target := newEnemy(c)
	trigger := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}}
	c.Events.Emit(event.OnStarSuperconduct, target, trigger)
	atk := &info.AttackEvent{}
	char.ApplyAttackMods(atk, target)
	if got := atk.Snapshot.Stats[attributes.CR]; got != 0.16 {
		t.Fatalf("CR against Star Superconduct target = %v, want 0.16", got)
	}
}

func TestGlacierAndSnowfield(t *testing.T) {
	c := coreWithSet(t, keys.GlacierAndSnowfield, 1)
	char := c.Player.ByIndex(0)

	if got := char.Stat(attributes.CryoP); got != 0.15 {
		t.Fatalf("2pc Cryo DMG%% = %v, want 0.15", got)
	}
	if got := char.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagSuperconductDamage}); got != 1 {
		t.Fatalf("Superconduct bonus = %v, want 1", got)
	}
	if got := char.ReactBonus(info.AttackInfo{Amped: true, AmpType: info.ReactionTypeMelt}); got != 0.15 {
		t.Fatalf("Melt bonus = %v, want 0.15", got)
	}
	if got := char.ReactBonus(info.AttackInfo{Amped: true, AmpType: info.ReactionTypeVaporize}); got != 0 {
		t.Fatalf("Vaporize bonus = %v, want 0", got)
	}

	c.Events.Emit(event.OnBurst)
	if got := char.Stat(attributes.CryoP); !closeTo(got, 0.45) {
		t.Fatalf("Cryo DMG%% after Burst = %v, want 0.45", got)
	}
	c.F = 10*60 + 1
	if got := char.Stat(attributes.CryoP); got != 0.15 {
		t.Fatalf("Cryo DMG%% after expiry = %v, want 0.15", got)
	}
}

func TestBloodRedProof(t *testing.T) {
	c := coreWithSet(t, keys.BloodRedProof, 1)
	char := c.Player.ByIndex(0)
	trigger := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}}
	c.Events.Emit(event.OnStarDiffusion, newEnemy(c), trigger)

	if got := char.Stat(attributes.ATKP); got != 0.18 {
		t.Fatalf("2pc ATK%% = %v, want 0.18", got)
	}
	if got := char.Stat(attributes.CR); !closeTo(got, 0.21) {
		t.Fatalf("triggered CR = %v, want 0.21", got)
	}
	if got := char.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionCryo}); got != 0.4 {
		t.Fatalf("Star Diffusion bonus = %v, want 0.4", got)
	}

	c.F = 10*60 + 1
	if got := char.Stat(attributes.CR); !closeTo(got, 0.05) {
		t.Fatalf("expired CR = %v, want 0.05", got)
	}
	if got := char.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionAnemo}); got != 0 {
		t.Fatalf("expired Star Diffusion bonus = %v, want 0", got)
	}
}

func TestArtifact15048(t *testing.T) {
	c := coreWithSet(t, keys.Artifact15048, 2)
	holder := c.Player.ByIndex(0)
	teammate := c.Player.ByIndex(1)
	trigger := &info.AttackEvent{Info: info.AttackInfo{
		ActorIndex: 0,
		AttackTag:  attacks.AttackTagReactionStarDiffusionAnemo,
	}}
	c.Events.Emit(event.OnStarDiffusion, newEnemy(c), trigger)

	if got := holder.Stat(attributes.ATKP); !closeTo(got, 0.30) {
		t.Fatalf("2pc plus triggered ATK%% = %v, want 0.30", got)
	}
	if got := teammate.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionCryo}); got != 0.5 {
		t.Fatalf("team Star Diffusion bonus = %v, want 0.5", got)
	}
	if got := teammate.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}); got != 0.5 {
		t.Fatalf("team Star Superconduct bonus = %v, want 0.5", got)
	}

	c.F = 12*60 + 1
	if got := holder.Stat(attributes.ATKP); got != 0.18 {
		t.Fatalf("ATK%% after 4pc expiry = %v, want 0.18", got)
	}
	if got := teammate.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}); got != 0 {
		t.Fatalf("team bonus after expiry = %v, want 0", got)
	}
}
