package lightbearingmoonshard

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
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func TestGeneratedRegistration(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))

	prof := testhelper.DefaultProfile(testhelper.TestCharKey, keys.LightbearingMoonshard)
	prof.Weapon.Level = 90
	prof.Weapon.MaxLevel = 90
	prof.Weapon.Refine = 1
	if _, err := c.AddChar(prof); err != nil {
		t.Fatalf("generated weapon key/catalog/registration is incomplete: %v", err)
	}
}

func TestBonusesTriggerOwnershipAndRefinement(t *testing.T) {
	for _, tc := range []struct {
		refine  int
		wantDEF float64
		wantLCR float64
	}{
		{refine: 1, wantDEF: 0.20, wantLCR: 0.64},
		{refine: 5, wantDEF: 0.40, wantLCR: 1.28},
	} {
		t.Run(string(rune('0'+tc.refine)), func(t *testing.T) {
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

			char := c.Player.Chars()[equipper]
			baseDEFP := char.Stat(attributes.DEFP)
			_, err = NewWeapon(c, char, info.WeaponProfile{Refine: tc.refine})
			if err != nil {
				t.Fatal(err)
			}
			if got := char.Stat(attributes.DEFP) - baseDEFP; math.Abs(got-tc.wantDEF) > 1e-9 {
				t.Fatalf("unexpected DEF%% bonus: got %v, want %v", got, tc.wantDEF)
			}

			directLCR := info.AttackInfo{AttackTag: attacks.AttackTagDirectLunarCrystallize}
			reactionLCR := info.AttackInfo{AttackTag: attacks.AttackTagReactionLunarCrystallize}
			nonLCR := info.AttackInfo{AttackTag: attacks.AttackTagElementalArt}

			c.Events.Emit(event.OnSkill)
			if got := char.ReactBonus(directLCR); got != 0 {
				t.Fatalf("teammate skill triggered the passive: got Lunar-Crystallize bonus %v", got)
			}

			c.Player.SetActive(equipper)
			c.Events.Emit(event.OnSkill)
			for name, atk := range map[string]info.AttackInfo{
				"direct":   directLCR,
				"reaction": reactionLCR,
			} {
				if got := char.ReactBonus(atk); math.Abs(got-tc.wantLCR) > 1e-9 {
					t.Fatalf("%s Lunar-Crystallize bonus: got %v, want %v", name, got, tc.wantLCR)
				}
			}
			if got := char.ReactBonus(nonLCR); got != 0 {
				t.Fatalf("non-Lunar-Crystallize attack received bonus %v", got)
			}

			c.Player.SetActive(teammate)
			for range 5 * 60 {
				c.F++
				_ = c.Tick()
			}
			if got := char.ReactBonus(directLCR); got != 0 {
				t.Fatalf("Lunar-Crystallize bonus did not expire: got %v", got)
			}
		})
	}
}
