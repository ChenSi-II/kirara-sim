package vodyanitsa

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func setupVodyanitsa(t *testing.T, cons int) (*core.Core, *char, *enemy.Enemy) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
	c.Combat.AddEnemy(target)
	p := testhelper.DefaultProfile(keys.Vodyanitsa, testhelper.TestWeaponKey)
	p.Base.Cons = cons
	p.Base.Ascension = 6
	p.Talents.Skill = 10
	p.Stats[attributes.HP] = 50000
	if _, err := c.AddChar(p); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = target.Key()
	return c, c.Player.ByIndex(0).Character.(*char), target
}

func vodyanitsaNear(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestUpdatedSkillAndSongResources(t *testing.T) {
	c, ch, target := setupVodyanitsa(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if ch.soloStacks != 25 || ch.concertStacks != 10 {
		t.Fatalf("song stacks = %d/%d", ch.soloStacks, ch.concertStacks)
	}
	for range 38 {
		c.F++
		c.Tick()
	}
	if !target.ResistModIsActive("vodyanitsa-microphone-hydro-res") || !target.ResistModIsActive("vodyanitsa-microphone-cryo-res") {
		t.Fatal("initial skill hit did not shred both resistances")
	}
	// At level 10 the new shred is 30%, formerly 26%.
	vodyanitsaNear(t, skillResist[ch.TalentLvlSkill()], .30)
	// Resources survive the 16s song and expire at 30s, independently of it.
	c.F = 17 * 60
	ch.soloStacks = 25
	atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Element: attributes.Hydro, AttackTag: attacks.AttackTagNormal}}
	c.Events.Emit(event.OnEnemyHit, target, atk)
	if ch.soloStacks != 24 || atk.Info.FlatDmg <= 0 {
		t.Fatal("resources expired with the song")
	}
	c.F = 30 * 60
	c.Events.Emit(event.OnEnemyHit, target, atk)
	if ch.soloStacks != 24 {
		t.Fatal("resources remained usable after 30s")
	}
}

func TestSongC1AndC6(t *testing.T) {
	c, ch, _ := setupVodyanitsa(t, 6)
	ch.AddStatus(microphoneKey, 1000, false)
	ch.skillSrc = 0
	ch.microphoneHeal(0, ch.TalentLvlSkill())()
	ally := c.Player.ByIndex(1)
	vodyanitsaNear(t, ally.Stat(attributes.ATK), .008*ch.MaxHP())
	if ally.StatusExpiry("vodyanitsa-c1") != 300 {
		t.Fatal("C1 should last 5 seconds")
	}
	c.F = 60
	c.Events.Emit(event.OnHeal, &info.HealInfo{Caller: 0, Message: "weapon heal"}, 0, 100.0, 0.0, 100.0)
	if ally.StatusExpiry("vodyanitsa-c1") != 300 {
		t.Fatal("unrelated healing refreshed C1")
	}
	atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, Element: attributes.Hydro, AttackTag: attacks.AttackTagNormal}}
	ally.ApplyAttackMods(atk, nil)
	vodyanitsaNear(t, atk.Snapshot.Stats[attributes.DmgP], .60)
}

func TestFlowingVortexConversionAndC2FollowsActiveCharacter(t *testing.T) {
	c, ch, target := setupVodyanitsa(t, 2)
	c.StarReactions.DiffusionStacks = 1
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if !ch.flowingVortex || !target.ResistModIsActive("vodyanitsa-flowing-vortex-anemo-res") {
		t.Fatal("existing vortex was not converted")
	}
	c.Events.Emit(event.OnStarDiffusionVortex, target, true)
	if ch.flowingVortex || ch.StatusExpiry(recentVortexKey) != 300 {
		t.Fatal("detonation did not open a 5s recent-vortex window")
	}
	ch.microphoneC2Buff()
	c.Player.SetActive(1)
	atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1, AttackTag: attacks.AttackTagReactionStarDiffusionAnemo}}
	c.Events.Emit(event.OnStarReactionAttack, target, atk)
	vodyanitsaNear(t, atk.Snapshot.Stats[attributes.CD], .60)
	atk.Info.ActorIndex = 0
	atk.Snapshot.Stats[attributes.CD] = 0
	c.Events.Emit(event.OnStarReactionAttack, target, atk)
	vodyanitsaNear(t, atk.Snapshot.Stats[attributes.CD], 0)
	c.F = 300
	ch.microphoneC2Buff()
	atk.Info.ActorIndex = 1
	atk.Info.Element = attributes.Hydro
	atk.Info.AttackTag = attacks.AttackTagNormal
	c.Player.ByIndex(1).ApplyAttackMods(atk, target)
	vodyanitsaNear(t, atk.Snapshot.Stats[attributes.CD], .50)
}
