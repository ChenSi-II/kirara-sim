package enemy_test

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
	_ "github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestCharacter()
	testhelper.RegisterTestWeapon()
}

func setupSpecialDamage(t *testing.T) (*core.Core, *enemy.Enemy) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	e := enemy.New(c, info.EnemyProfile{Level: 90, Pos: info.Coord{R: 1}, Resist: map[attributes.Element]float64{
		attributes.Cryo: .1, attributes.Anemo: .1,
	}})
	c.Combat.AddEnemy(e)
	if _, err := c.AddChar(testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.ByIndex(0).AddReactBonusMod(character.ReactBonusMod{
		Base:   modifier.NewBase("test-special-reaction-bonus", -1),
		Amount: func(info.AttackInfo) float64 { return .4 },
	})
	return c, e
}

func specialSnapshot() info.Snapshot {
	s := info.Snapshot{CharLvl: 90}
	s.Stats[attributes.BaseATK] = 2000
	s.Stats[attributes.BaseDEF] = 3000
	s.Stats[attributes.BaseHP] = 40000
	s.Stats[attributes.EM] = 1000
	s.Stats[attributes.CR] = 1
	s.Stats[attributes.CD] = .5
	return s
}

func assertSpecialDamage(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Fatalf("damage = %.12f, want %.12f", got, want)
	}
}

func TestDirectSpecialDamage(t *testing.T) {
	for _, tc := range []struct {
		name        string
		tag         attacks.AttackTag
		coefficient float64
	}{
		{"superconduct initial", attacks.AttackTagReactionStarSuperconduct, 1},
		{"superconduct one stack", attacks.AttackTagReactionStarSuperconduct, 1.45},
		{"superconduct twelve stacks", attacks.AttackTagReactionStarSuperconduct, 2},
		{"diffusion anemo", attacks.AttackTagReactionStarDiffusionAnemo, 1},
		{"diffusion cryo", attacks.AttackTagReactionStarDiffusionCryo, 1},
		{"lunar charged", attacks.AttackTagDirectLunarCharged, 3},
		{"lunar crystallize", attacks.AttackTagDirectLunarCrystallize, 1.6},
		{"lunar bloom", attacks.AttackTagDirectLunarBloom, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, scaling := range []string{"atk", "def", "hp", "em", "flat"} {
				t.Run(scaling, func(t *testing.T) {
					c, e := setupSpecialDamage(t)
					c.StarReactions.SuperconductActive = true
					c.StarReactions.SuperconductCoefficient = 2
					c.StarReactions.DiffusionActive = true
					c.StarReactions.DiffusionStacks = 6
					if tc.tag == attacks.AttackTagReactionStarSuperconduct {
						c.StarReactions.SuperconductCoefficient = tc.coefficient
					}
					ai := info.AttackInfo{
						ActorIndex: 0, Abil: "direct special test", AttackTag: tc.tag,
						Element: attributes.Cryo, Mult: 2, BaseDmgBonus: .2, FlatDmg: 100, Elevation: .25,
					}
					if tc.tag == attacks.AttackTagReactionStarDiffusionAnemo {
						ai.Element = attributes.Anemo
					}
					base := 4000.0
					switch scaling {
					case "def":
						ai.UseDef, base = true, 6000
					case "hp":
						ai.UseHP, base = true, 80000
					case "em":
						ai.UseEM, base = true, 2000
					case "flat":
						ai.Mult, base = 0, 0
					}
					// Talent base, reaction/EM bonuses, flat addition, elevation,
					// 10% resistance and guaranteed 50% crit damage, each once.
					want := (base*tc.coefficient*3.4*1.2 + 100) * 1.25 * .9 * 1.5
					snap := specialSnapshot()
					assertSpecialDamage(t, e.HandleAttack(&info.AttackEvent{Info: ai, Snapshot: snap}), want)

					// Neither ordinary damage bonuses nor defense/level changes
					// can alter direct special damage.
					e.Level = 200
					e.AddDefMod(info.DefMod{Base: modifier.NewBase("test-defense", -1), Value: -.5})
					snap.CharLvl = 1
					snap.Stats[attributes.DmgP] = 5
					snap.Stats[attributes.CryoP] = 5
					snap.Stats[attributes.AnemoP] = 5
					assertSpecialDamage(t, e.HandleAttack(&info.AttackEvent{Info: ai, Snapshot: snap}), want)
				})
			}
		})
	}
}

func TestStarDiffusionReactionDamageIsCalculatedOnce(t *testing.T) {
	c, e := setupSpecialDamage(t)
	c.StarReactions.Enabled = true
	e.SetAuraDurability(info.ReactionModKeyCryo, 25, 0)
	contributions, hits := 0, 0
	c.Events.Subscribe(event.OnStarReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if !atk.Info.IsStarDiffusionReaction || atk.Info.IsDirectStarDamage() {
			t.Fatal("reaction contribution misidentified as a talent attack")
		}
		contributions++
		atk.Snapshot = specialSnapshot()
		atk.Info.BaseDmgBonus, atk.Info.FlatDmg, atk.Info.Elevation = .2, 100, .25
	}, "test-diffusion-contribution")
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagReactionStarDiffusionAnemo {
			return
		}
		hits++
		if !atk.Info.IsStarDiffusionReaction || atk.Info.Mult != 0 {
			t.Fatal("combined reaction packet must retain its explicit marker")
		}
		// The sole contributor has weight 60%. Crit and elevation are
		// already inside that contribution; the final packet adds resistance.
		want := (.75*combat.CalcReactionBaseDmg(90)*3.4*1.2 + 100) * 1.25 * 1.5 * .6 * .9
		assertSpecialDamage(t, args[2].(float64), want)
	}, "test-diffusion-final-damage")
	c.QueueAttackWithSnap(info.AttackInfo{
		ActorIndex: 0, Element: attributes.Anemo, Durability: 25,
		AttackTag: attacks.AttackTagElementalArt,
	}, info.Snapshot{}, combat.NewSingleTargetHit(e.Key()), 0)
	for range 2 {
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	if contributions != 1 || hits != 1 {
		t.Fatalf("contributions=%d, reaction hits=%d; want one each", contributions, hits)
	}
}
