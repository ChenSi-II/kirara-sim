package sandrone

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Charge multipliers are exported by origin_data but omitted from the
// generated action table because charged attacks have stateful follow-up hits.
var (
	resolutionSweep = []float64{.43, .465, .5, .55, .585, .625, .68, .735, .79, .85, .91, .97, 1.03, 1.09, 1.15}
	resolutionRay   = []float64{1.2255, 1.32525, 1.425, 1.5675, 1.66725, 1.78125, 1.938, 2.09475, 2.2515, 2.4225, 2.5935, 2.7645, 2.9355, 3.1065, 3.2775}
)

// Charge is a held action, not an off-field summon. The requested duration is
// in frames; its default 6s and release recovery (0f) remain assumptions.
// Hit schedules come from images 1/6; see PLACEHOLDER_FRAMES.md.
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	duration := 6 * 60
	if d, ok := p["duration"]; ok {
		duration = max(1, d)
	}
	c.resolutionSrc++
	src := c.resolutionSrc
	c.resolutionRays = 0
	c.resolutionChannel = true
	c.sweepTailUntil = -1
	if !c.powerOverdrive {
		c.AddStatus("sandrone-resolution", duration+1, true)
	}
	c.resolutionTick++
	c.QueueCharTask(c.powerTick(c.resolutionTick), 60)
	sweeps, rays := resolutionSweepHitmarks, resolutionRayHitmarks
	if c.Base.Cons >= 1 {
		sweeps, rays = resolutionC1SweepHitmarks, resolutionC1RayHitmarks
	}
	// LastAction survives waits, so only apply the combined recording to
	// an immediate E follow-up (37f animation plus at most 1f scheduling).
	sinceSkill := c.Core.F - c.skillStart
	followsSkill := c.Core.Player.LastAction.Char == c.Index() && c.Core.Player.LastAction.Type == action.ActionSkill && sinceSkill >= 37 && sinceSkill <= 38
	if followsSkill {
		sweeps, rays = skillResolutionSweepHitmarks, skillResolutionRayHitmarks
	}
	if c.powerOverdrive {
		// Re-pressing may shoot in overdrive; it must not reset the gauge
		// or grant Resolution. The restart startup is still unmeasured and
		// uses the recorded transition's first-shot offset.
		c.QueueCharTask(c.overdriveRay(c.resolutionTick, 0), overdriveHitmarks[0])
	} else {
		for i, delay := 0, sweeps[0]; delay <= duration; i, delay = i+1, nextChannelHit(sweeps, i+1, 20) {
			c.QueueCharTask(c.resolutionSweepHit(src, i < len(sweeps)), delay)
		}
		for i, delay := 0, rays[0]; delay <= duration; i, delay = i+1, nextChannelHit(rays, i+1, 60) {
			c.QueueCharTask(c.resolutionRay(src), delay)
		}
	}
	endChannel := func() {
		if src != c.resolutionSrc {
			return
		}
		c.resolutionChannel = false
		c.DeleteStatus("sandrone-resolution")
	}
	// Include hits on the requested final frame, then stop even when no next
	// action is submitted. Interruptions invalidate the callbacks immediately.
	c.QueueCharTask(endChannel, duration+1)
	f := frames.InitAbilSlice(duration)
	queueAfter := duration
	if _, explicitHold := p["duration"]; !explicitHold && followsSkill {
		// Default E -> held CA -> E loops may leave CA when the next E is
		// ready, rather than adding the old arbitrary 6s lock to every E.
		// This cancel is inferred; explicit duration keeps the requested hold.
		f[action.ActionSkill] = min(duration, max(sweeps[0], c.Cooldown(action.ActionSkill)))
		queueAfter = min(duration, sweeps[0])
	}
	return action.Info{
		Frames: frames.NewAbilFunc(f), AnimationLength: duration, CanQueueAfter: queueAfter,
		State:     action.ChargeAttackState,
		OnRemoved: func(action.AnimationState) { endChannel() },
	}, nil
}

func nextChannelHit(recorded []int, i, interval int) int {
	if i < len(recorded) {
		return recorded[i]
	}
	return recorded[len(recorded)-1] + (i-len(recorded)+1)*interval
}

