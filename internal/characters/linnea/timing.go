package linnea

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Image 9: absolute hit offsets from E start, seconds rounded to 60 FPS.
// 月笼 means a Lunar Crystallize construct, not the team's Moonsign level.
// Source details and unmeasured continuations: PLACEHOLDER_FRAMES.md.
var lumiNoMoonPummels = [][2]int{
	{118, 141}, {263, 285}, {410, 434}, {557, 580}, {702, 724},
	{848, 872}, {995, 1019}, {1142, 1165}, {1289, 1307},
	{1434, 1457}, // extrapolated final pair; image's tenth pair is blank
}

var lumiLaterMoonPummels = [][2]int{
	{119, 141}, {292, 315}, {418, 441}, {613, 637},
	{757, 779}, {932, 958}, {1082, 1098}, {1252, 1277},
	{1419, 1442}, // extrapolated after the last recorded pair
}

var lumiInitialMoonPummels = [][2]int{
	{200, 224}, {340, 361}, {521, 545}, {661, 682},
	{842, 866}, {980, 1003}, {1163, 1187}, {1303, 1324},
}

func (c *char) hasMoondrift() bool {
	moondrifts, _ := c.Core.Constructs.ConstructsByType(construct.GeoConstructLunarCrystallize)
	return len(moondrifts) > 0
}

func (c *char) startLumiAttacks(origin int) {
	c.lumiAttackSrc++
	if c.lumiForm == lumiStandard {
		// Q refreshing the standard form was not measured. Retain its
		// provisional five-second startup; subsequent intervals use image 9.
		c.scheduleStandardLumi(origin + 163)
		return
	}
	if c.hasMoondrift() {
		c.scheduleLumiPummels(origin, lumiInitialMoonPummels)
		// Last hit is extrapolated at the observed 321-frame heavy interval.
		c.scheduleLumiHeavies(origin, []int{138, 459, 780, 1101, 1422})
		return
	}
	c.scheduleLumiPummels(origin, lumiNoMoonPummels)
	c.QueueCharTask(c.awaitMoondrift(c.lumiSrc, c.lumiAttackSrc, origin), max(0, origin+230-c.Core.F))
}

func (c *char) awaitMoondrift(src, generation, origin int) func() {
	return func() {
		if !c.lumiTimelineActive(src, generation) || c.lumiForm != lumiSuper {
			return
		}
		if !c.hasMoondrift() {
			// No video covers a much later construct arrival. Polling once a
			// second and restarting the measured heavy pattern is provisional.
			c.QueueCharTask(c.awaitMoondrift(src, generation, origin), 60)
			return
		}
		c.lumiAttackSrc++
		if c.Core.F == origin+230 {
			// The second column starts without a construct, but has one by
			// the first heavy hit. Replace future normal hits with that trace.
			c.scheduleLumiPummels(origin, lumiLaterMoonPummels)
			c.scheduleLumiHeavies(origin, []int{230, 553, 868, 1193})
			return
		}
		origin = c.Core.F - 138
		c.scheduleLumiPummels(origin, lumiInitialMoonPummels)
		c.scheduleLumiHeavies(origin, []int{138, 459, 780, 1101, 1422})
	}
}

func (c *char) lumiTimelineActive(src, generation int) bool {
	return src == c.lumiSrc && generation == c.lumiAttackSrc && c.StatusIsActive(lumiKey)
}

func (c *char) scheduleLumiPummels(origin int, hitmarks [][2]int) {
	src, generation := c.lumiSrc, c.lumiAttackSrc
	for _, pair := range hitmarks {
		for hit, hitmark := range pair {
			delay := origin + hitmark - c.Core.F
			if delay < 0 {
				continue
			}
			c.QueueCharTask(func() {
				if !c.lumiTimelineActive(src, generation) {
					return
				}
				ai := info.AttackInfo{
					ActorIndex: c.Index(), Abil: "Lumi Pound-Pound Pummeler",
					AttackTag: attacks.AttackTagElementalArt, ICDTag: attacks.ICDTagElementalArt,
					ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt,
					Element: attributes.Geo, Durability: 25, UseDef: true,
					Mult: skill[hit][c.TalentLvlSkill()],
				}
				c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 3), 0, 0, c.particleCB)
			}, delay)
		}
	}
}

func (c *char) scheduleLumiHeavies(origin int, hitmarks []int) {
	src, generation := c.lumiSrc, c.lumiAttackSrc
	for _, hitmark := range hitmarks {
		delay := origin + hitmark - c.Core.F
		if delay < 0 {
			continue
		}
		c.QueueCharTask(func() {
			if !c.lumiTimelineActive(src, generation) || c.lumiForm != lumiSuper || !c.hasMoondrift() {
				return
			}
			ai := info.AttackInfo{
				ActorIndex: c.Index(), Abil: "Heavy Overdrive Hammer",
				AttackTag: attacks.AttackTagDirectLunarCrystallize, ICDTag: attacks.ICDTagNone,
				ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt,
				Element: attributes.Geo, UseDef: true, Mult: skill[2][c.TalentLvlSkill()],
			}
			c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 4), 0, 0)
			c.c2TriggerHarmony()
		}, delay)
	}
}

func (c *char) scheduleStandardLumi(crushHit int) {
	// Image 9 continuous taps: E3 at 115f, then pairs 252/273, 579/601,
	// 906/928 and 1232/1252. Subtract E3 to also support normal-attack feeding.
	c.scheduleLumiPummels(crushHit, [][2]int{{137, 158}, {464, 486}, {791, 813}, {1117, 1137}})
}
