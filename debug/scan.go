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

	fmt.Printf("Critical: T=%v K, P=%v Pa, Rho=%v mol/m3\n",
		f.States.Critical.T, f.States.Critical.P, f.States.Critical.RhoMolar)

	// Test a range of densities at T=300K
	T := 300.0
	P_target := 101325.0

	fmt.Printf("\nT=%v K, P_target=%v Pa\n", T, P_target)
	fmt.Printf("\nDensity scan:\n")

	densities := []float64{1e-8, 1e-6, 1e-4, 1e-2, 1, 10, 40.6, 100, 1000, 5000, 10000}

	for _, rho := range densities {
		state.Update(T, rho)
		P := state.Pressure()
		fmt.Printf("Rho=%12.6e -> P=%12.6e (error=%12.6e)\n", rho, P, P-P_target)
	}
}
