package reactable

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

func TestStarSuperconductReplacesSuperconduct(t *testing.T) {
	c, trg := testCoreWithTrgs(1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.StarReactions.Enabled = true

	starCount := 0
	normalCount := 0
	c.Events.Subscribe(event.OnStarSuperconduct, func(args ...any) { starCount++ }, "test-star-superconduct")
	c.Events.Subscribe(event.OnSuperconduct, func(args ...any) { normalCount++ }, "test-normal-superconduct")

	c.QueueAttackEvent(makeAOEAttack(attributes.Cryo, 25), 0)
	advanceCoreFrame(c)
	c.QueueAttackEvent(makeAOEAttack(attributes.Electro, 25), 0)
	advanceCoreFrameMultiple(c, 2)

	if starCount != 1 || normalCount != 0 {
		t.Fatalf("star reactions = %d, normal reactions = %d", starCount, normalCount)
	}
	if !c.StarReactions.SuperconductActive {
		t.Fatal("star superconduct domain was not activated")
	}
	if c.StarReactions.SuperconductCoefficient != 1 {
		t.Fatalf("initial coefficient = %v, want 1", c.StarReactions.SuperconductCoefficient)
	}
	if got := c.Player.ByIndex(0).Stat(attributes.CryoP); !floatApproxEqual(got, 0.28) {
		t.Fatalf("initial cryo bonus = %v, want 0.28", got)
	}
	if trg[0].last.Info.AttackTag != attacks.AttackTagReactionStarSuperconduct {
		t.Fatalf("last attack tag = %v, want Star Superconduct", trg[0].last.Info.AttackTag)
	}
}

func TestStarSuperconductSettlements(t *testing.T) {
	c, trg := testCoreWithTrgs(1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.StarReactions.Enabled = true
	trg[0].activateStarSuperconductDomain()

	application := &info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}}
	for range 20 {
		trg[0].React(application)
	}
	if c.StarReactions.SuperconductStacks != 12 {
		t.Fatalf("stacks = %d, want capped value 12", c.StarReactions.SuperconductStacks)
	}

	advanceCoreFrameMultiple(c, starSuperconductInterval)
	if c.StarReactions.SuperconductStacks != 0 {
		t.Fatalf("stacks after settlement = %d, want 0", c.StarReactions.SuperconductStacks)
	}
	if !floatApproxEqual(c.StarReactions.SuperconductCoefficient, 2) {
		t.Fatalf("12-stack coefficient = %v, want 2", c.StarReactions.SuperconductCoefficient)
	}
	if got := c.Player.ByIndex(0).Stat(attributes.ElectroP); !floatApproxEqual(got, 0.40) {
		t.Fatalf("12-stack electro bonus = %v, want 0.40", got)
	}

	advanceCoreFrameMultiple(c, starSuperconductInterval)
	if c.StarReactions.SuperconductCoefficient != 1 {
		t.Fatalf("0-stack coefficient = %v, want exceptional value 1", c.StarReactions.SuperconductCoefficient)
	}
	if got := c.Player.ByIndex(0).Stat(attributes.CryoP); !floatApproxEqual(got, 0.28) {
		t.Fatalf("0-stack cryo bonus = %v, want 0.28", got)
	}

	trg[0].React(application)
	advanceCoreFrameMultiple(c, starSuperconductInterval)
	if !floatApproxEqual(c.StarReactions.SuperconductCoefficient, 1.45) {
		t.Fatalf("1-stack coefficient = %v, want 1.45", c.StarReactions.SuperconductCoefficient)
	}
	if got := c.Player.ByIndex(0).Stat(attributes.ElectroP); !floatApproxEqual(got, 0.29) {
		t.Fatalf("1-stack electro bonus = %v, want 0.29", got)
	}
}

