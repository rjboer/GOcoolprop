package transport

import (
	"GOcoolprop/pkg/fluid"
	"fmt"
	"math"
)

// Viscosity calculates the viscosity in Pa*s.
func Viscosity(f *fluid.FluidData, T, Rho float64) (float64, error) {
	// Check for hardcoded fluids (e.g. Water)
	if f.Transport.Viscosity.Hardcoded != "" {
		return 0, fmt.Errorf("hardcoded viscosity for %s not implemented yet", f.Info.Name)
	}

	// 1. Dilute Gas Contribution
	mu0, err := ViscosityDilute(f, T)
	if err != nil {
		return 0, err
	}

	// 2. Residual / Higher Order Contribution
	muRes, err := ViscosityResidual(f, T, Rho)
	if err != nil {
		return 0, err
	}

	return mu0 + muRes, nil
}

func ViscosityDilute(f *fluid.FluidData, T float64) (float64, error) {
	d := f.Transport.Viscosity.Dilute
	if d == nil {
		return 0, nil // No dilute term?
	}

	if d.Type == "collision_integral" {
		// mu0 = C * sqrt(M*T) / (sigma^2 * Omega)
		// Units:
		// C is from JSON (e.g. 2.66958e-8 for N2)
		// M should be in g/mol (e.g. 28.0134)
		// sigma should be in nm (e.g. 0.3656)
		// Result is in Pa*s (based on C value)

		// Convert M to g/mol
		Mg := d.MolarMass * 1000.0

		// Convert sigma to nm
		sigma_nm := f.Transport.Viscosity.SigmaEta * 1e9

		Tstar := T / f.Transport.Viscosity.EpsilonOverK

		// Omega(T*) = exp(sum(a_i * (ln T*)^i))
		lnT := math.Log(Tstar)
		sum := 0.0
		for i, a := range d.A {
			sum += a * math.Pow(lnT, float64(i))
		}
		omega := math.Exp(sum)

		// Calculate mu0 (appears to be in Pa*s with JSON C value)
		mu0 := d.C * math.Sqrt(Mg*T) / (sigma_nm * sigma_nm * omega)

		return mu0, nil
	}

	return 0, fmt.Errorf("unknown dilute viscosity type: %s", d.Type)
}

func ViscosityResidual(f *fluid.FluidData, T, Rho float64) (float64, error) {
	h := f.Transport.Viscosity.HigherOrder
	if h == nil {
		return 0, nil
	}

	if h.Type == "modified_Batschinski_Hildebrand" {
		// Reference: Lemmon and Jacobsen (2004) for Nitrogen
		// mu_res = sum(a_i * delta^d1_i * tau^t1_i * exp(gamma_i * delta^l_i))

		delta := Rho / h.RhoReduce
		tau := h.TReduce / T

		sum := 0.0
		for i := range h.A {
			term := h.A[i]
			term *= math.Pow(delta, h.D1[i])
			term *= math.Pow(tau, h.T1[i])

			if h.Gamma[i] != 0 {
				term *= math.Exp(h.Gamma[i] * math.Pow(delta, h.L[i]))
			}

			sum += term
		}

		return sum, nil
	}

	return 0, fmt.Errorf("unknown residual viscosity type: %s", h.Type)
}
