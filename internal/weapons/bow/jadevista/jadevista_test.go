package jadevista

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func TestPartyElementBonusesAndRefinement(t *testing.T) {
	for _, tc := range []struct {
		refine  int
		wantEM  float64
		wantATK float64
	}{
		{refine: 1, wantEM: 128, wantATK: 0.12},
		{refine: 5, wantEM: 256, wantATK: 0.24},
	} {
		t.Run(string(rune('0'+tc.refine)), func(t *testing.T) {
			c, err := core.New(core.Opt{Seed: 1})
			if err != nil {
				t.Fatal(err)
			}
			c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))

			indices := make([]int, 0, 4)
			for range 4 {
				prof := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
				idx, err := c.AddChar(prof)
				if err != nil {
					t.Fatal(err)
				}
				indices = append(indices, idx)
			}
			c.Player.SetActive(indices[0])
			if err := c.Init(); err != nil {
				t.Fatal(err)
			}
			// UpdateBaseStats restores the fake character's catalog element, so
			// assign this synthetic party after initialization.
			elements := []attributes.Element{
				attributes.Cryo,
				attributes.Cryo,
				attributes.Cryo,
				attributes.Pyro,
			}
			for i, element := range elements {
				c.Player.Chars()[indices[i]].Base.Element = element
			}

			char := c.Player.Chars()[indices[0]]
			baseEM := char.Stat(attributes.EM)
			baseATKP := char.Stat(attributes.ATKP)
			weapon, err := NewWeapon(c, char, info.WeaponProfile{Refine: tc.refine})
			if err != nil {
				t.Fatal(err)
			}
			if err := weapon.Init(); err != nil {
				t.Fatal(err)
			}

			if got := char.Stat(attributes.EM) - baseEM; got != tc.wantEM {
				t.Fatalf("unexpected EM bonus: got %v, want %v", got, tc.wantEM)
			}
			if got := char.Stat(attributes.ATKP) - baseATKP; math.Abs(got-tc.wantATK) > 1e-9 {
				t.Fatalf("unexpected ATK%% bonus: got %v, want %v", got, tc.wantATK)
			}
		})
	}
}
