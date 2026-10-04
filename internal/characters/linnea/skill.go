package linnea

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	lumiKey      = "linnea-lumi"
	lumiDuration = 25 * 60
	skillHitmark = 24
)

// Skill supports taps=1..5. Five taps immediately fill Lumi and perform
// Million Ton Crush; later Normal Attacks can fill the remaining taps.
// Image 9 supplies swap endpoints and summon hit timelines; the summon
// creation frame and intermediate tap counts remain provisional.
// See PLACEHOLDER_FRAMES.md for image provenance and remaining approximations.
func (c *char) Skill(p map[string]int) (action.Info, error) {
	taps := p["taps"]
	if taps == 0 {
		taps = 1
	}
	taps = min(max(taps, 1), 5)

	c.SetCD(action.ActionSkill, int(skillParam[4][c.TalentLvlSkill()]*60))
	c.c1OnSkill()
	c.Core.Tasks.Add(func() {
		c.summonLumiAt(lumiSuper, c.Core.F-skillHitmark)
		c.lumiFeed = taps
		if c.lumiFeed >= 5 {
			c.millionTonCrush(115 - skillHitmark)
		}
	}, skillHitmark)

	// Single tap can swap at 0.450s; five taps at 1.200s. Interpolate
	// the unmeasured 2–4 tap endpoints, and provisionally share this endpoint
	// with other transitions (the image only measures swapping).
	animation := 27 + (45*(taps-1)+2)/4
	f := frames.InitAbilSlice(animation)
	return action.Info{
		Frames:          frames.NewAbilFunc(f),
		AnimationLength: animation,
		CanQueueAfter:   animation,
		State:           action.SkillState,
	}, nil
}

func (c *char) summonLumi(form lumiForm) {
	// Q has no measured summon timeline; reuse E's relative hit profile.
	c.summonLumiAt(form, c.Core.F)
}

func (c *char) summonLumiAt(form lumiForm, origin int) {
	c.lumiSrc = c.Core.F
	c.lumiForm = form
	c.lumiFeed = 0
	c.AddStatus(lumiKey, lumiDuration, true)
	c.startLumiAttacks(origin)
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
		c.millionTonCrush(12)
	}
}

func (c *char) millionTonCrush(delay int) {
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
		delay,
	)
	c.lumiForm = lumiStandard
	c.lumiFeed = 0
	c.lumiAttackSrc++ // discard the super-form hits already queued
	c.scheduleStandardLumi(c.Core.F + delay)
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
