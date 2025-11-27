package main

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
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

	fmt.Printf("Target: T=%v K, P=%v Pa\n", T, P_target)
	fmt.Printf("Ideal gas estimate: Rho=%v mol/m3\n\n", rhoIdeal)

	// Define objective function
	obj := func(rho float64) float64 {
		state.Update(T, rho)
		p := state.Pressure()
		residual := p - P_target
		fmt.Printf("  obj(rho=%12.6e) = P=%12.6e, residual=%12.6e\n", rho, p, residual)
		return residual
	}

	// Test bounds
	minRho := 1e-8
	maxRho := rhoIdeal * 1.5

	fmt.Println("Testing bounds:")
	pMin := obj(minRho)
	pMax := obj(maxRho)

	fmt.Printf("\nBounds: [%e, %e]\n", minRho, maxRho)
	fmt.Printf("Residuals: [%e, %e]\n", pMin, pMax)
	fmt.Printf("Product: %e (negative = bracketed)\n\n", pMin*pMax)

	// Call Brent solver
	fmt.Println("Calling Brent solver:")
	result, err := solver.Brent(obj, minRho, maxRho, 0.1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("\nResult: Rho=%v mol/m3\n", result)

		// Verify
		state.Update(T, result)
		pResult := state.Pressure()
		fmt.Printf("Verification: P=%v Pa (error=%v)\n", pResult, pResult-P_target)
	}
}
