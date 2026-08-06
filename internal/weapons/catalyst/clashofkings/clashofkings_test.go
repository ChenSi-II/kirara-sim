package clashofkings

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
	if atkBonus(1) != 0.20 || emBonus(1) != 100 {
		t.Fatalf("unexpected R1 values")
	}
	if atkBonus(5) != 0.40 || emBonus(5) != 200 {
		t.Fatalf("unexpected R5 values")
	}
}

func TestSkillOwnerChargedExtensionAndWrongAttack(t *testing.T) {
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
	baseEM := holder.Stat(attributes.EM)
	if _, err := NewWeapon(c, holder, info.WeaponProfile{Refine: 1}); err != nil {
		t.Fatal(err)
	}

	c.Events.Emit(event.OnSkill)
	if holder.StatusIsActive(buffStatusKey) {
		t.Fatal("teammate skill triggered holder passive")
	}
	c.Player.SetActive(holderIndex)
	c.Events.Emit(event.OnSkill)
	if got := holder.Stat(attributes.ATKP) - baseATK; math.Abs(got-0.20) > 1e-9 {
		t.Fatalf("ATK bonus: got %v", got)
	}
	if got := holder.Stat(attributes.EM) - baseEM; got != 100 {
		t.Fatalf("EM bonus: got %v", got)
	}

	emit := func(actor int, tag attacks.AttackTag) {
		c.Events.Emit(event.OnEnemyDamage, nil, &info.AttackEvent{
			Info: info.AttackInfo{ActorIndex: actor, AttackTag: tag},
		})
	}
	emit(holderIndex, attacks.AttackTagNormal)
	if got := holder.StatusDuration(buffStatusKey); got != 6*60 {
		t.Fatalf("normal attack extended duration to %d", got)
	}
	emit(teammateIndex, attacks.AttackTagExtra)
	if got := holder.StatusDuration(buffStatusKey); got != 6*60 {
		t.Fatalf("teammate charged attack extended duration to %d", got)
	}
	emit(holderIndex, attacks.AttackTagExtra)
	if got := holder.StatusDuration(buffStatusKey); got != 12*60 {
		t.Fatalf("charged attack extension: got duration %d", got)
	}
	emit(holderIndex, attacks.AttackTagExtra)
	if got := holder.StatusDuration(buffStatusKey); got != 12*60 {
		t.Fatalf("duration extended more than once: got %d", got)
	}
}
