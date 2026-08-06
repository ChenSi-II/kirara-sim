package jahoda

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error)  { return c.plunge(p, false) }
func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) { return c.plunge(p, true) }

// TODO: replace conservative hitmarks/cancels when verified frame data is available.
func (c *char) plunge(p map[string]int, high bool) (action.Info, error) {
	c.Core.Player.SetAirborne(player.Grounded)
	if p["collision"] != 0 {
		ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Plunge Collision", AttackTag: attacks.AttackTagPlunge, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Physical, Mult: collision[c.TalentLvlAttack()]}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 1), 30, 30)
	}
	mult, radius, abil := lowPlunge[c.TalentLvlAttack()], 3.0, "Low Plunge"
	if high {
		mult, radius, abil = highPlunge[c.TalentLvlAttack()], 5.0, "High Plunge"
	}
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: abil, AttackTag: attacks.AttackTagPlunge, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Physical, Mult: mult}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius), 40, 40)
	f := frames.InitAbilSlice(65)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 65, CanQueueAfter: 40, State: action.PlungeAttackState}, nil
}
