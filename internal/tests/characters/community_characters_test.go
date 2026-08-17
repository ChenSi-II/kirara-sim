package characters

import (
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/alyosha"
	_ "github.com/genshinsim/gcsim/internal/characters/iansan"
	_ "github.com/genshinsim/gcsim/internal/characters/ifa"
	_ "github.com/genshinsim/gcsim/internal/characters/illuga"
	_ "github.com/genshinsim/gcsim/internal/characters/jahoda"
	_ "github.com/genshinsim/gcsim/internal/characters/kachina"
	_ "github.com/genshinsim/gcsim/internal/characters/lohen"
	_ "github.com/genshinsim/gcsim/internal/characters/nefer"
	_ "github.com/genshinsim/gcsim/internal/characters/odette"
	_ "github.com/genshinsim/gcsim/internal/characters/prune"
	_ "github.com/genshinsim/gcsim/internal/characters/sandrone"
	_ "github.com/genshinsim/gcsim/internal/characters/zibai"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func TestCommunityCharactersRuntime(t *testing.T) {
	characters := []keys.Char{
		keys.Alyosha, keys.Iansan, keys.Ifa, keys.Illuga, keys.Jahoda, keys.Kachina,
		keys.Lohen, keys.Nefer, keys.Odette, keys.Prune, keys.Sandrone, keys.Zibai,
	}
	for _, key := range characters {
		t.Run(key.String(), func(t *testing.T) {
			c, targets := makeCore(1)
			profile := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
			profile.Base.Cons = 6
			profile.Base.Ascension = 6
			idx, err := c.AddChar(profile)
			if err != nil {
				t.Fatalf("add character: %v", err)
			}
			c.Player.SetActive(idx)
			if err := c.Init(); err != nil {
				t.Fatalf("initialize core: %v", err)
			}
			c.Combat.DefaultTarget = targets[0].Key()

			hits := 0
			c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
				if args[1].(*info.AttackEvent).Info.ActorIndex == idx {
					hits++
				}
			}, "community-runtime-hits")

			if err := c.Player.Exec(action.ActionAttack, key, nil); err != nil {
				t.Fatalf("attack: %v", err)
			}
			for range 90 {
				advanceCoreFrame(c)
			}
			if err := c.Player.Exec(action.ActionSkill, key, nil); err != nil {
				t.Fatalf("skill: %v", err)
			}
			for range 180 {
				advanceCoreFrame(c)
			}
			c.Player.Chars()[idx].AddEnergy("community-runtime", 1000)
			if err := c.Player.Exec(action.ActionBurst, key, nil); err != nil {
				t.Fatalf("burst: %v", err)
			}
			for range 240 {
				advanceCoreFrame(c)
			}
			if hits == 0 {
				t.Error("expected at least one character damage event")
			}
		})
	}
}

func TestCommunityTransformationActions(t *testing.T) {
	tests := []struct {
		key    keys.Char
		setup  func(int)
		wait   int
		action action.Action
	}{
		{key: keys.Ifa, wait: 45, action: action.ActionAttack},
		{key: keys.Nefer, setup: func(_ int) {}, wait: 45, action: action.ActionCharge},
		{key: keys.Odette, wait: 70, action: action.ActionSkill},
		{key: keys.Zibai, wait: 7 * 60, action: action.ActionSkill},
	}
	for _, tc := range tests {
		t.Run(tc.key.String(), func(t *testing.T) {
			c, targets := makeCore(1)
			profile := testhelper.DefaultProfile(tc.key, testhelper.TestWeaponKey)
			profile.Base.Cons, profile.Base.Ascension = 6, 6
			idx, err := c.AddChar(profile)
			if err != nil {
				t.Fatal(err)
			}
			c.Player.SetActive(idx)
			if err := c.Init(); err != nil {
				t.Fatal(err)
			}
			c.Combat.DefaultTarget = targets[0].Key()
			if tc.key == keys.Nefer {
				c.Player.SetVerdantDew(1)
			}
			if err := c.Player.Exec(action.ActionSkill, tc.key, nil); err != nil {
				t.Fatalf("initial skill: %v", err)
			}
			for range tc.wait {
				advanceCoreFrame(c)
			}
			if err := c.Player.Exec(tc.action, tc.key, nil); err != nil {
				t.Fatalf("transformed action: %v", err)
			}
			for range 120 {
				advanceCoreFrame(c)
			}
		})
	}
}
