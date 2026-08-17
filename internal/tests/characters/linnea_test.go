package characters

import (
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/linnea"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func TestLinneaSkillRuntime(t *testing.T) {
	c, targets := makeCore(1)
	prof := testhelper.DefaultProfile(keys.Linnea, testhelper.TestWeaponKey)
	prof.Base.Cons = 6
	prof.Base.Ascension = 6
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add Linnea: %v", err)
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("initialize core: %v", err)
	}
	c.Combat.DefaultTarget = targets[0].Key()

	var millionHits, pummelerHits int
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return
		}
		atk := args[1].(*info.AttackEvent)
		switch atk.Info.Abil {
		case "Million Ton Crush":
			millionHits++
		case "Lumi Pound-Pound Pummeler":
			pummelerHits++
		}
	}, "linnea-runtime-hits")

	if err := c.Player.Exec(action.ActionSkill, keys.Linnea, map[string]int{"taps": 5}); err != nil {
		t.Fatalf("execute skill: %v", err)
	}
	for range 220 {
		advanceCoreFrame(c)
	}

	if c.Flags.Custom[reactable.LunarCrystallizeEnableKey] != 1 {
		t.Error("Linnea did not enable Lunar-Crystallize")
	}
	if millionHits != 1 {
		t.Errorf("expected one Million Ton Crush hit, got %d", millionHits)
	}
	if pummelerHits == 0 {
		t.Error("expected Lumi to continue attacking after switching to Standard Power Form")
	}
	got, err := c.Player.Chars()[idx].Condition([]string{"field-catalog"})
	if err != nil {
		t.Fatalf("query Field Catalog: %v", err)
	}
	if got != 8 {
		t.Errorf("expected C6 Million Ton Crush to leave 8 Field Catalog stacks, got %v", got)
	}

	linnea := c.Player.Chars()[idx]
	linnea.SetHPByRatio(0.25)
	linnea.AddEnergy("linnea-runtime-test", 100)
	beforeHeal := linnea.CurrentHP()
	if err := c.Player.Exec(action.ActionBurst, keys.Linnea, nil); err != nil {
		t.Fatalf("execute burst: %v", err)
	}
	for range 60 {
		advanceCoreFrame(c)
	}
	afterInitialHeal := linnea.CurrentHP()
	if afterInitialHeal <= beforeHeal {
		t.Error("expected Burst initial party healing")
	}
	form, err := linnea.Condition([]string{"lumi-form"})
	if err != nil {
		t.Fatalf("query Lumi form: %v", err)
	}
	if form != "standard" {
		t.Errorf("Burst changed an existing Lumi's form; got %v", form)
	}
	for range 121 {
		advanceCoreFrame(c)
	}
	if linnea.CurrentHP() <= afterInitialHeal {
		t.Error("expected Burst continuous active-character healing")
	}
}
