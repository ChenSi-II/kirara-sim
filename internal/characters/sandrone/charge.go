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
var resolutionSweep = []float64{.43, .465, .5, .55, .585, .625, .68, .735, .79, .85, .91, .97, 1.03, 1.09, 1.15}
var resolutionRay = []float64{1.2255, 1.32525, 1.425, 1.5675, 1.66725, 1.78125, 1.938, 2.09475, 2.2515, 2.4225, 2.5935, 2.7645, 2.9355, 3.1065, 3.2775}

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	c.resolutionSrc = c.Core.F
	c.resolutionRays = 0
	c.powerOverdrive = false
	c.AddStatus("sandrone-resolution", 6*60, true)
	c.resolutionTick = c.Core.F
	c.QueueCharTask(c.powerTick(c.resolutionTick), 60)
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Resolution Sweep", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Cryo, Durability: 25, Mult: resolutionSweep[c.TalentLvlAttack()]}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 30, 30)
	for delay := 60; delay <= 6*60; delay += 90 {
		c.QueueCharTask(c.resolutionRay(c.resolutionSrc), delay)
	}
	f := frames.InitAbilSlice(52)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 52, CanQueueAfter: 36, State: action.ChargeAttackState}, nil
}

// powerTick models Fagio's continuous decoding-power state machine. It is
// deliberately kept separate from the attack ray schedule so power continues
// to decay while Sandrone is off field and overdrive can use a longer cadence.
func (c *char) powerTick(src int) func() {
	return func() {
		if src != c.resolutionTick {
			return
		}
		if c.StatusIsActive("sandrone-resolution") && !c.powerOverdrive {
			c.resolutionPower = min(100, c.resolutionPower+5)
			if c.resolutionPower >= 100 {
				c.powerOverdrive = true
				c.DeleteStatus("sandrone-resolution")
				c.QueueCharTask(c.overdriveRay(src), 60)
			}
		} else {
			decay := 5
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

func (c *char) overdriveRay(src int) func() {
	return func() {
		if src != c.resolutionTick || !c.powerOverdrive {
			return
		}
		lvl := c.TalentLvlAttack()
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Power Overdrive Ray", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: resolutionSweep[lvl]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0)
		if c.powerOverdrive {
			c.QueueCharTask(c.overdriveRay(src), 120)
		}
	}
}

func (c *char) resolutionRay(src int) func() {
	return func() {
		if src != c.resolutionSrc || !c.StatusIsActive("sandrone-resolution") {
			return
		}
		c.resolutionRays++
		gain := 20
		if c.Base.Cons >= 1 {
			gain /= 2
		}
		c.resolutionPower = min(100, c.resolutionPower+gain)
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Faggio Condensing Ray", AttackTag: attacks.AttackTagExtra, ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Cryo, Durability: 25, Mult: resolutionRay[c.TalentLvlAttack()]}
		clusterMult := 1.0
		if c.Core.StarReactions.SuperconductActive {
			ai.AttackTag, ai.ICDTag, ai.Durability = attacks.AttackTagReactionStarSuperconduct, attacks.ICDTagNone, 0
			clusterMult = .80
		} else if c.Core.StarReactions.DiffusionActive {
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
