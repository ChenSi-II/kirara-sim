package venti

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

const burstStart = 94

func init() {
	burstFrames = frames.InitAbilSlice(95) // Q -> N1/CA/E/D
	burstFrames[action.ActionJump] = 94    // Q -> J
	burstFrames[action.ActionSwap] = 93    // Q -> Swap
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.hexC4()
	c.burstSrc = c.Core.F
	c.burstEnd = c.Core.F + burstStart + 480
	c.burstExtensions = 0
	c.DeleteStatus("venti-hexerei-eye")
	if c.IsHexerei && c.Base.Cons >= 2 {
		c.AddStatus("venti-winds-advent", 15*60, true)
		c.ResetActionCooldown(action.ActionSkill)
	}
	src := c.burstSrc
	// reset location
	c.qAbsorb = attributes.NoElement
	player := c.Core.Combat.Player()
	c.qPos = info.CalcOffsetPoint(player.Pos(), info.Point{Y: 5}, player.Direction())
	c.absorbCheckLocation = combat.NewBoxHitOnTarget(c.qPos, info.Point{Y: -1}, 2.5, 2.5)

	// 8 second duration, tick every .4 second
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Wind's Grand Ode",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurstAnemo,
		ICDGroup:   attacks.ICDGroupVenti,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       burstDot[c.TalentLvlBurst()],
	}
	ap := combat.NewCircleHitOnTarget(c.qPos, nil, 4)

	c.aiAbsorb = ai
	c.aiAbsorb.Abil = "Wind's Grand Ode (Absorbed)"
	c.aiAbsorb.Mult = burstAbsorbDot[c.TalentLvlBurst()]
	c.aiAbsorb.Element = attributes.NoElement

	// snapshot is around cd frame and 1st tick?
	var snap info.Snapshot
	c.Core.Tasks.Add(func() {
		snap = c.Snapshot(&ai)
		c.snapAbsorb = c.Snapshot(&c.aiAbsorb)
	}, 104)

	var cb info.AttackCBFunc
	if c.Base.Cons >= 6 {
		cb = c.c6(attributes.Anemo)
	}

	// Keep the original hitmarks; recurse so hurricane arrows can extend both components.
	var tick func()
	tick = func() {
		if src != c.burstSrc || c.Core.F >= c.burstEnd {
			return
		}
		c.Core.QueueAttackWithSnap(ai, snap, ap, 0, cb)
		c.Core.Tasks.Add(tick, 24)
	}
	c.Core.Tasks.Add(tick, 106)
	c.Core.Tasks.Add(c.absorbCheckQ(src, 0, int((480-24*4)/18)), 106+24*3)
	if c.Base.Ascension >= 4 {
		var finish func()
		finish = func() {
			if src != c.burstSrc {
				return
			}
			if c.Core.F < c.burstEnd {
				c.Core.Tasks.Add(finish, c.burstEnd-c.Core.F)
				return
			}
			c.a4()
		}
		c.Core.Tasks.Add(finish, 480+burstStart)
	}

	c.SetCDWithDelay(action.ActionBurst, 15*60, 81)
	c.ConsumeEnergy(84)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}

func (c *char) burstAbsorbedTicks() {
	var cb info.AttackCBFunc
	if c.Base.Cons >= 6 {
		cb = c.c6(c.qAbsorb)
	}

	ap := combat.NewCircleHitOnTarget(c.qPos, nil, 6)
	src := c.burstSrc
	baseEnd := c.Core.F + 15*24
	var tick func()
	tick = func() {
		if src != c.burstSrc || c.Core.F >= min(c.burstEnd, baseEnd+c.burstExtensions*60) {
			return
		}
		c.Core.QueueAttackWithSnap(c.aiAbsorb, c.snapAbsorb, ap, 0, cb)
		c.Core.Tasks.Add(tick, 24)
	}
	tick()
}

func (c *char) absorbCheckQ(src, count, maxcount int) func() {
	return func() {
		if src != c.burstSrc || c.Core.F >= c.burstEnd {
			return
		}
		c.qAbsorb = c.Core.Combat.AbsorbCheck(c.Index(), c.absorbCheckLocation, attributes.Pyro, attributes.Hydro, attributes.Electro, attributes.Cryo)
		if c.qAbsorb != attributes.NoElement {
			c.aiAbsorb.Element = c.qAbsorb
			switch c.qAbsorb {
			case attributes.Pyro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstPyro
			case attributes.Hydro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstHydro
			case attributes.Electro:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstElectro
			case attributes.Cryo:
				c.aiAbsorb.ICDTag = attacks.ICDTagElementalBurstCryo
			}
			// trigger dmg ticks here
			c.burstAbsorbedTicks()
			return
		}
		// otherwise queue up
		c.Core.Tasks.Add(c.absorbCheckQ(src, count+1, maxcount), 18)
	}
}
