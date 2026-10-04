package vodyanitsa

import (
	"fmt"
	"reflect"
	"testing"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func TestImageSummonTimelinePersistsOffField(t *testing.T) {
	for _, cons := range []int{0, 2} {
		t.Run(fmt.Sprintf("C%d", cons), func(t *testing.T) {
			c, ch, _ := setupVodyanitsa(t, cons)
			// The lightweight test character has a no-op Heal method. Use the
			// actual template healing path to observe off-field healing events.
			ally := c.Player.ByIndex(1)
			ally.Character = tmpl.NewWithWrapper(c, ally)
			var opening, attacks, heals []int
			c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
				a := args[1].(*info.AttackEvent)
				switch a.Info.Abil {
				case "Water Nymph Overture":
					opening = append(opening, c.F)
				case "Microphone Performance":
					attacks = append(attacks, c.F)
				}
			}, "timing-summon-damage")
			c.Events.Subscribe(event.OnHeal, func(args ...any) {
				heal := args[0].(*info.HealInfo)
				if heal.Caller == ch.Index() && heal.Message == "Microphone Performance" {
					heals = append(heals, c.F)
				}
			}, "timing-summon-healing")
			ai, err := ch.Skill(nil)
			if err != nil {
				t.Fatal(err)
			}
			if ai.AnimationLength != 63 || ai.Frames(action.ActionSwap) != 63 {
				t.Fatalf("E should permit swap at 63, got %+v", ai)
			}
			for c.F < 25*60 {
				c.F++
				c.Tick()
				if c.F == 63 {
					c.Player.SetActive(1)
				}
			}
			wantAttacks := []int{228, 401, 580, 752, 934}
			wantHeals := []int{138, 222, 312, 398, 489, 574, 666, 748, 839, 924}
			if cons >= 2 {
				wantAttacks = append(wantAttacks, 1106, 1285, 1457)
				wantHeals = append(wantHeals, 1017, 1101, 1190, 1275, 1364, 1451)
			}
			if !reflect.DeepEqual(opening, []int{38}) || !reflect.DeepEqual(attacks, wantAttacks) || !reflect.DeepEqual(heals, wantHeals) {
				t.Fatalf("initial %v, attacks %v, heals %v; expected [38], %v, %v", opening, attacks, heals, wantAttacks, wantHeals)
			}
		})
	}
}

func TestImageBurstDamagePrecedesAnimationEnd(t *testing.T) {
	c, ch, _ := setupVodyanitsa(t, 0)
	var hits []int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Water Nymph Aria" {
			hits = append(hits, c.F)
		}
	}, "timing-burst")
	ai, err := ch.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	for range 105 {
		c.F++
		c.Tick()
	}
	if !reflect.DeepEqual(hits, []int{102}) || ai.AnimationLength != 105 || ai.Frames(action.ActionSwap) != 105 {
		t.Fatalf("Q hits %v, end %d, swap %d", hits, ai.AnimationLength, ai.Frames(action.ActionSwap))
	}
}
