package linnea

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	lumiKey      = "linnea-lumi"
	lumiDuration = 25 * 60
	skillHitmark = 24
)

// Skill supports taps=1..5. Five taps immediately fill Lumi and perform
// Million Ton Crush; later Normal Attacks can fill the remaining taps.
// TODO: replace the conservative animation and summon attack timings once
// frame data is available locally.
func (c *char) Skill(p map[string]int) (action.Info, error) {
	taps := p["taps"]
	if taps == 0 {
		taps = 1
	}
	taps = min(max(taps, 1), 5)

	c.SetCD(action.ActionSkill, int(skillParam[4][c.TalentLvlSkill()]*60))
	c.c1OnSkill()
	c.Core.Tasks.Add(func() {
		c.summonLumi(lumiSuper)
		c.lumiFeed = taps
		if c.lumiFeed >= 5 {
			c.millionTonCrush()
		}
	}, skillHitmark)

	f := frames.InitAbilSlice(48)
	return action.Info{
		Frames:          frames.NewAbilFunc(f),
		AnimationLength: 48,
		CanQueueAfter:   34,
		State:           action.SkillState,
	}, nil
}

func (c *char) summonLumi(form lumiForm) {
	c.lumiSrc = c.Core.F
	c.lumiForm = form
	c.lumiFeed = 0
	c.AddStatus(lumiKey, lumiDuration, true)
	src := c.lumiSrc
	if form == lumiSuper {
		c.QueueCharTask(c.lumiPummeler(src), 2*60)
		c.QueueCharTask(c.lumiHeavy(src), 5*60)
	} else {
		c.QueueCharTask(c.lumiPummeler(src), 5*60)
	}
}

func (c *char) refreshLumi() {
	form := c.lumiForm
	feed := c.lumiFeed
	c.summonLumi(form)
	c.lumiFeed = feed
}

func (c *char) feedLumi() {
	if !c.StatusIsActive(lumiKey) || c.lumiForm != lumiSuper {
		return
	}
	c.lumiFeed++
	if c.lumiFeed >= 5 {
		c.millionTonCrush()
	}
}

func (c *char) lumiPummeler(src int) func() {
	return func() {
		if src != c.lumiSrc || !c.StatusIsActive(lumiKey) {
			return
		}
		lvl := c.TalentLvlSkill()
		for hit := 0; hit < 2; hit++ {
			ai := info.AttackInfo{
				ActorIndex: c.Index(),
				Abil:       "Lumi Pound-Pound Pummeler",
				AttackTag:  attacks.AttackTagElementalArt,
				ICDTag:     attacks.ICDTagElementalArt,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeBlunt,
				Element:    attributes.Geo,
				Durability: 25,
				UseDef:     true,
				Mult:       skill[hit][lvl],
			}
			c.Core.QueueAttack(
				ai,
				combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3),
				hit*8,
				hit*8,
				c.particleCB,
			)
		}
		interval := 2 * 60
		if c.lumiForm == lumiStandard {
			interval = 5 * 60
		}
		c.QueueCharTask(c.lumiPummeler(src), interval)
	}
}

func (c *char) lumiHeavy(src int) func() {
	return func() {
		if src != c.lumiSrc || !c.StatusIsActive(lumiKey) || c.lumiForm != lumiSuper {
			return
		}
		moondrifts, _ := c.Core.Constructs.ConstructsByType(construct.GeoConstructLunarCrystallize)
		if len(moondrifts) > 0 {
			ai := info.AttackInfo{
				ActorIndex: c.Index(),
				Abil:       "Heavy Overdrive Hammer",
				AttackTag:  attacks.AttackTagDirectLunarCrystallize,
				ICDTag:     attacks.ICDTagNone,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeBlunt,
				Element:    attributes.Geo,
				UseDef:     true,
				Mult:       skill[2][c.TalentLvlSkill()],
			}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
			c.c2TriggerHarmony()
		}
		c.QueueCharTask(c.lumiHeavy(src), 5*60)
	}
}

func (c *char) millionTonCrush() {
	if !c.StatusIsActive(lumiKey) || c.lumiForm != lumiSuper {
		return
	}
	lvl := c.TalentLvlSkill()
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Million Ton Crush",
		AttackTag:  attacks.AttackTagDirectLunarCrystallize,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Geo,
		UseDef:     true,
		Mult:       skill[3][lvl],
		FlatDmg:    c.consumeCatalogForMillion(),
	}
	snap := c.Snapshot(&ai)
	if c.Base.Cons >= 2 {
		snap.Stats[attributes.CD] += 1.5
	}
	c.Core.QueueAttackWithSnap(
		ai,
		snap,
		combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 5),
		12,
	)
	c.lumiForm = lumiStandard
	c.lumiFeed = 0
	c.c2TriggerHarmony()
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || c.StatusIsActive("linnea-particle-icd") {
		return
	}
	// Particle count is present in local behavior data, but its exact proc
	// distribution is not; use one proc per skill summon as a conservative model.
	c.AddStatus("linnea-particle-icd", 9*60, true)
	c.Core.QueueParticle(c.Base.Key.String(), 3, attributes.Geo, c.ParticleDelay)
}
