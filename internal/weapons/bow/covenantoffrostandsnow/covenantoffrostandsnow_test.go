package covenantoffrostandsnow

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func testCore(t *testing.T, refine int) (*core.Core, int, int) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))

	prof := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
	equipper, err := c.AddChar(prof)
	if err != nil {
		t.Fatal(err)
	}
	teammate, err := c.AddChar(prof)
	if err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(teammate)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	_, err = NewWeapon(c, c.Player.Chars()[equipper], info.WeaponProfile{Refine: refine})
	if err != nil {
		t.Fatal(err)
	}
	return c, equipper, teammate
}

func TestElementalSkillTriggerAndRefinement(t *testing.T) {
	for _, tc := range []struct {
		refine int
		wantEM float64
	}{
		{refine: 1, wantEM: 120},
		{refine: 5, wantEM: 240},
	} {
		t.Run(string(rune('0'+tc.refine)), func(t *testing.T) {
			c, equipper, teammate := testCore(t, tc.refine)
			char := c.Player.Chars()[equipper]
			baseEM := char.Stat(attributes.EM)

			c.Events.Emit(event.OnSkill)
			if got := char.Stat(attributes.EM) - baseEM; got != 0 {
				t.Fatalf("teammate skill triggered the passive: got EM bonus %v", got)
			}

			c.Player.SetActive(equipper)
			c.Events.Emit(event.OnSkill)
			if got := char.Stat(attributes.EM) - baseEM; got != tc.wantEM {
				t.Fatalf("unexpected EM bonus: got %v, want %v", got, tc.wantEM)
			}

			c.Player.SetActive(teammate)
			for range 12 * 60 {
				c.F++
				_ = c.Tick()
			}
			if got := char.Stat(attributes.EM) - baseEM; got != 0 {
				t.Fatalf("EM bonus did not expire: got %v", got)
			}
		})
	}
}
