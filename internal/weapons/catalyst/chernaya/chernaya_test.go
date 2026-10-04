package chernaya

import (
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/testhelper"
	"math"
	"testing"
)

func init() { testhelper.RegisterTestCharacter(); testhelper.RegisterTestWeapon() }
func TestRefinementsTeamReactionsResetAndExpiry(t *testing.T) {
	for _, r := range []int{1, 5} {
		c, err := core.New(core.Opt{Seed: 1})
		if err != nil {
			t.Fatal(err)
		}
		c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
		p := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
		if _, err = c.AddChar(p); err != nil {
			t.Fatal(err)
		}
		if _, err = c.AddChar(p); err != nil {
			t.Fatal(err)
		}
		if err = c.Init(); err != nil {
			t.Fatal(err)
		}
		holder := c.Player.ByIndex(0)
		em, cd := holder.Stat(attributes.EM), holder.Stat(attributes.CD)
		wi, _ := NewWeapon(c, holder, info.WeaponProfile{Refine: r})
		w := wi.(*Weapon)
		w.Init()
		if math.Abs(holder.Stat(attributes.CD)-cd-(.18+.06*float64(r))) > 1e-9 {
			t.Fatal("permanent CD")
		}
		c.Player.SetActive(1)
		c.Events.Emit(event.OnSkill)
		if w.until != 0 {
			t.Fatal("teammate skill activated weapon")
		}
		c.Player.SetActive(0)
		c.Events.Emit(event.OnSkill)
		base := 36 + 12*float64(r)
		if holder.Stat(attributes.EM)-em != base {
			t.Fatal("skill EM")
		}
		c.Player.SetActive(1)
		for i := 0; i < 6; i++ {
			c.Events.Emit(event.OnStarSuperconduct, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}})
		}
		want := base + 3*(18+6*float64(r))
		if w.stacks != 3 || holder.Stat(attributes.EM)-em != want {
			t.Fatal("off-field team reaction stacking")
		}
		checkCD := func(tag attacks.AttackTag, want float64) {
			a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: tag}}
			holder.ApplyAttackMods(a, nil)
			if math.Abs(a.Snapshot.Stats[attributes.CD]-want) > 1e-9 {
				t.Fatalf("conditional CD %v want %v", a.Snapshot.Stats[attributes.CD], want)
			}
		}
		checkCD(attacks.AttackTagReactionStarSuperconduct, 0)
		c.StarReactions.SuperconductActive = true
		checkCD(attacks.AttackTagReactionStarSuperconduct, 3*(.09+.03*float64(r)))
		checkCD(attacks.AttackTagElementalArt, 0)
		c.Player.SetActive(0)
		c.Events.Emit(event.OnSkill)
		if w.stacks != 0 || holder.Stat(attributes.EM)-em != base {
			t.Fatal("recast failed to reset")
		}
		c.F = 1200
		if holder.Stat(attributes.EM) != em {
			t.Fatal("EM did not expire")
		}
		checkCD(attacks.AttackTagReactionStarSuperconduct, 0)
	}
}
