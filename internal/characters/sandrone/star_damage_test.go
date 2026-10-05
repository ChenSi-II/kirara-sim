package sandrone

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func TestStellarTalentDamageUsesReactionFormula(t *testing.T) {
	for _, mode := range []string{"superconduct", "diffusion"} {
		for _, ability := range []string{"skill", "burst", "ray"} {
			for _, cons := range []int{0, 1} {
				name := mode + "/" + ability + "/C0"
				if cons == 1 {
					name = mode + "/" + ability + "/C1"
				}
				t.Run(name, func(t *testing.T) {
					c, ch, _ := setupTiming(t, cons)
					coefficient := 1.0
					tag := attacks.AttackTagReactionStarDiffusionCryo
					if mode == "superconduct" {
						c.StarReactions.SuperconductActive = true
						c.StarReactions.SuperconductCoefficient = 1.45
						coefficient = 1.45
						tag = attacks.AttackTagReactionStarSuperconduct
					} else {
						c.StarReactions.DiffusionActive = true
						c.Events.Emit(event.OnStarDiffusion, c.Combat.PrimaryTarget(), &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}})
						c.StarReactions.DiffusionStacks = 6
					}
					buff := make([]float64, attributes.EndStatType)
					buff[attributes.CR] = 1
					buff[attributes.CryoP] = 3
					buff[attributes.DmgP] = 3
					ch.AddStatMod(character.StatMod{
						Base: modifier.NewBase("test-stellar-stats", -1), AffectedStat: attributes.NoStat,
						Amount: func() []float64 { return buff },
					})
					// E/Q can hit while Resolution is active. The ray test starts
					// the actual held action below, which supplies its own status.
					if ability != "ray" {
						ch.AddStatus("sandrone-resolution", 300, true)
					}
					hits := 0
					c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
						atk := args[1].(*info.AttackEvent)
						if atk.Info.ActorIndex != ch.Index() || atk.Info.AttackTag != tag {
							return
						}
						hits++
						if !atk.Info.IsDirectStarDamage() || !args[3].(bool) {
							t.Fatal("expected a guaranteed critical direct Star talent hit")
						}
						em := atk.Snapshot.Stats[attributes.EM]
						reactBonus := 0.0
						baseBonus := min(ch.TotalAtk()/100*.007, .14)
						if cons == 1 {
							reactBonus += .30
						}
						want := atk.Info.Mult * atk.Snapshot.Stats.TotalATK() * coefficient *
							(1 + 6*em/(2000+em) + reactBonus) * (1 + baseBonus) * (1 + atk.Snapshot.Stats[attributes.CD])
						if got := args[2].(float64); math.Abs(got-want) > 1e-8 {
							t.Fatalf("%s final damage=%v, want %v (reaction bonus=%v)", atk.Info.Abil, got, want, reactBonus)
						}
					}, "test-stellar-talent-damage")
					var err error
					switch ability {
					case "skill":
						_, err = ch.Skill(nil)
					case "burst":
						_, err = ch.Burst(nil)
					case "ray":
						_, err = ch.ChargeAttack(map[string]int{"duration": 120})
					}
					if err != nil {
						t.Fatal(err)
					}
					advanceTo(t, c, 180)
					if hits != 1 {
						t.Fatalf("got %d direct Star hits, want 1", hits)
					}
				})
			}
		}
	}
}
