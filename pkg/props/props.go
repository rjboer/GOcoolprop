package props

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/flash"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"GOcoolprop/pkg/transport"
	"fmt"
	"strings"
)

func PropSI(output, name1 string, val1 float64, name2 string, val2 float64, fluidName string) (float64, error) {
	// Load fluid
	f, err := fluid.LoadFluidByName(fluidName, "data")
	if err != nil {
		// Try relative path if running from tests
		f, err = fluid.LoadFluidByName(fluidName, "../../data")
		if err != nil {
			return 0, fmt.Errorf("fluid not found: %v", err)
		}
	}

	state, err := core.NewState(f)
	if err != nil {
		return 0, err
	}

	var T, Rho float64
	var knownSatP, knownSatT float64
	var hasKnownSatP, hasKnownSatT bool

	// Normalize inputs
	name1 = strings.ToUpper(name1)
	name2 = strings.ToUpper(name2)
	output = strings.ToUpper(output)

	// -------- Input cases --------

	// Case 1: T and D (density given directly)
	if (name1 == "T" && name2 == "D") || (name1 == "D" && name2 == "T") {
		if name1 == "T" {
			T = val1
			Rho = val2
		} else {
			Rho = val1
			T = val2
		}

	} else if (name1 == "T" && name2 == "P") || (name1 == "P" && name2 == "T") {
		var P_target float64
		if name1 == "T" {
			T = val1
			P_target = val2
		} else {
			P_target = val1
			T = val2
		}
		Rho, err = flash.DensityTP(f, T, P_target)
		if err != nil {
			return 0, fmt.Errorf("T-P density solve failed for output=%s with inputs %s=%v, %s=%v, fluid=%s: %w", output, name1, val1, name2, val2, fluidName, err)
		}

	} else if (name1 == "T" && name2 == "H") || (name1 == "H" && name2 == "T") {
		// Case 3: T and H -> solve for D using T-H flash
		var H_target float64
		if name1 == "T" {
			T = val1
			H_target = val2
		} else {
			H_target = val1
			T = val2
		}

		Rho, err = flash.FlashTH(f, T, H_target)
		if err != nil {
			return 0, fmt.Errorf("T-H flash failed: %v", err)
		}

	} else if (name1 == "P" && name2 == "H") || (name1 == "H" && name2 == "P") {
		// Case 4: P and H -> solve for T and D using P-H flash
		var P_target, H_target float64
		if name1 == "P" {
			P_target = val1
			H_target = val2
		} else {
			H_target = val1
			P_target = val2
		}

		T, Rho, err = flash.FlashPH(f, P_target, H_target)
		if err != nil {
			return 0, fmt.Errorf("P-H flash failed: %v", err)
		}

	} else if (name1 == "P" && name2 == "S") || (name1 == "S" && name2 == "P") {
		// Case 5: P and S -> solve for T and D using P-S flash
		var P_target, S_target float64
		if name1 == "P" {
			P_target = val1
			S_target = val2
		} else {
			S_target = val1
			P_target = val2
		}

		T, Rho, err = flash.FlashPS(f, P_target, S_target)
		if err != nil {
			return 0, fmt.Errorf("P-S flash failed: %v", err)
		}

	} else if (name1 == "P" && name2 == "Q") || (name1 == "Q" && name2 == "P") {
		// Case 6: P and Q -> saturated state at this P
		var P_target, Q_target float64
		if name1 == "P" {
			P_target = val1
			Q_target = val2
		} else {
			Q_target = val1
			P_target = val2
		}

		T, err = saturation.Tsat(f, P_target)
		if err != nil {
			return 0, fmt.Errorf("Tsat failed: %v", err)
		}
		knownSatP = P_target
		knownSatT = T
		hasKnownSatP = true
		hasKnownSatT = true

		rhoL, err := saturation.RhoL(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoL failed: %v", err)
		}
		rhoV, err := saturation.RhoV(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoV failed: %v", err)
		}

		if Q_target == 0 {
			Rho = rhoL
		} else if Q_target == 1 {
			Rho = rhoV
		} else {
			return 0, fmt.Errorf("input pair %s,%s with Q=%g is unsupported for output=%s fluid=%s; only saturation endpoints Q=0 and Q=1 are supported", name1, name2, Q_target, output, fluidName)
		}

	} else if (name1 == "T" && name2 == "Q") || (name1 == "Q" && name2 == "T") {
		// Case 7: T and Q -> saturated state at this T
		var Q_target float64
		if name1 == "T" {
			T = val1
			Q_target = val2
		} else {
			Q_target = val1
			T = val2
		}

		rhoL, err := saturation.RhoL(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoL failed: %v", err)
		}
		if pSat, pErr := saturation.Psat(f, T); pErr == nil {
			knownSatP = pSat
			hasKnownSatP = true
		}
		knownSatT = T
		hasKnownSatT = true
		rhoV, err := saturation.RhoV(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoV failed: %v", err)
		}

		if Q_target == 0 {
			Rho = rhoL
		} else if Q_target == 1 {
			Rho = rhoV
		} else {
			return 0, fmt.Errorf("input pair %s,%s with Q=%g is unsupported for output=%s fluid=%s; only saturation endpoints Q=0 and Q=1 are supported", name1, name2, Q_target, output, fluidName)
		}

	} else {
		return 0, fmt.Errorf("input pair %s,%s not supported for output=%s fluid=%s", name1, name2, output, fluidName)
	}

	// Update state with final T, Rho
	state.Update(T, Rho)

	// -------- Outputs --------
	switch output {
	case "T":
		if hasKnownSatT {
			return knownSatT, nil
		}
		return state.T, nil
	case "D", "DMOLAR":
		return state.Rho, nil
	case "P":
		if hasKnownSatP {
			return knownSatP, nil
		}
		return state.Pressure(), nil
	case "S", "SMOLAR":
		return state.MolarEntropy(), nil
	case "H", "HMOLAR":
		return state.MolarEnthalpy(), nil
	case "U", "UMOLAR":
		return state.MolarInternalEnergy(), nil
	case "CV", "CVMOLAR":
		return state.Cv(), nil
	case "CP", "CPMOLAR":
		return state.Cp(), nil
	case "P_SAT":
		if hasKnownSatP {
			return knownSatP, nil
		}
		return saturation.Psat(f, state.T)
	case "T_SAT":
		if hasKnownSatT {
			return knownSatT, nil
		}
		return saturation.Tsat(f, state.Pressure())
	case "Q":
		if state.T >= f.States.Critical.T {
			return 0, fmt.Errorf("output Q undefined above the critical temperature for fluid=%s", fluidName)
		}
		rhoL, err := saturation.RhoL(f, state.T)
		if err != nil {
			return 0, err
		}
		rhoV, err := saturation.RhoV(f, state.T)
		if err != nil {
			return 0, err
		}

		v := 1.0 / state.Rho
		vL := 1.0 / rhoL
		vV := 1.0 / rhoV
		q := (v - vL) / (vV - vL)
		if absRel(state.Rho, rhoL) <= 1e-5 {
			return 0, nil
		}
		if absRel(state.Rho, rhoV) <= 1e-5 {
			return 1, nil
		}
		psat, err := saturation.Psat(f, state.T)
		if err != nil {
			return 0, err
		}
		if absRel(state.Pressure(), psat) > 1e-5 {
			return 0, fmt.Errorf("output Q is defined only on saturation states for fluid=%s", fluidName)
		}
		return 0, fmt.Errorf("output Q for fluid=%s is only supported at saturation endpoints; inferred interior value=%g", fluidName, q)
	case "V", "VISCOSITY":
		return transport.Viscosity(f, state.T, state.Rho)
	case "L", "CONDUCTIVITY":
		return transport.Conductivity(f, state.T, state.Rho)
	case "I", "SURFACE_TENSION":
		return transport.SurfaceTension(f, state.T)
	default:
		return 0, fmt.Errorf("output %s not supported", output)
	}
}

func absRel(a, b float64) float64 {
	scale := maxFloat(1, maxFloat(absFloat(a), absFloat(b)))
	return absFloat(a-b) / scale
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