func (c *char) resolutionSweepHit(src int, recorded bool) func() {
	return func() {
		if src != c.resolutionSrc || !c.resolutionChannel || c.Core.Player.Active() != c.Index() {
			return
		}
		// The final recorded sweep lands after the ray which triggers
		// overdrive (C0: 229 > 228; C1: 409 > 402). Preserve this in-flight
		// damage, not the ability to fire new/extrapolated resolution shots.
		tail := recorded && c.powerOverdrive && c.Core.F < c.sweepTailUntil
		if !c.StatusIsActive("sandrone-resolution") && !tail {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Resolution Sweep", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Cryo, Durability: 25, Mult: resolutionSweep[c.TalentLvlAttack()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
	}
}

// powerTick models Fagio's continuous decoding-power state machine. It is
// deliberately kept separate from the attack ray schedule so power continues
// to decay while Sandrone is off field and overdrive can use a longer cadence.
func (c *char) powerTick(src int) func() {
	return func() {
		if src != c.resolutionTick {
			return
		}
		if c.resolutionChannel && c.Core.Player.Active() == c.Index() && c.StatusIsActive("sandrone-resolution") && !c.powerOverdrive {
			gain := 10.0 // fitted gauge rate, not independently measured
			if c.Base.Cons >= 1 {
				gain /= 2
			}
			c.gainResolutionPower(gain)
		} else {
			decay := 5.0
			if c.Core.Player.Active() != c.Index() {
				decay *= 3
			}
			c.reduceResolutionPower(decay)
			if c.resolutionPower == 0 && !c.powerOverdrive {
				return
			}
		}
		c.QueueCharTask(c.powerTick(src), 60)
	}
}

func (c *char) gainResolutionPower(gain float64) {
	c.resolutionPower = min(100, c.resolutionPower+gain)
	if c.resolutionPower < 100-1e-9 || c.powerOverdrive {
		return
	}
	c.resolutionPower = 100
	c.powerOverdrive = true
	c.DeleteStatus("sandrone-resolution")
	c.sweepTailUntil = c.Core.F + overdriveHitmarks[0]
	c.QueueCharTask(c.overdriveRay(c.resolutionTick, 0), overdriveHitmarks[0])
}

func (c *char) overdriveRay(src, shot int) func() {
	return func() {
		if src != c.resolutionTick || !c.powerOverdrive || !c.resolutionChannel || c.Core.Player.Active() != c.Index() {
			return
		}
		lvl := c.TalentLvlAttack()
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Power Overdrive Ray", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: resolutionSweep[lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		if c.powerOverdrive {
			delay := nextChannelHit(overdriveHitmarks, shot+1, 24) - nextChannelHit(overdriveHitmarks, shot, 24)
			c.QueueCharTask(c.overdriveRay(src, shot+1), delay)
		}
	}
}

func (c *char) resolutionRay(src int) func() {
	return func() {
		if src != c.resolutionSrc || !c.resolutionChannel || c.Core.Player.Active() != c.Index() || !c.StatusIsActive("sandrone-resolution") {
			return
		}
		c.resolutionRays++
		// User-confirmed zero-power boundary: C0 ray #3, C1 ray #6.
		// Both recordings accumulate 30 continuous power before that ray.
		// Allocate the remaining 70 across three C0 (six C1) rays. This is
		// a gauge interpolation, not an independently measured per-ray gain.
		gain := 70.0 / 3
		if c.Base.Cons >= 1 {
			gain /= 2
		}
		// Rays can themselves reach the threshold; do not defer the mode
		// switch until the next once-per-second power tick.
		c.gainResolutionPower(gain)
		// Image 6 identifies Z2/Z3 as blunt, including their Stellar variants.
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Condensing Ray", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Cryo, Durability: 25, Mult: resolutionRay[c.TalentLvlAttack()]}
		clusterMult := 1.0
		if c.SuperconductRadiance() {
			ai.AttackTag, ai.ICDTag, ai.Durability = attacks.AttackTagReactionStarSuperconduct, attacks.ICDTagNone, 0
			ai.Mult *= 2.0 / 3.0 // Talent table: 81.7% superconduct vs 122.55% ordinary/diffusion at level 1.
			clusterMult = .80
		} else if c.DiffusionRadiance() {
			ai.AttackTag, ai.ICDTag, ai.Durability = attacks.AttackTagReactionStarDiffusionCryo, attacks.ICDTagNone, 0
			clusterMult = 1.20
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		if c.Base.Cons >= 6 && c.resolutionRays == 3 {
			cluster := ai
			cluster.Abil = "Faggio Cluster Condensing Ray"
			cluster.Mult = clusterMult
			for i := 0; i < 4; i++ {
				c.Core.QueueAttack(cluster, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 6+i*6, 6+i*6)
			}
		}
	}
}
