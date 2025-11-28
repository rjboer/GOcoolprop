package transport

import (
	"GOcoolprop/pkg/fluid"
	"fmt"
	"math"
)

// Conductivity calculates the thermal conductivity in W/(m*K).
func Conductivity(f *fluid.FluidData, T, Rho float64) (float64, error) {
	if f.Transport.Conductivity.Hardcoded != "" {
		return 0, fmt.Errorf("hardcoded conductivity for %s not implemented yet", f.Info.Name)
	}

	// 1. Dilute Gas Contribution
	lambda0, err := ConductivityDilute(f, T)
	if err != nil {
		return 0, err
	}

	// 2. Residual Contribution
	lambdaRes, err := ConductivityResidual(f, T, Rho)
	if err != nil {
		return 0, err
	}

	// 3. Critical Enhancement (Optional/TODO)
	// lambdaCrit := ...

	return lambda0 + lambdaRes, nil
}

func ConductivityDilute(f *fluid.FluidData, T float64) (float64, error) {
	d := f.Transport.Conductivity.Dilute
	if d == nil {
		return 0, nil
	}

	if d.Type == "polynomial_and_exponential" || d.Type == "rational_polynomial" {
		// lambda0 = sum(A_i * T^i) / sum(B_i * T^i)

		num := 0.0
		for i, a := range d.A {
			num += a * math.Pow(T, float64(i))
		}

		den := 1.0
		if len(d.B) > 0 {
			den = 0.0
			for i, b := range d.B {
				den += b * math.Pow(T, float64(i))
func ConductivityResidual(f *fluid.FluidData, T, Rho float64) (float64, error) {
	r := f.Transport.Conductivity.Residual
	if r == nil {
		return 0, nil
	}

	if r.Type == "polynomial_and_exponential" {
		// lambda_res = sum(A_i * tau^t_i * delta^d_i * exp(-gamma_i * delta^l_i))

		Tc := f.States.Critical.T
		Rhoc := f.States.Critical.RhoMolar

		tau := Tc / T
		delta := Rho / Rhoc

		sum := 0.0
		for i := range r.A {
			term := r.A[i]
			term *= math.Pow(tau, r.T[i])
			term *= math.Pow(delta, r.D[i])

			if r.Gamma[i] != 0 {
				term *= math.Exp(-r.Gamma[i] * math.Pow(delta, r.L[i]))
			}
			sum += term
		}

		return sum, nil
	}

	return 0, fmt.Errorf("unknown residual conductivity type: %s", r.Type)
}
