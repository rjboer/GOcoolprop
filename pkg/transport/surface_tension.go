package transport

import (
	"GOcoolprop/pkg/fluid"
	"fmt"
	"math"
)

// SurfaceTension calculates the surface tension in N/m.
// Only valid at saturation conditions.
func SurfaceTension(f *fluid.FluidData, T float64) (float64, error) {
	// Check if surface tension data exists
	// It might be in TRANSPORT.SurfaceTension OR in ANCILLARIES.SurfaceTension

	// Check Transport first
	st := f.Transport.SurfaceTension
	// Check Ancillaries if Transport is empty (or check if A is populated)
	if len(st.A) == 0 {
		st = f.Ancillaries.SurfaceTension
	}

	if len(st.A) == 0 {
		return 0, fmt.Errorf("surface tension data not found")
	}

	// sigma = sum(a_i * (1 - T/Tc)^n_i)
	Tc := st.Tc
	if Tc == 0 {
		Tc = f.States.Critical.T
	}

	if T > Tc {
		return 0, fmt.Errorf("temperature %v K above critical temperature %v K", T, Tc)
	}

	theta := 1.0 - T/Tc
	sum := 0.0
	for i := range st.A {
		sum += st.A[i] * math.Pow(theta, st.N[i])
	}

	return sum, nil
}
