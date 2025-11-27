package props

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/flash"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
	"fmt"
	"strings"
)

func PropSI(output, name1 string, val1 float64, name2 string, val2 float64, fluidName string) (float64, error) {
	// Load fluid
	f, err := fluid.LoadFluidByName(fluidName, "data")
	if err != nil {
		// Try absolute path if running from test
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

	// Identify inputs
	// Case 1: T and D (Density)
	if (name1 == "T" && name2 == "D") || (name1 == "D" && name2 == "T") {
		if name1 == "T" {
			T = val1
			Rho = val2
		} else {
			Rho = val1
			T = val2
		}
	} else if (name1 == "T" && name2 == "P") || (name1 == "P" && name2 == "T") {
		// Case 2: T and P -> Solve for D
		var P_target float64
		if name1 == "T" {
			T = val1
			P_target = val2
		} else {
			P_target = val1
			T = val2
		}

		// Solve for Rho
		// Strategy: Try gas phase first (low density), then liquid phase if needed

		obj := func(rho float64) float64 {
			state.Update(T, rho)
			return state.Pressure() - P_target
		}

		// Get critical pressure to determine phase
		Pc := f.States.Critical.P

		// Try gas phase first (for P < 0.9*Pc)
		if P_target < 0.9*Pc {
			// Gas phase bounds: very low density to ~2x ideal gas estimate
			R := f.EOS[0].GasConstant
			rhoIdeal := P_target / (R * T)

			minRho := 1e-8
			maxRho := rhoIdeal * 1.5 // Safety factor for real gas effects

			pMin := obj(minRho)
			pMax := obj(maxRho)

			if pMin*pMax < 0 {
				// Gas phase root exists
				var err error
				Rho, err = solver.Brent(obj, minRho, maxRho, 0.1)
				if err == nil {
					// Found gas phase solution
					goto solved
				}
			}
		}

		// Try liquid phase
		{
			minRho := f.States.Critical.RhoMolar * 0.8
			maxRho := f.States.TripleLiquid.RhoMolar * 1.2
			if maxRho == 0 {
				maxRho = 60000 // Fallback
			}

			pMin := obj(minRho)
			pMax := obj(maxRho)

			if pMin*pMax < 0 {
				var err error
				Rho, err = solver.Brent(obj, minRho, maxRho, 0.1)
				if err == nil {
					goto solved
				}
			}
		}

		return 0, fmt.Errorf("no solution found for T=%v, P=%v", T, P_target)

	solved:
		// Continue to output
	} else if (name1 == "T" && name2 == "H") || (name1 == "H" && name2 == "T") {
		// Case 3: T and H -> Solve for D using T-H flash
		var H_target float64
		if name1 == "T" {
			T = val1
			H_target = val2
		} else {
			H_target = val1
			T = val2
		}

		var err error
		Rho, err = flash.FlashTH(f, T, H_target)
		if err != nil {
			return 0, fmt.Errorf("T-H flash failed: %v", err)
		}
	} else if (name1 == "P" && name2 == "H") || (name1 == "H" && name2 == "P") {
		// Case 4: P and H -> Solve for T and D using P-H flash
		var P_target, H_target float64
		if name1 == "P" {
			P_target = val1
			H_target = val2
		} else {
			H_target = val1
			P_target = val2
		}

		var err error
		T, Rho, err = flash.FlashPH(f, P_target, H_target)
		if err != nil {
			return 0, fmt.Errorf("P-H flash failed: %v", err)
		}
	} else if (name1 == "P" && name2 == "S") || (name1 == "S" && name2 == "P") {
		// Case 5: P and S -> Solve for T and D using P-S flash
		var P_target, S_target float64
		if name1 == "P" {
			P_target = val1
			S_target = val2
		} else {
			S_target = val1
			P_target = val2
		}

		var err error
		T, Rho, err = flash.FlashPS(f, P_target, S_target)
		if err != nil {
			return 0, fmt.Errorf("P-S flash failed: %v", err)
		}
	} else {
		return 0, fmt.Errorf("input pair %s, %s not supported yet", name1, name2)
	}

	// Update state with final T, Rho
	state.Update(T, Rho)

	// Return requested output
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
	default:
		return 0, fmt.Errorf("output %s not supported", output)
	}
}
