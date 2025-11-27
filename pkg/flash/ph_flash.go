package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
	"fmt"
)

// FlashPH solves for Temperature and Density given Pressure and Enthalpy.
// Returns T (K) and Rho (mol/m³).
func FlashPH(fluidData *fluid.FluidData, P_target, H_target float64) (float64, float64, error) {
	state := core.NewState(fluidData)

	// Define the system of equations and Jacobian
	funcJS := func(T, Rho float64) (f1, f2, J11, J12, J21, J22 float64) {
		state.Update(T, Rho)

		// Residuals
		f1 = state.Pressure() - P_target
		f2 = state.MolarEnthalpy() - H_target

		// Jacobian elements
		J11 = state.DPdT()
		J12 = state.DPdRho()
		J21 = state.DHdT() // Cp
		J22 = state.DHdRho()

		return
	}

	// Initial Guess Strategy
	// 1. Assume ideal gas to get initial T and Rho
	R := fluidData.EOS[0].GasConstant

	// Rough guess for T based on H (assuming ideal gas with constant Cp ~ 2.5R or 3.5R)
	// H = Cp*T => T = H/Cp
	// For noble gases Cp=2.5R, for diatomics Cp=3.5R. Let's take 4R as a safe average.
	Cp_guess := 4.0 * R
	T_guess := H_target / Cp_guess
	if T_guess < fluidData.States.TripleLiquid.T {
		T_guess = fluidData.States.TripleLiquid.T * 1.1
	}

	Rho_guess := P_target / (R * T_guess)

	// Refine guess if we are likely liquid
	// If P is high and H is low, we might be liquid.
	// But for now, let's try the Newton solver with this guess.
	// If it fails or diverges, we might need a better guess or a 1D fallback (iterate T, solve rho(P,T), check H).

	// Try 2D Newton
	T, Rho, err := solver.Newton2D(funcJS, T_guess, Rho_guess, 1e-6, 100)
	if err == nil && T > 0 && Rho > 0 {
		return T, Rho, nil
	}

	// If failed, try a liquid-like guess
	// Liquid is incompressible-ish, so Rho ~ Rho_triple_liquid
	Rho_guess = fluidData.States.TripleLiquid.RhoMolar
	if Rho_guess == 0 {
		Rho_guess = fluidData.States.Critical.RhoMolar * 2.5
	}
	// T guess? Maybe saturation T at P? Or just standard T.
	T_guess = 300.0

	T, Rho, err = solver.Newton2D(funcJS, T_guess, Rho_guess, 1e-6, 100)
	if err == nil && T > 0 && Rho > 0 {
		return T, Rho, nil
	}

	return 0, 0, fmt.Errorf("FlashPH failed for P=%v, H=%v", P_target, H_target)
}
