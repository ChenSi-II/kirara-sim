package simulation_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/eval"
	"github.com/genshinsim/gcsim/pkg/gcs/parser"
	"github.com/genshinsim/gcsim/pkg/simulation"
)

// Exercise real scripts through parsing, evaluation, queuing, cooldown/energy
// checks and damage events. No manual resource refills or forced actions.
func TestSandroneRayCountRotations(t *testing.T) {
	for _, tc := range []struct {
		cons                      int
		rotation                  string
		counts                    []int
		skillFrames, chargeFrames []int
		burstFrame, finished      int
	}{
		{0, "skill, charge[rays=3], skill, charge[rays=3], skill, burst, charge[rays=3]", []int{3, 3, 3}, []int{1, 249, 497}, []int{38, 286, 633}, 534, 861},
		{1, "skill, charge[rays=6], skill, burst, charge[rays=6]", []int{6, 6}, []int{1, 429}, []int{38, 565}, 466, 967},
	} {
		for _, param := range []string{"rays", "射线"} {
			t.Run(fmt.Sprintf("C%d/%s", tc.cons, param), func(t *testing.T) {
				input := fmt.Sprintf(`
options duration=25;
target lvl=90 resist=0.1;
sandrone char lvl=90/90 cons=%d talent=9,9,9;
sandrone add weapon="wastergreatsword" refine=1 lvl=90/90;
active sandrone;
sandrone %s;
sandrone attack;
`, tc.cons, strings.ReplaceAll(tc.rotation, "rays", param))
				file := ast.NewFile()
				cfg, program, err := parser.New(file, input).Parse()
				if err != nil {
					t.Fatal(err)
				}
				if len(cfg.Errors) != 0 {
					t.Fatal(cfg.Errors)
				}
				c, err := core.New(core.Opt{Seed: 1})
				if err != nil {
					t.Fatal(err)
				}
				evaluator, err := eval.NewEvaluator(file, program, c)
				if err != nil {
					t.Fatal(err)
				}
				sim, err := simulation.New(cfg, evaluator, c)
				if err != nil {
					t.Fatal(err)
				}
				var skillFrames, chargeFrames, burstFrames, nextFrames []int
				var counts []int
				var lastRay int
				overdrive := 0
				c.Events.Subscribe(event.OnSkill, func(...any) { skillFrames = append(skillFrames, c.F) }, "test-skill")
				c.Events.Subscribe(event.OnBurst, func(...any) { burstFrames = append(burstFrames, c.F) }, "test-burst")
				c.Events.Subscribe(event.OnAttack, func(...any) { nextFrames = append(nextFrames, c.F) }, "test-next")
				c.Events.Subscribe(event.OnChargeAttack, func(...any) {
					chargeFrames = append(chargeFrames, c.F)
					counts = append(counts, 0)
				}, "test-charge")
				c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
					a := args[1].(*info.AttackEvent)
					switch a.Info.Abil {
					case "Faggio Condensing Ray":
						counts[len(counts)-1]++
						lastRay = c.F
					case "Faggio Power Overdrive Ray":
						overdrive++
					}
				}, "test-rays")
				if _, err := sim.Run(); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(counts, tc.counts) || overdrive != 0 {
					t.Fatalf("rays per charge=%v want=%v; overdrive=%d", counts, tc.counts, overdrive)
				}
				if !reflect.DeepEqual(skillFrames, tc.skillFrames) || !reflect.DeepEqual(chargeFrames, tc.chargeFrames) || !reflect.DeepEqual(burstFrames, []int{tc.burstFrame}) {
					t.Fatalf("action times E=%v CA=%v Q=%v", skillFrames, chargeFrames, burstFrames)
				}
				if lastRay != tc.finished || !reflect.DeepEqual(nextFrames, []int{tc.finished}) {
					t.Fatalf("last ray=%d next action=%v, want both at %d", lastRay, nextFrames, tc.finished)
				}
				t.Logf("E=%v CA=%v Q=%v rays=%v finished=%.3fs", skillFrames, chargeFrames, burstFrames, counts, float64(lastRay-1)/60)
			})
		}
	}
}

func TestSandroneCountedChargeAtEndOfScript(t *testing.T) {
	for _, cons := range []int{0, 1} {
		count := 3 * (cons + 1)
		input := fmt.Sprintf(`
options duration=25;
target lvl=90 resist=0.1 hp=1000000000;
sandrone char lvl=90/90 cons=%d talent=9,9,9;
sandrone add weapon="wastergreatsword" refine=1 lvl=90/90;
active sandrone;
sandrone charge[射线=%d];
`, cons, count)
		file := ast.NewFile()
		cfg, program, err := parser.New(file, input).Parse()
		if err != nil {
			t.Fatal(err)
		}
		cfg.Settings.DamageMode = true
		c, err := core.New(core.Opt{Seed: 1, DamageMode: true})
		if err != nil {
			t.Fatal(err)
		}
		evaluator, err := eval.NewEvaluator(file, program, c)
		if err != nil {
			t.Fatal(err)
		}
		sim, err := simulation.New(cfg, evaluator, c)
		if err != nil {
			t.Fatal(err)
		}
		rays := 0
		c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
			if args[1].(*info.AttackEvent).Info.Abil == "Faggio Condensing Ray" {
				rays++
			}
		}, "test-last-action-rays")
		if _, err := sim.Run(); err != nil {
			t.Fatal(err)
		}
		if rays != count {
			t.Fatalf("C%d end-of-script truncated charge: %d/%d rays", cons, rays, count)
		}
	}
}
