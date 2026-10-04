package varka

import (
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/fischl"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func timingCore(t *testing.T, hitlag bool) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, EnableHitlag: hitlag, DefHalt: true, Delays: info.Delays{Swap: 12}})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
	c.Combat.AddEnemy(target)
	for _, key := range []keys.Char{keys.Varka, keys.Fischl} {
		p := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
		p.Base.Ascension = 6

		if _, err := c.AddChar(p); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.ByIndex(1).IsHexerei = true
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Combat.DefaultTarget = target.Key()
	return c
}

// This is a C0, two-Hexerei, single-target calibration including hitlag.
// It exercises real animation locks, damage-driven cooldown reduction and swap
// delay. The aggregate observation does not certify any individual hitmark.
func TestUserRotationTiming(t *testing.T) {
	c := timingCore(t, true)
	ch := c.Player.ByIndex(0).Character.(*char)
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
			t.Fatalf("%s not ready within rotation budget", a)
		}
	}
	// e4azaEazaz2aEaza; the four z inputs must remain ordinary state CAs.
	for _, step := range "eaaaazaEazazaaEaza" {
		a := action.ActionAttack
		switch step {
		case 'e', 'E':
			a = action.ActionSkill
		case 'z':
			a = action.ActionCharge
		}
		wait(a, keys.Varka)
		if step == 'E' && !ch.StatusIsActive(skillKey) {
			t.Fatal("special E fell outside Sturm und Drang")
		}
		if step == 'z' && ch.fourWindsCharges() > 0 {
			t.Fatal("rotation unexpectedly consumed E charge for Azure Devour")
		}
		if err := c.Player.Exec(a, keys.Varka, nil); err != nil {
			t.Fatal(err)
		}
	}
	wait(action.ActionSwap, keys.Fischl)
	if c.F != 750 {
		t.Fatalf("rotation ready to swap at %d, want 750f (12.5s)", c.F)
	}
	if err := c.Player.Exec(action.ActionSwap, keys.Fischl, nil); err != nil {
		t.Fatal(err)
	}
	for c.Player.Active() == 0 && c.F < 800 {
		advance()
	}
	if c.F != 762 {
		t.Fatalf("swap complete at %d, want 762f (12.7s)", c.F)
	}
}
