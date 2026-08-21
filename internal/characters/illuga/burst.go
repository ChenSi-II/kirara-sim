package illuga

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const orioleSongKey = "illuga-haunted-night-oriole-song"

func (c *char) Burst(map[string]int) (action.Info, error) {
	lvl := c.TalentLvlBurst()
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Shadowless Reflection", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagElementalBurst, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypeBlunt, Element: attributes.Geo, Durability: 50, UseDef: true, Mult: burstParam[1][lvl], FlatDmg: burstParam[0][lvl] * c.Stat(attributes.EM)}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5), 42, 42)
	dur := int(burstParam[6][lvl] * 60)
	c.nightingaleStacks = int(burstParam[4][lvl])
	c.consumedStacks = 0
	c.constructStacks = min(15, 5*c.Core.Constructs.Count())
	c.nightingaleStacks += c.constructStacks
	c.AddStatus(orioleSongKey, dur, true)
	c.lightkeepersOath()
	if c.Base.Cons >= 4 {
		for _, ch := range c.Core.Player.Chars() {
			target := ch
			target.AddStatMod(character.StatMod{Base: modifier.NewBaseWithHitlag("illuga-c4", dur), AffectedStat: attributes.DEF, Amount: func() []float64 {
				buff := make([]float64, attributes.EndStatType)
				if target.Index() != c.Core.Player.Active() {
					return buff
				}
				buff[attributes.DEF] = 200
				return buff
			}})
		}
	}
	c.SetCD(action.ActionBurst, int(burstParam[7][lvl]*60))
	c.ConsumeEnergy(60)
	f := frames.InitAbilSlice(82)
	return action.Info{Frames: frames.NewAbilFunc(f), AnimationLength: 82, CanQueueAfter: 70, State: action.BurstState}, nil
}

func (c *char) c2Attack() {
	ai := info.AttackInfo{ActorIndex: c.Index(), Abil: "Aedon (C2)", AttackTag: attacks.AttackTagElementalBurst, ICDTag: attacks.ICDTagNone, ICDGroup: attacks.ICDGroupDefault, StrikeType: attacks.StrikeTypePierce, Element: attributes.Geo, FlatDmg: 4*c.Stat(attributes.EM) + 2*c.TotalDef(false)}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, 2), 0, 0)
}
