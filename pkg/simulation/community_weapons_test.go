package simulation_test

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	_ "github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
}

func TestCommunityWeaponRegistrationAndAddChar(t *testing.T) {
	weapons := []keys.Weapon{
		keys.HereticsMoltenBlade,
		keys.Emberwell,
		keys.LightbearingMoonshard,
		keys.WhitelakeFrostfeather,
		keys.ExaiphanesBlade,
		keys.FlameForgedInsight,
		keys.ForgedByTheGoldenMelody,
		keys.BladeOfAtonement,
		keys.ATeaspoonOfTranscendence,
		keys.Frostbreath,
		keys.SongOfTheVigil,
		keys.DisasterAndRemorse,
		keys.ClashOfKings,
		keys.EchoesOfTheHeart,
		keys.JadeVista,
		keys.CovenantOfFrostAndSnow,
	}
	for _, weapon := range weapons {
		t.Run(weapon.String(), func(t *testing.T) {
			c, err := core.New(core.Opt{Seed: 1})
			if err != nil {
				t.Fatal(err)
			}
			c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
			profile := testhelper.DefaultProfile(testhelper.TestCharKey, weapon)
			profile.Weapon.Level = 90
			profile.Weapon.MaxLevel = 90
			profile.Weapon.Refine = 1
			if _, err := c.AddChar(profile); err != nil {
				t.Fatalf("weapon cannot be loaded by Core.AddChar: %v", err)
			}
		})
	}
}
