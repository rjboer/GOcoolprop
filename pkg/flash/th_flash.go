package flash

import (
	"fmt"
	"math"

	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
)

// FlashTH solves for density given temperature and molar enthalpy.
// Returns density in mol/m³.
func FlashTH(fluidData *fluid.FluidData, T, H_target float64) (float64, error) {
	state := core.NewState(fluidData)

	// Objective: H(T, rho) - H_target = 0
	obj := func(rho float64) float64 {
		state.Update(T, rho)
		return state.MolarEnthalpy() - H_target
	}

	rhoCrit := fluidData.States.Critical.RhoMolar
	rhoTripleLiq := fluidData.States.TripleLiquid.RhoMolar
	if rhoTripleLiq == 0 {
		// Fallback if no triple-point data
		rhoTripleLiq = rhoCrit * 2.5
	}

	// ---- Phase preference based on enthalpy ----

	rhoGasGuess := math.Max(1e-8, 0.01*rhoCrit)
	rhoLiqGuess := rhoTripleLiq

	state.Update(T, rhoGasGuess)
	H_gas := state.MolarEnthalpy()

	state.Update(T, rhoLiqGuess)
	H_liq := state.MolarEnthalpy()

	preferLiquid := math.Abs(H_target-H_liq) < math.Abs(H_target-H_gas)

	// ---- Global density range to search ----
	// Low end: very dilute gas; High end: comfortably above typical liquid density.
	rhoMin := 1e-8
	rhoMax := rhoTripleLiq * 3.0
	if rhoMax == 0 {
		rhoMax = rhoCrit * 5.0
	}
	if rhoMax <= rhoMin {
		return 0, fmt.Errorf("invalid rho range [%g, %g]", rhoMin, rhoMax)
	}

	// ---- Scan for sign changes on log scale ----

	const nScan = 200 // tune if needed
	logMin := math.Log(rhoMin)
	logMax := math.Log(rhoMax)
	dlog := (logMax - logMin) / float64(nScan)

	type rootInfo struct {
		rho float64
	}
	roots := make([]rootInfo, 0, 4)

	// Helper to de-duplicate roots
	addRoot := func(r float64) {
		if r <= 0 || math.IsNaN(r) || math.IsInf(r, 0) {
			return
		}
		const relTol = 1e-6
		for _, rr := range roots {
			if math.Abs(r-rr.rho) <= relTol*math.Max(1.0, math.Abs(rr.rho)) {
				return // already have a root here
			}
		}
		roots = append(roots, rootInfo{rho: r})
	}

	// Initial point
	prevRho := rhoMin
	prevVal := obj(prevRho)
	if math.IsNaN(prevVal) || math.IsInf(prevVal, 0) {
		// Try to move a bit inward if the edge is pathological
		prevRho = math.Exp(logMin + dlog)
		prevVal = obj(prevRho)
	}

	for i := 1; i <= nScan; i++ {
		rho := math.Exp(logMin + dlog*float64(i))
		if rho <= prevRho {
			continue
		}

		val := obj(rho)
		if math.IsNaN(val) || math.IsInf(val, 0) {
			// Skip regions where EOS blows up
			prevRho, prevVal = rho, val
			continue
		}

		// Direct hit?
		if val == 0 {
			addRoot(rho)
		}

		// Sign change bracket [prevRho, rho]
		if prevVal*val < 0 {
			a := prevRho
			b := rho
			// Safety: ensure increasing order
			if a > b {
				a, b = b, a
			}

			// Use the same tol semantics as your existing usage: on H.
			const tolH = 1.0 // J/mol
			root, err := solver.Brent(obj, a, b, tolH)
			if err == nil {
				addRoot(root)
			}
		}

		prevRho, prevVal = rho, val
	}

	if len(roots) == 0 {
		return 0, fmt.Errorf("FlashTH: no root found for T=%g K, H=%g J/mol", T, H_target)
	}

	// ---- Pick the "most physical" root given the phase hint ----

	chosen := roots[0].rho
	if preferLiquid {
		// For liquid-like H, prefer the highest density root
		for _, r := range roots {
			if r.rho > chosen {
				chosen = r.rho
			}
		}
	} else {
		// For gas-like H, prefer the lowest density root
		for _, r := range roots {
			if r.rho < chosen {
				chosen = r.rho
			}
		}
	}

	return chosen, nil
}