func TestStarDiffusionReplacesCryoSwirl(t *testing.T) {
	c, trg := testCoreWithTrgs(1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.StarReactions.Enabled = true
	trg[0].SetAuraDurability(info.ReactionModKeyCryo, 25, 0)

	starCount := 0
	swirlCount := 0
	c.Events.Subscribe(event.OnStarDiffusion, func(args ...any) { starCount++ }, "test-star-diffusion")
	c.Events.Subscribe(event.OnSwirlCryo, func(args ...any) { swirlCount++ }, "test-cryo-swirl")

	a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Element: attributes.Anemo, Durability: 25}}
	trg[0].React(a)
	advanceCoreFrame(c)

	if starCount != 1 || swirlCount != 0 {
		t.Fatalf("star reactions = %d, cryo swirls = %d", starCount, swirlCount)
	}
	if c.StarReactions.DiffusionStacks != 1 {
		t.Fatalf("vortex stacks = %d, want 1", c.StarReactions.DiffusionStacks)
	}
	if trg[0].last.Info.AttackTag != attacks.AttackTagReactionStarDiffusionAnemo {
		t.Fatalf("last attack tag = %v, want Star Diffusion Anemo", trg[0].last.Info.AttackTag)
	}
}

func TestStarDiffusionVortexMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		stacks     int
		wantMult   float64
		immediate  bool
		waitFrames int
	}{
		{name: "two stacks use 2x", stacks: 2, wantMult: 2, waitFrames: starDiffusionInterval + 1},
		{name: "three stacks use 3x", stacks: 3, wantMult: 3, waitFrames: starDiffusionInterval + 1},
		{name: "six stacks detonate immediately", stacks: 6, wantMult: 3, immediate: true, waitFrames: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, trg := testCoreWithTrgs(1)
			if err := c.Init(); err != nil {
				t.Fatal(err)
			}
			c.StarReactions.Enabled = true
			trg[0].SetAuraDurability(info.ReactionModKeyCryo, 1000, 0)

			for range tc.stacks {
				a := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, Element: attributes.Anemo, Durability: 25}}
				if !trg[0].tryStarDiffusion(a, attributes.Cryo) {
					t.Fatal("star diffusion did not trigger")
				}
			}
			wantPreDetonationStacks := tc.stacks
			if tc.immediate {
				wantPreDetonationStacks = 0
			}
			if c.StarReactions.DiffusionStacks != wantPreDetonationStacks {
				t.Fatalf("stacks before timed detonation = %d, want %d", c.StarReactions.DiffusionStacks, wantPreDetonationStacks)
			}

			advanceCoreFrameMultiple(c, tc.waitFrames)
			if c.StarReactions.DiffusionStacks != 0 {
				t.Fatalf("stacks after detonation = %d, want 0", c.StarReactions.DiffusionStacks)
			}
			if trg[0].last.Info.AttackTag != attacks.AttackTagReactionStarDiffusionCryo {
				t.Fatalf("last attack tag = %v, want Star Diffusion Cryo", trg[0].last.Info.AttackTag)
			}
			base := combat.CalcReactionBaseDmg(c.Player.ByIndex(0).Base.Level)
			want := tc.wantMult * 0.6 * base
			if !floatApproxEqual(trg[0].last.Info.FlatDmg, want) {
				t.Fatalf("cryo detonation flat damage = %v, want %v", trg[0].last.Info.FlatDmg, want)
			}
		})
	}
}

func TestStarDiffusionVortexAppliesCryo(t *testing.T) {
	c, trg := testCoreWithTrgs(1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.StarReactions.Enabled = true
	c.StarReactions.DiffusionStacks = 1
	c.StarReactions.DiffusionOwner = 0
	c.StarReactions.DiffusionTarget = trg[0]

	trg[0].detonateStarDiffusion()
	advanceCoreFrameMultiple(c, 2)

	if got := trg[0].last.Info.Durability; got != starDiffusionCryoDurability {
		t.Fatalf("cryo detonation durability = %v, want %v", got, starDiffusionCryoDurability)
	}
	wantAura := 20 - 20/float64(6*25+420)
	if got := trg[0].GetAuraDurability(info.ReactionModKeyCryo); !floatApproxEqual(float64(got), wantAura) {
		t.Fatalf("cryo aura after detonation = %v, want %v", got, wantAura)
	}
}

func floatApproxEqual(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}
