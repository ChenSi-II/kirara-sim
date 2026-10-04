package simulation_test

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func latestWeaponCore(t *testing.T, key keys.Weapon, refine int) (*core.Core, *character.CharWrapper, *character.CharWrapper) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	for i := range 2 {
		p := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
		if i == 0 {
			p.Weapon.Key = key
			p.Weapon.Refine = refine
			p.Weapon.Level = 90
			p.Weapon.MaxLevel = 90
		}
		if _, err := c.AddChar(p); err != nil {
			t.Fatal(err)
		}
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	holder := c.Player.ByIndex(0)
	holder.EnergyMax, holder.Energy = 100, 0
	return c, holder, c.Player.ByIndex(1)
}

func weaponNear(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Fatalf("got %.10f, want %.10f", got, want)
	}
}

func weaponHit(c *core.Core, actor int, tag attacks.AttackTag) {
	c.Events.Emit(event.OnEnemyDamage, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor, AttackTag: tag}}, 1.0, false)
}

func TestDiebianCycleCooldownAndSwap(t *testing.T) {
	for _, r := range []int{1, 5} {
		c, h, _ := latestWeaponCore(t, keys.Diebian, r)
		base := h.Stat(attributes.CD)
		ai := info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionAnemo}
		c.Events.Emit(event.OnSkill)
		weaponNear(t, h.Stat(attributes.CD)-base, .40+.16*float64(r))
		c.Events.Emit(event.OnBurst)
		weaponNear(t, h.ReactBonus(ai), .27+.09*float64(r))
		c.Events.Emit(event.OnSkill)
		weaponNear(t, h.Energy, 4.5+.5*float64(r))
		for range 3 {
			c.Events.Emit(event.OnSkill)
		}
		weaponNear(t, h.Energy, 4.5+.5*float64(r))
		c.Events.Emit(event.OnCharacterSwap, 0, 1)
		c.Player.SetActive(1)
		weaponNear(t, h.Stat(attributes.CD), base)
		weaponNear(t, h.ReactBonus(ai), 0)
		c.Events.Emit(event.OnSkill)
		c.Player.SetActive(0)
		c.Events.Emit(event.OnSkill)
		weaponNear(t, h.Stat(attributes.CD)-base, .40+.16*float64(r))
		c.Events.Emit(event.OnSkill)
		c.Events.Emit(event.OnSkill)
		weaponNear(t, h.Energy, 4.5+.5*float64(r)) // swap does not reset energy ICD
		c.F = 240
		for range 3 {
			c.Events.Emit(event.OnSkill)
		}
		weaponNear(t, h.Energy, 2*(4.5+.5*float64(r)))
		c.F += 600
		weaponNear(t, h.Stat(attributes.CD), base)
		weaponNear(t, h.ReactBonus(ai), 0)
	}
}

func TestXuanliuSonggeRefreshAndReactionWindow(t *testing.T) {
	for _, r := range []int{1, 5} {
		c, h, ally := latestWeaponCore(t, keys.XuanliuSongge, r)
		c.Player.SetActive(1)
		hpStats := make([]float64, attributes.EndStatType)
		hpStats[attributes.HP] = 50000
		h.AddStatMod(character.StatMod{Base: modifier.NewBase("test-hp", -1), AffectedStat: attributes.HP, Amount: func() []float64 { return hpStats }})
		baseHP := h.Stat(attributes.HPP)
		baseATK := ally.Stat(attributes.ATKP)
		weaponNear(t, h.Stat(attributes.Heal), .03+.01*float64(r))
		heal := func() { c.Events.Emit(event.OnHeal, &info.HealInfo{Caller: 0}, 1, 0.0, 100.0, 100.0) }
		for _, f := range []int{0, 60, 120} {
			c.F = f
			heal()
		}
		hp := .03 + .01*float64(r)
		weaponNear(t, h.Stat(attributes.HPP)-baseHP, 3*hp)
		wantATK := func(mult float64) float64 {
			return math.Min(math.Floor(math.Max(h.MaxHP()-40000, 0)/1000)*(.003+.001*float64(r)), .06+.02*float64(r)) * 3 * mult
		}
		weaponNear(t, ally.Stat(attributes.ATKP)-baseATK, wantATK(1))
		c.Events.Emit(event.OnFrozen, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}})
		weaponNear(t, h.Stat(attributes.HPP)-baseHP, 3*hp*1.75)
		weaponNear(t, ally.Stat(attributes.ATKP)-baseATK, wantATK(1.75))
		c.F = 420
		weaponNear(t, h.Stat(attributes.HPP)-baseHP, 3*hp)
		c.F = 600
		weaponNear(t, h.Stat(attributes.HPP)-baseHP, 3*hp) // all stacks refreshed at frame 120
		c.F = 720
		weaponNear(t, h.Stat(attributes.HPP), baseHP)
		weaponNear(t, ally.Stat(attributes.ATKP), baseATK)
	}
}

