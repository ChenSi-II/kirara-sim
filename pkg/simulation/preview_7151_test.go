package simulation_test

import (
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
	"testing"
)

func TestPreview7151ActualActionExecution(t *testing.T) {
	for _, key := range []keys.Char{keys.Mitya, keys.Valeriy} {
		for _, cons := range []int{0, 6} {
			c, err := core.New(core.Opt{Seed: 1})
			if err != nil {
				t.Fatal(err)
			}
			c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
			target := enemy.New(c, info.EnemyProfile{Level: 90, Resist: map[attributes.Element]float64{}, Pos: info.Coord{R: 1}})
			c.Combat.AddEnemy(target)
			weapon := keys.Chernaya
			if key == keys.Valeriy {
				weapon = keys.FavoniusSword
			}
			p := testhelper.DefaultProfile(key, weapon)
			p.Base.Cons = cons
			p.Base.Ascension = 6
			p.Weapon.Level = 90
			p.Weapon.MaxLevel = 90
			p.Weapon.Refine = 1
			if _, err = c.AddChar(p); err != nil {
				t.Fatal(err)
			}
			c.Player.SetActive(0)
			if err = c.Init(); err != nil {
				t.Fatal(err)
			}
			c.Combat.DefaultTarget = target.Key()
			tick := func(n int) {
				for i := 0; i < n; i++ {
					c.F++
					if err := c.Tick(); err != nil {
						t.Fatal(err)
					}
				}
			}
			for _, a := range []action.Action{action.ActionSkill, action.ActionBurst, action.ActionCharge, action.ActionAttack} {
				c.Player.ByIndex(0).AddEnergy("test", 100)
				if err := c.Player.Exec(a, key, map[string]int{"hold": 1, "duration": 300}); err != nil {
					t.Fatalf("%s C%d %s: %v", key, cons, a, err)
				}
				tick(320)
			}
			if c.Combat.TotalDamage <= 0 {
				t.Fatal("no damage")
			}
		}
	}
}
