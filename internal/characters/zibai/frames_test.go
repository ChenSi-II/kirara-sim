package zibai

import (
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/fischl"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() { testhelper.RegisterTestWeapon() }

func rotationCore(t *testing.T) (*core.Core, *char) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, IgnoreBurstEnergy: true, Delays: info.Delays{Swap: 12}})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
	c.Combat.AddEnemy(target)
	for _, key := range []keys.Char{keys.Zibai, keys.Fischl} {
		p := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
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

func TestUserRotationAnimationBudget(t *testing.T) {
	c, ch := rotationCore(t)
	normals, strides := 0, 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.ActorIndex != 0 {
			return
		}
		if a.Info.Abil == "Spirit Steed's Stride 2" {
			strides++
		}
		if a.Info.Abil == "Lunar Phase Normal 1-1" {
			normals++
		}
	}, "rotation-hits")
	advance := func() {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	wait := func(a action.Action, key keys.Char) {
		for c.Player.ReadyCheck(a, key, nil) != nil && c.F < 1200 {
			advance()
		}
		if c.F >= 1200 {
			t.Fatalf("%s unavailable", a)
		}
	}
	// Isolate animation fitting from the unspecified team's Lunar Crystallize
	// uptime. Resource supply is controlled ONLY by this fixture, not by the
	// runtime implementation; real rotations can wait for phase/energy.
	for _, step := range "eaqaEaaaaEaaaaEaaaaE" {
		a := action.ActionAttack
		switch step {
		case 'e':
			a = action.ActionSkill
		case 'E':
			a = action.ActionSkill
			ch.phase = 100
		case 'q':
			a = action.ActionBurst
		}
		wait(a, keys.Zibai)
		if err := c.Player.Exec(a, keys.Zibai, nil); err != nil {
			t.Fatal(err)
		}
	}
	wait(action.ActionSwap, keys.Fischl)
	if c.F != 750 {
		t.Fatalf("swap ready at %d, want 750f", c.F)
	}
	if normals != 5 || strides != 4 {
		t.Fatalf("got %d N1s and %d strides, want 5 and 4", normals, strides)
	}
	if err := c.Player.Exec(action.ActionSwap, keys.Fischl, nil); err != nil {
		t.Fatal(err)
	}
	for c.Player.Active() == 0 && c.F < 800 {
		advance()
	}
	if c.F != 762 {
		t.Fatalf("swap complete at %d, want 762f", c.F)
	}
}

func TestRotationEstimateDoesNotBypassPhaseRequirement(t *testing.T) {
	_, ch := rotationCore(t)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	if ready, _ := ch.ActionReady(action.ActionSkill, nil); ready {
		t.Fatal("stride ready without phase")
	}
	ch.addPhase(70)
	if ready, _ := ch.ActionReady(action.ActionSkill, nil); !ready {
		t.Fatal("stride unavailable with enough phase")
	}
}
