package vesna

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/internal/characters/vodyanitsa"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func setupVesna(t *testing.T, cons int) (*core.Core, *char) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
	c.Combat.AddEnemy(target)
	for _, key := range []keys.Char{keys.Vesna, keys.Vodyanitsa} {
		p := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
		p.Base.Cons = cons
		p.Base.Ascension = 6
		p.Talents.Skill = 10
		if _, err := c.AddChar(p); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = target.Key()
	return c, c.Player.ByIndex(0).Character.(*char)
}

func TestVesnaSkillStagesUseSourceMultipliers(t *testing.T) {
	c, ch := setupVesna(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	for range 60 {
		c.F++
		c.Tick()
	}
	var hits []float64
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex == 0 {
			hits = append(hits, a.Info.Mult)
		}
	}, "test-stage-hits")
	for stage, want := range [][]float64{{.72}, {1.08, 2.016 * 1.1}, {.8064 * 1.2, .8064 * 1.2, .8064 * 1.2, .8064 * 1.2, 2.8224 * 1.2}} {
		hits = nil
		ch.magic = 5
		if _, err := ch.Skill(nil); err != nil {
			t.Fatal(err)
		}
		for range 76 {
			c.F++
			c.Tick()
		}
		if len(hits) != len(want) {
			t.Fatalf("stage %d hit count %d, want %d", stage+1, len(hits), len(want))
		}
		for i, v := range want {
			if math.Abs(hits[i]-v) > 1e-8 {
				t.Fatalf("stage %d hit %d: %v, want %v", stage+1, i, hits[i], v)
			}
		}
	}
}

func TestVesnaIndependentComposureAndC2(t *testing.T) {
	c, ch := setupVesna(t, 2)
	base := ch.Stat(attributes.ATKP)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if math.Abs(ch.Stat(attributes.ATKP)-base-.40) > 1e-8 {
		t.Fatal("C2 should grant 40% ATK")
	}
	c.F = 600
	ch.addComposure()
	c.F = 1200
	if ch.activeComposure() != 1 || math.Abs(ch.Stat(attributes.ATKP)-base) > 1e-8 {
		t.Fatal("old composure layers did not expire independently")
	}
	ch.endSpiritbladeArmament()
	if ch.activeComposure() != 1 {
		t.Fatal("natural armament end should not clear composure")
	}
	c.Events.Emit(event.OnCharacterSwap, 0, 1)
	if ch.activeComposure() != 0 {
		t.Fatal("swap did not clear composure")
	}
}

func TestVodyanitsaExtendsVesnaRadiance(t *testing.T) {
	c, ch := setupVesna(t, 0)
	c.Events.Emit(event.OnStarDiffusion, nil, &info.AttackEvent{})
	if ch.StatusDuration(radianceKey) != 480 {
		t.Fatal("base radiance should last 8s")
	}
	singer := c.Player.ByIndex(1)
	singer.AddStatus(vodyanitsa.SongKey, 1000, false)
	c.Events.Emit(event.OnStarDiffusion, nil, &info.AttackEvent{})
	if ch.StatusDuration(radianceKey) != 720 {
		t.Fatal("song should extend new radiance to 12s")
	}
	singer.DeleteStatus(vodyanitsa.SongKey)
	c.F = 1
	c.Events.Emit(event.OnStarDiffusion, nil, &info.AttackEvent{})
	if ch.StatusDuration(radianceKey) != 480 {
		t.Fatal("expired song still extended radiance")
	}
}
