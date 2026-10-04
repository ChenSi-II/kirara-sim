package alyosha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const hunterAdvanceKey = "alyosha-hunter-advance"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	// Image 4 absolute Q timelines, rounded at 60 FPS. Field duration starts
	// after the cast begins; do not truncate the final observed projectile.
	// See PLACEHOLDER_FRAMES.md for the image provenance.
	fieldHits := []int{82, 199, 317, 434, 552, 668, 785}
	tugarinHits := []int{128, 245, 361, 479, 596, 715, 829}
	dur := 860 // field disappearance at 14.333s
	if c.Base.Cons >= 2 {
		fieldHits = append(fieldHits, 903, 1023, 1137)
		tugarinHits = append(tugarinHits, 946, 1068, 1181)
		dur = 1217 // 20.283s, including the field's startup
	}
	c.SetCD(action.ActionBurst, int(burstParam[3][lvl]*60))
	c.ConsumeEnergy(18)
	c.burstSrc = c.Core.F
	src := c.burstSrc
	c.AddStatus(hunterAdvanceKey, dur, true)
	for _, delay := range fieldHits {
		c.QueueCharTask(c.huntingFieldTick(src), delay)
	}
	for _, delay := range tugarinHits {
		c.QueueCharTask(c.tugarinTick(src), delay)
	}
	f := frames.InitAbilSlice(57) // Q ends at 0.950s
	// QE totals 1.450s; tap E lasts 0.633s, so Q→E is 49 frames.
	f[action.ActionSkill] = 49
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 57, CanQueueAfter: 49, State: action.BurstState}, nil
}

func (c *char) huntingFieldTick(src int) func() {
	return func() {
		if src != c.burstSrc || !c.StatusIsActive(hunterAdvanceKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Fulgurite Hunting Field", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeDefault, Element: attributes.Electro, Durability: 25, Mult: burst[0][c.TalentLvlBurst()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
	}
}

func (c *char) tugarinTick(src int) func() {
	return func() {
		if src != c.burstSrc || !c.StatusIsActive(hunterAdvanceKey) {
			return
		}
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Tugarin", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Electro, Durability: 25, Mult: burst[1][c.TalentLvlBurst()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
		c.activateMark()
		if c.Base.Ascension >= 1 {
			c.healActive(1.2, "Awakened by the Baying Hounds")
		}
		if c.Base.Cons >= 4 {
			c.healLowest(.6, "Harvest the Spoils")
		}
	}
}

func (c *char) healActive(scale float64, message string) {
	c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: c.Core.Player.Active(), Message: message, Src: scale * c.TotalAtk(), Bonus: c.Stat(attributes.Heal)})
}

func (c *char) healLowest(scale float64, message string) {
	target := c.Core.Player.Active()
	lowest := 2.0
	for _, ch := range c.Core.Player.Chars() {
		ratio := ch.CurrentHPRatio()
		if ratio < lowest {
			lowest, target = ratio, ch.Index()
		}
	}
	c.Core.Player.Heal(info.HealInfo{Caller: c.Index(), Target: target, Message: message, Src: scale * c.TotalAtk(), Bonus: c.Stat(attributes.Heal)})
}
