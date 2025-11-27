package main

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"fmt"
)

func main() {
	// Load Nitrogen
	f, err := fluid.LoadFluidByName("Nitrogen", "data")
	if err != nil {
		fmt.Printf("Error loading: %v\n", err)
		return
	}

	state := core.NewState(f)

	T := 300.0
	P_target := 101325.0
	R := f.EOS[0].GasConstant
	rhoIdeal := P_target / (R * T)

	fmt.Printf("Ideal gas estimate: Rho=%v mol/m3\n", rhoIdeal)
	fmt.Printf("Safety factor 1.5: maxRho=%v\n", rhoIdeal*1.5)

	// Test bounds
	minRho := 1e-8
	maxRho := rhoIdeal * 1.5

	state.Update(T, minRho)
	pMin := state.Pressure() - P_target

	state.Update(T, maxRho)
	pMax := state.Pressure() - P_target

	fmt.Printf("\nBounds check:\n")
	fmt.Printf("minRho=%e: P-P_target=%e\n", minRho, pMin)
	fmt.Printf("maxRho=%e: P-P_target=%e\n", maxRho, pMax)
	fmt.Printf("Product: %e (should be negative for bracketing)\n", pMin*pMax)

	// Test at ideal gas density
	state.Update(T, rhoIdeal)
	pIdeal := state.Pressure()
	fmt.Printf("\nAt ideal gas density:\n")
	fmt.Printf("Rho=%v: P=%v (target=%v, error=%v)\n", rhoIdeal, pIdeal, P_target, pIdeal-P_target)
}
