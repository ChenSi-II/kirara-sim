package lohen

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func TestAuditWillUsesBaseAttackAndPerActorCooldown(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	stats := make([]float64, attributes.EndStatType)
	stats[attributes.ATKP] = 5
	ch.AddStatMod(character.StatMod{Base: modifier.NewBase("audit-atk", -1), AffectedStat: attributes.ATKP, Amount: func() []float64 { return stats }})
	e := c.Combat.PrimaryTarget()
	a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}}
	base := ch.Stat(attributes.BaseATK)
	c.Events.Emit(event.OnEnemyDamage, e, a, 10*base, false)
	if ch.will != 20 {
		t.Fatalf("10x Base ATK must grant 20, got %d", ch.will)
	}
	c.Events.Emit(event.OnEnemyDamage, e, a, 30*base, false)
	if ch.will != 20 {
		t.Fatal("same-frame second hit bypassed 0.1s cooldown")
	}
	advanceTo(t, c, 7)
	c.Events.Emit(event.OnEnemyDamage, e, a, 30*base, false)
	if ch.will != 100 {
		t.Fatalf("30x Base ATK must grant 20+60, got total %d", ch.will)
	}
}

func TestAuditStanceNormalsUseSkillTalent(t *testing.T) {
	c, ch, _ := setupTiming(t, 0)
	if _, err := ch.Skill(nil); err != nil {
		t.Fatal(err)
	}
	seen := false
	want := skillParam[0][ch.skillLevel()]
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		a := args[1].(*info.AttackEvent)
		if a.Info.Abil != "Normal 1-1" {
			return
		}
		seen = true
		if math.Abs(a.Info.Mult-want) > 1e-9 {
			t.Fatalf("stance N1 multiplier=%v, want skill talent %v", a.Info.Mult, want)
		}
	}, "audit-stance-normal")
	if _, err := ch.Attack(nil); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, c, 50)
	if !seen {
		t.Fatal("normal did not hit")
	}
}
