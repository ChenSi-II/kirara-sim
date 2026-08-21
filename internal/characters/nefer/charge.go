package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) ChargeAttack(map[string]int) (action.Info, error) {
	phantasm := c.StatusIsActive(shadowDanceKey) && c.Core.Player.Dew() > 0 && c.phantasmUses < 3
	if c.StatusIsActive(shadowDanceKey) && !phantasm && c.StatusIsActive("nefer-a2-dew-window") && !c.StatusIsActive("nefer-a2-dew-granted") {
		c.AddStatus("nefer-a2-dew-granted", 5*60, true)
		amount := 1 + min(max(c.Stat(attributes.EM)-500, 0)/100*.1, .5)
		c.a2DewRemainder += amount
		gained := int(c.a2DewRemainder)
		c.a2DewRemainder -= float64(gained)
		for i := 0; i < gained; i++ {
			c.Core.Player.AddVerdantDew()
		}
	}
	if phantasm {
		return c.phantasmPerformance()
	}
	for hit, mult := range charge {
		ai := info.AttackInfo{
			ActorIndex: c.Index(), Abil: "Charged Attack", AttackTag: attacks.AttackTagExtra,
			ICDTag: attacks.ICDTagNormalAttack, ICDGroup: attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault, Element: attributes.Dendro,
			Durability: 25, Mult: mult[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 30+hit*6, 30+hit*6)
	}
	f := frames.InitAbilSlice(52)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 52, CanQueueAfter: 36, State: action.ChargeAttackState}, nil
}
