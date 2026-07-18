package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"GOcoolprop/pkg/solver"
	"fmt"
	"math"
)

// FlashTH solves for density given temperature and molar enthalpy.
// Only stable single-phase states and saturation endpoints are supported.
func FlashTH(fluidData *fluid.FluidData, T, H_target float64) (float64, error) {
	var (
		hL, hV            float64
		rhoL, rhoV        float64
		hasSatBand        bool
		satBandLow        float64
		satBandHigh       float64
		preferredSeed     float64
		preferredLiquid   bool
	)
	if T > 0 && T < fluidData.States.Critical.T {
		var err error
		rhoL, err = saturation.RhoL(fluidData, T)
		if err == nil {
			rhoV, err = saturation.RhoV(fluidData, T)
			if err == nil {
				state, err := core.NewState(fluidData)
				if err == nil {
					state.Update(T, rhoL)
					hL = state.MolarEnthalpy()
					state.Update(T, rhoV)
					hV = state.MolarEnthalpy()
					hasSatBand = true
					satBandLow = math.Min(hL, hV)
					satBandHigh = math.Max(hL, hV)
					preferredLiquid = math.Abs(H_target-hL) <= math.Abs(H_target-hV)
					if preferredLiquid {
						preferredSeed = rhoL
					} else {
						preferredSeed = rhoV
					}
					scale := math.Max(1, math.Max(math.Abs(H_target), math.Max(math.Abs(hL), math.Abs(hV))))
					if math.Abs(H_target-hL)/scale <= 1e-8 {
						return rhoL, nil
					}
					if math.Abs(H_target-hV)/scale <= 1e-8 {
						return rhoV, nil
					}
				}
			}
		}
	}

	state, err := core.NewState(fluidData)
	if err != nil {
		return 0, err
	}
	obj := func(rho float64) float64 {
		state.Update(T, rho)
		return state.MolarEnthalpy() - H_target
	}
	stableRoot := func(root float64) bool {
		state.Update(T, root)
		return state.Pressure() > 0 && state.DPdRho() > 0 && !math.IsNaN(state.MolarEnthalpy()) && !math.IsInf(state.MolarEnthalpy(), 0)
	}
	trySeedBracket := func(seed float64) (float64, bool) {
		if seed <= 0 {
			return 0, false
		}
		multipliers := [][2]float64{
			{0.995, 1.005},
			{0.99, 1.01},
			{0.98, 1.02},
			{0.95, 1.05},
			{0.9, 1.1},
			{0.75, 1.25},
			{0.5, 1.5},
			{0.25, 2.0},
		}
		for _, pair := range multipliers {
			a := seed * pair[0]
			b := seed * pair[1]
			if a <= 0 {
				a = seed * 0.1
			}
			fa := obj(a)
			fb := obj(b)
			if math.IsNaN(fa) || math.IsInf(fa, 0) || math.IsNaN(fb) || math.IsInf(fb, 0) {
				continue
			}
			if fa == 0 && stableRoot(a) {
				return a, true
			}
			if fb == 0 && stableRoot(b) {
				return b, true
			}
			if fa*fb < 0 {
				root, err := solver.Brent(obj, a, b, 1e-10)
				if err == nil && stableRoot(root) {
					return root, true
				}
			}
		}
		return 0, false
	}

	if hasSatBand {
		if root, ok := trySeedBracket(preferredSeed); ok {
			return root, nil
		}
		if H_target > satBandLow && H_target < satBandHigh {
			if preferredLiquid {
				if root, ok := trySeedBracket(rhoV); ok {
					return root, nil
				}
			} else {
				if root, ok := trySeedBracket(rhoL); ok {
					return root, nil
				}
			}
		}
	}

	R := fluidData.EOS[0].GasConstant
	ideal := 1.0
	if T > 0 {
		ideal = fluidData.States.Critical.P / (R * T)
	}
	maxRho := state.ReducingRho * 5
	if fluidData.States.TripleLiquid.RhoMolar > 0 {
		maxRho = fluidData.States.TripleLiquid.RhoMolar * 3
	}

	roots, err := bracketedRoots(1e-12, maxRho, 220, obj, 1e-10)
	if err != nil {
		if hasSatBand && H_target > satBandLow && H_target < satBandHigh {
			return 0, fmt.Errorf("two-phase T-H flash unsupported for fluid=%s at T=%g", fluidData.Info.Name, T)
		}
		return 0, fmt.Errorf("T-H flash failed for fluid=%s T=%g H=%g: %w", fluidData.Info.Name, T, H_target, err)
	}

	stableRoots := make([]float64, 0, len(roots))
	for _, root := range roots {
		if stableRoot(root) {
			stableRoots = append(stableRoots, root)
		}
	}
	if len(stableRoots) == 0 {
		if hasSatBand && H_target > satBandLow && H_target < satBandHigh {
			return 0, fmt.Errorf("two-phase T-H flash unsupported for fluid=%s at T=%g", fluidData.Info.Name, T)
		}
		return 0, fmt.Errorf("T-H flash failed for fluid=%s T=%g H=%g: no stable root found", fluidData.Info.Name, T, H_target)
	}

	if hasSatBand {
		if H_target > satBandHigh {
			chosen := stableRoots[0]
			for _, root := range stableRoots[1:] {
				if root < chosen {
					chosen = root
				}
			}
			return chosen, nil
		}
		if H_target < satBandLow {
			chosen := stableRoots[0]
			for _, root := range stableRoots[1:] {
				if root > chosen {
					chosen = root
				}
			}
			return chosen, nil
		}
	}

	chosen := stableRoots[0]
	for _, root := range stableRoots[1:] {
		if abs(root-ideal) < abs(chosen-ideal) {
			chosen = root
		}
	}
	if hasSatBand && H_target > satBandLow && H_target < satBandHigh && len(roots) > 1 {
		if math.Abs(H_target-hL) <= math.Abs(H_target-hV) {
			for _, root := range stableRoots {
				if root > chosen {
					chosen = root
				}
			}
		} else {
			for _, root := range stableRoots {
				if root < chosen {
					chosen = root
				}
			}
		}
	}
	return chosen, nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
