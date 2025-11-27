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
	rho := 4.062918e+01

	fmt.Printf("Testing repeatability at Rho=%v:\n\n", rho)

	for i := 0; i < 5; i++ {
		state.Update(T, rho)
		p := state.Pressure()
		residual := p - P_target
		fmt.Printf("Call %d: P=%v, residual=%v\n", i+1, p, residual)
	}
}
