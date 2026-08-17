package common

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func TestSubscribeOwnerReactionsFiltersActorAndWorksOffField(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))

	profile := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
	holderIndex, err := c.AddChar(profile)
	if err != nil {
		t.Fatal(err)
	}
	teammateIndex, err := c.AddChar(profile)
	if err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(teammateIndex)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}

	holder := c.Player.Chars()[holderIndex]
	procs := 0
	SubscribeOwnerReactions(c, holder, "owner-reaction-test", func(*info.AttackEvent) {
		procs++
	})

	c.Events.Emit(event.OnOverload, nil, &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: teammateIndex},
	})
	if procs != 0 {
		t.Fatalf("teammate reaction triggered holder passive %d times", procs)
	}

	c.Events.Emit(event.OnOverload, nil, &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: holderIndex},
	})
	if procs != 1 {
		t.Fatalf("off-field holder reaction triggered passive %d times, want 1", procs)
	}
}
