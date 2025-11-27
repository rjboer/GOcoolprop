package props

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/flash"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"GOcoolprop/pkg/solver"
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

	state := core.NewState(f)

	var T, Rho float64

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
		// Case 2: T and P -> solve for D

		var P_target float64
		if name1 == "T" {
			T = val1
			P_target = val2
		} else {
			P_target = val1
			T = val2
		}

		// ---- Compressed-liquid shortcut ----
		// If T < Tc and P > Psat(T), we are in compressed liquid region.
		// The EOS struggles here, so approximate by saturated liquid density at T.
		if T < f.States.Critical.T {
			if PsatT, errPsat := saturation.Psat(f, T); errPsat == nil && P_target > PsatT {
				if rhoL, err := saturation.RhoL(f, T); err == nil && rhoL > 0 {
					Rho = rhoL
					goto solved
				}
			}
		}

		// General root: find rho s.t. P(T, rho) = P_target
		obj := func(rho float64) float64 {
			state.Update(T, rho)
			return state.Pressure() - P_target
		}

		Pc := f.States.Critical.P

		// Decide which phase to try based on Tsat(P)
		TsatAtP, errTsat := saturation.Tsat(f, P_target)
		tryGas := true
		tryLiq := true
		if errTsat == nil {
			if T < TsatAtP {
				// subcooled liquid region
				tryGas = false
			} else if T > TsatAtP {
				// superheated gas region
				tryLiq = false
			}
		}

		found := false

		// ---- Gas-phase root (for low pressures) ----
		if tryGas && P_target < 0.9*Pc {
			Rg := f.EOS[0].GasConstant
			rhoIdeal := P_target / (Rg * T)

			minRho := rhoIdeal * 0.1
			if minRho < 1e-8 {
				minRho = 1e-8
			}
			maxRho := rhoIdeal * 3.0

			pMin := obj(minRho)
			pMax := obj(maxRho)

			if pMin*pMax < 0 {
				if rhoGas, err := solver.Brent(obj, minRho, maxRho, 0.1); err == nil {
					Rho = rhoGas
					found = true
				}
			}
		}

		// ---- Liquid-phase root around saturated liquid density ----
		if !found && tryLiq {
			var rhoLGuess float64

			if rhoSat, err := saturation.RhoL(f, T); err == nil && rhoSat > 0 {
				rhoLGuess = rhoSat
			} else if f.States.TripleLiquid.RhoMolar > 0 {
				rhoLGuess = f.States.TripleLiquid.RhoMolar
			} else if f.States.Critical.RhoMolar > 0 {
				rhoLGuess = f.States.Critical.RhoMolar
			} else {
				rhoLGuess = 60000.0
			}

			minRho := rhoLGuess * 0.2
			maxRho := rhoLGuess * 2.0
			if minRho < 1e-3 {
				minRho = 1e-3
			}

			pMin := obj(minRho)
			pMax := obj(maxRho)

			if pMin*pMax < 0 {
				if rhoLiq, err := solver.Brent(obj, minRho, maxRho, 0.1); err == nil {
					Rho = rhoLiq
					found = true
				}
			}
		}

		// ---- Final fallback: wide bracket between critical and triple-liquid ----
		if !found {
			minRho := f.States.Critical.RhoMolar * 0.5
			maxRho := f.States.TripleLiquid.RhoMolar * 1.5
			if maxRho == 0 {
				maxRho = 60000.0
			}
			if minRho <= 0 {
				minRho = 1e-3
			}

			pMin := obj(minRho)
			pMax := obj(maxRho)

			if pMin*pMax < 0 {
				if rhoAny, err := solver.Brent(obj, minRho, maxRho, 0.1); err == nil {
					Rho = rhoAny
					found = true
				}
			}
		}

		if !found {
			return 0, fmt.Errorf("no solution found for T=%v, P=%v", T, P_target)
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

		rhoL, err := saturation.RhoL(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoL failed: %v", err)
		}
		rhoV, err := saturation.RhoV(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoV failed: %v", err)
		}

		if Q_target <= 0 {
			Rho = rhoL
		} else if Q_target >= 1 {
			Rho = rhoV
		} else {
			vL := 1.0 / rhoL
			vV := 1.0 / rhoV
			v := Q_target*vV + (1-Q_target)*vL
			Rho = 1.0 / v
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
		rhoV, err := saturation.RhoV(f, T)
		if err != nil {
			return 0, fmt.Errorf("RhoV failed: %v", err)
		}

		if Q_target <= 0 {
			Rho = rhoL
		} else if Q_target >= 1 {
			Rho = rhoV
		} else {
			vL := 1.0 / rhoL
			vV := 1.0 / rhoV
			v := Q_target*vV + (1-Q_target)*vL
			Rho = 1.0 / v
		}

	} else {
		return 0, fmt.Errorf("input pair %s, %s not supported yet", name1, name2)
	}

solved:
	// Update state with final T, Rho
	state.Update(T, Rho)

	// -------- Outputs --------
	switch output {
	case "T":
		return state.T, nil
	case "D", "DMOLAR":
		return state.Rho, nil
	case "P":
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
		return saturation.Psat(f, state.T)
	case "T_SAT":
		return saturation.Tsat(f, state.Pressure())
	case "Q":
		// Quality Q = (v - vL) / (vV - vL)
		if state.T >= f.States.Critical.T {
			return 0, fmt.Errorf("supercritical, Q undefined")
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

		return (v - vL) / (vV - vL), nil
	default:
		return 0, fmt.Errorf("output %s not supported", output)
	}
}
