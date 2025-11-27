package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
	"fmt"
)

// FlashPS solves for Temperature and Density given Pressure and Entropy.
// Returns T (K) and Rho (mol/m³).
func FlashPS(fluidData *fluid.FluidData, P_target, S_target float64) (float64, float64, error) {
	state := core.NewState(fluidData)

	// Define the system of equations and Jacobian
	funcJS := func(T, Rho float64) (f1, f2, J11, J12, J21, J22 float64) {
		state.Update(T, Rho)

		// Residuals
		f1 = state.Pressure() - P_target
		f2 = state.MolarEntropy() - S_target

		// Jacobian elements
		J11 = state.DPdT()
		J12 = state.DPdRho()
		J21 = state.DSdT() // Cv/T
		J22 = state.DSdRho()

		return
	}

	// Initial Guess Strategy
	// 1. Assume ideal gas to get initial T and Rho
	R := fluidData.EOS[0].GasConstant

	// Rough guess for T based on S (assuming ideal gas)
	// S = S0 + Cp*ln(T/T0) - R*ln(P/P0)
	// This is hard to invert without reference state.
	// Let's use a standard guess T=300K and refine from there.
	T_guess := 300.0
	Rho_guess := P_target / (R * T_guess)

	// Try 2D Newton
	T, Rho, err := solver.Newton2D(funcJS, T_guess, Rho_guess, 1e-6, 100)
	if err == nil && T > 0 && Rho > 0 {
		return T, Rho, nil
	}

	// If failed, try a liquid-like guess
	Rho_guess = fluidData.States.TripleLiquid.RhoMolar
	if Rho_guess == 0 {
		Rho_guess = fluidData.States.Critical.RhoMolar * 2.5
	}
	T_guess = 300.0

	T, Rho, err = solver.Newton2D(funcJS, T_guess, Rho_guess, 1e-6, 100)
	if err == nil && T > 0 && Rho > 0 {
		return T, Rho, nil
	}

	return 0, 0, fmt.Errorf("FlashPS failed for P=%v, S=%v", P_target, S_target)
}