func TestGestOfTheMightyWolfTriggerOwnership(t *testing.T) {
	c, h, ally := latestWeaponCore(t, keys.GestOfTheMightyWolf, 1)
	c.Player.SetActive(1)
	c.Events.Emit(event.OnSkill)
	c.Events.Emit(event.OnChargeAttack)
	weaponHit(c, 1, attacks.AttackTagNormal)
	weaponHit(c, 0, attacks.AttackTagElementalArt)
	weaponNear(t, h.Stat(attributes.DmgP), 0)
	c.Player.SetActive(0)
	c.Events.Emit(event.OnSkill)
	weaponNear(t, h.Stat(attributes.DmgP), .15)
	weaponHit(c, 0, attacks.AttackTagNormal) // same-frame ICD
	weaponNear(t, h.Stat(attributes.DmgP), .15)
	c.F++
	weaponHit(c, 0, attacks.AttackTagNormal)
	weaponNear(t, h.Stat(attributes.DmgP), .225)
	c.F++
	c.Events.Emit(event.OnChargeAttack)
	weaponNear(t, h.Stat(attributes.DmgP), .30)
	baseCD := h.Stat(attributes.CD)
	h.IsHexerei = true
	weaponNear(t, h.Stat(attributes.CD), baseCD)
	ally.IsHexerei = true
	weaponNear(t, h.Stat(attributes.CD)-baseCD, .30)
	c.F += 240
	weaponNear(t, h.Stat(attributes.DmgP), 0)
	weaponNear(t, h.Stat(attributes.CD), baseCD)
}

func TestTranscendenceStarStacks(t *testing.T) {
	for _, r := range []int{1, 5} {
		c, h, _ := latestWeaponCore(t, keys.ATeaspoonOfTranscendence, r)
		weaponHit(c, 1, attacks.AttackTagExtra)
		weaponHit(c, 0, attacks.AttackTagNormal)
		star := info.AttackInfo{AttackTag: attacks.AttackTagReactionStarSuperconduct}
		weaponNear(t, h.ReactBonus(star), 0)
		for _, f := range []int{0, 0, 12, 24, 36} {
			c.F = f
			weaponHit(c, 0, attacks.AttackTagExtra)
		}
		for _, tag := range []attacks.AttackTag{attacks.AttackTagReactionStarSuperconduct, attacks.AttackTagReactionStarDiffusionAnemo, attacks.AttackTagReactionStarDiffusionCryo} {
			weaponNear(t, h.ReactBonus(info.AttackInfo{AttackTag: tag}), 3*(.12+.04*float64(r)))
		}
		weaponNear(t, h.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionLunarCharge}), 0)
		c.F = 336
		weaponNear(t, h.ReactBonus(star), 0)
	}
}

func TestWhitelakeFrostfeatherStarEffects(t *testing.T) {
	for _, r := range []int{1, 5} {
		c, h, _ := latestWeaponCore(t, keys.WhitelakeFrostfeather, r)
		c.Player.SetActive(1)
		for _, f := range []int{0, 6, 12} {
			c.F = f
			weaponHit(c, 0, attacks.AttackTagElementalArt)
		}
		atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0, AttackTag: attacks.AttackTagReactionStarDiffusionAnemo}}
		c.Events.Emit(event.OnStarReactionAttack, nil, atk)
		weaponNear(t, atk.Snapshot.Stats[attributes.CD], .35+.15*float64(r))
		c.Events.Emit(event.OnStarDiffusion, nil, atk)
		weaponNear(t, h.Energy, 3.5+.5*float64(r))
		weaponHit(c, 0, attacks.AttackTagReactionStarDiffusionCryo)
		weaponNear(t, h.Energy, 3.5+.5*float64(r))
		c.F = 222
		weaponHit(c, 0, attacks.AttackTagReactionStarSuperconduct)
		weaponNear(t, h.Energy, 2*(3.5+.5*float64(r)))
		c.F = 480
		atk.Snapshot.Stats[attributes.CD] = 0
		c.Events.Emit(event.OnStarReactionAttack, nil, atk)
		weaponNear(t, atk.Snapshot.Stats[attributes.CD], 0)
		weaponHit(c, 0, attacks.AttackTagReactionStarSuperconduct)
		weaponNear(t, h.Energy, 2*(3.5+.5*float64(r)))
	}
}

func TestLatestReactionWeaponsStarBranches(t *testing.T) {
	for _, key := range []keys.Weapon{keys.Emberwell, keys.BladeOfAtonement, keys.SongOfTheVigil, keys.EchoesOfTheHeart} {
		t.Run(key.String(), func(t *testing.T) {
			c, h, _ := latestWeaponCore(t, key, 1)
			c.Player.SetActive(1)
			baseATK := h.Stat(attributes.ATKP)
			star := info.AttackInfo{AttackTag: attacks.AttackTagReactionStarDiffusionAnemo}
			c.Events.Emit(event.OnStarDiffusion, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 1}})
			weaponNear(t, h.Stat(attributes.ATKP), baseATK)
			weaponNear(t, h.ReactBonus(star), 0)
			c.Events.Emit(event.OnStarDiffusion, nil, &info.AttackEvent{Info: info.AttackInfo{ActorIndex: 0}})
			switch key {
			case keys.Emberwell, keys.EchoesOfTheHeart:
				weaponNear(t, h.ReactBonus(star), .16)
			case keys.BladeOfAtonement:
				weaponNear(t, h.Stat(attributes.ATKP)-baseATK, .16)
			case keys.SongOfTheVigil:
				weaponNear(t, h.Stat(attributes.ATKP)-baseATK, .20)
				weaponNear(t, h.Energy, 4)
			}
			c.F = 720
			weaponNear(t, h.Stat(attributes.ATKP), baseATK)
			weaponNear(t, h.ReactBonus(star), 0)
		})
	}
}
