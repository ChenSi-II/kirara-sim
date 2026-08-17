package whitelakefrostfeather

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func TestRefinementValues(t *testing.T) {
	if got := atkPerStack(1); got != 0.08 {
		t.Fatalf("R1 ATK per stack: got %v", got)
	}
	if got := atkPerStack(5); got != 0.16 {
		t.Fatalf("R5 ATK per stack: got %v", got)
	}
}

func TestSkillHitOwnershipICDStacksAndExpiry(t *testing.T) {
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
	baseATK := holder.Stat(attributes.ATKP)
	if _, err := NewWeapon(c, holder, info.WeaponProfile{Refine: 1}); err != nil {
		t.Fatal(err)
	}
	emit := func(actor int, tag attacks.AttackTag) {
		c.Events.Emit(event.OnEnemyDamage, nil, &info.AttackEvent{
			Info: info.AttackInfo{ActorIndex: actor, AttackTag: tag},
		})
	}

	emit(teammateIndex, attacks.AttackTagElementalArt)
	emit(holderIndex, attacks.AttackTagNormal)
	if got := holder.Stat(attributes.ATKP) - baseATK; got != 0 {
		t.Fatalf("wrong actor or attack type triggered %v ATK", got)
	}

	emit(holderIndex, attacks.AttackTagElementalArt)
	emit(holderIndex, attacks.AttackTagElementalArt)
	if got := holder.Stat(attributes.ATKP) - baseATK; math.Abs(got-0.08) > 1e-9 {
		t.Fatalf("0.1s ICD failed: got %v ATK", got)
	}
	c.F += 6
	emit(holderIndex, attacks.AttackTagElementalArt)
	if got := holder.Stat(attributes.ATKP) - baseATK; math.Abs(got-0.16) > 1e-9 {
		t.Fatalf("independent stack failed: got %v ATK", got)
	}
	c.F = 8 * 60
	if got := holder.Stat(attributes.ATKP) - baseATK; math.Abs(got-0.08) > 1e-9 {
		t.Fatalf("first stack did not expire independently: got %v ATK", got)
	}
}
