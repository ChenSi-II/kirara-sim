package combat

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

type reactionBonusSrc interface {
	ReactBonus(atk info.AttackInfo) float64
}

func CalcReactionBaseDmg(lvl int) float64 {
	idx := lvl - 1
	idx = min(idx, 99)
	idx = max(idx, 0)
	return reactionLvlBase[idx]
}

// CalcSpecialReactionDmg computes damage before resistance and crit. Talent
// damage supplies its multiplier times the scaling stat as baseDmg; reaction
// contributions supply CalcReactionBaseDmg(level). Keeping the coefficient
// separate prevents talent multipliers from being mistaken for reaction or
// vortex coefficients. Direct Star Diffusion always uses coefficient 1.
func CalcSpecialReactionDmg(baseDmg, coefficient, reactBonus float64, atk info.AttackInfo, em float64) float64 {
	return (coefficient*(1+((6*em)/(2000+em))+reactBonus)*baseDmg*(1+atk.BaseDmgBonus) + atk.FlatDmg) * (1 + atk.Elevation)
}

// CalcLunarReactionDmg is kept for callers outside the core package. New
// Lunar-like reactions should use CalcSpecialReactionDmg.
func CalcLunarReactionDmg(lvl int, reactBonus float64, atk info.AttackInfo, em float64) float64 {
	coefficient := 0.0
	switch atk.AttackTag {
	case attacks.AttackTagReactionLunarCharge:
		coefficient = 3
	case attacks.AttackTagReactionLunarCrystallize:
		coefficient = 1.6
	}
	return CalcSpecialReactionDmg(CalcReactionBaseDmg(lvl), coefficient, reactBonus, atk, em)
}

func CalcReactionDmg(lvl int, src reactionBonusSrc, atk info.AttackInfo, em float64) (float64, info.Snapshot) {
	snap := info.Snapshot{
		CharLvl: lvl,
	}
	snap.Stats[attributes.EM] = em
	return (1 + ((16 * em) / (2000 + em)) + src.ReactBonus(atk)) * CalcReactionBaseDmg(lvl), snap
}

func CalcCatalyzeDmg(lvl int, src reactionBonusSrc, atk info.AttackInfo, em float64) float64 {
	return (1 + ((5 * em) / (1200 + em)) + src.ReactBonus(atk)) * CalcReactionBaseDmg(lvl)
}
