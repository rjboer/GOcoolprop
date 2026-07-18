package transport

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"fmt"
	"math"
)

// Conductivity calculates the thermal conductivity in W/(m*K).
func Conductivity(f *fluid.FluidData, T, Rho float64) (float64, error) {
	if f.Transport.Conductivity.Type == "ECS" {
		return 0, fmt.Errorf("conductivity model %q for %s is not implemented", f.Transport.Conductivity.Type, f.Info.Name)
	}
	if f.Transport.Conductivity.Hardcoded != "" {
		return 0, fmt.Errorf("hardcoded conductivity for %s not implemented yet", f.Info.Name)
	}

	lambda0, err := ConductivityDilute(f, T)
	if err != nil {
		return 0, err
	}
	lambdaRes, err := ConductivityResidual(f, T, Rho)
	if err != nil {
		return 0, err
	}
	return lambda0 + lambdaRes, nil
}

func ConductivityDilute(f *fluid.FluidData, T float64) (float64, error) {
	d := f.Transport.Conductivity.Dilute
	if d == nil {
		return 0, nil
	}

	switch d.Type {
	case "eta0_and_poly":
		state, err := core.NewState(f)
		if err != nil {
			return 0, err
		}
		state.Update(T, 1e-12)
		mu0, err := ViscosityDilute(f, T)
		if err != nil {
			return 0, err
		}
		eta0MicroPaS := mu0 * 1e6
		sum := 0.0
		for i, a := range d.A {
			if i == 0 {
				sum += a * eta0MicroPaS
				continue
			}
			exponent := 0.0
			if i < len(d.T) {
				exponent = d.T[i]
			}
			sum += a * math.Pow(state.Tau, exponent)
		}
		return sum, nil
	case "polynomial_and_exponential", "rational_polynomial":
		num := 0.0
		for i, a := range d.A {
			num += a * math.Pow(T, float64(i))
		}

		den := 1.0
		if len(d.B) > 0 {
			den = 0.0
			for i, b := range d.B {
				den += b * math.Pow(T, float64(i))
			}
		}
		if den == 0 {
			return 0, fmt.Errorf("zero conductivity denominator for %s", f.Info.Name)
		}
		return num / den, nil
	default:
		return 0, fmt.Errorf("unknown dilute conductivity type: %s", d.Type)
	}
}

func ConductivityResidual(f *fluid.FluidData, T, Rho float64) (float64, error) {
	r := f.Transport.Conductivity.Residual
	if r == nil {
		return 0, nil
	}

	switch r.Type {
	case "polynomial_and_exponential":
		Tc := f.EOS[0].States.Reducing.T
		if Tc == 0 {
			Tc = f.States.Critical.T
		}
		Rhoc := f.EOS[0].States.Reducing.RhoMolar
		if Rhoc == 0 {
			Rhoc = f.States.Critical.RhoMolar
		}
		tau := Tc / T
		delta := Rho / Rhoc

		sum := 0.0
		for i := range r.A {
			term := r.A[i] * math.Pow(tau, r.T[i]) * math.Pow(delta, r.D[i])
			if i < len(r.Gamma) && r.Gamma[i] != 0 {
				l := 0.0
				if i < len(r.L) {
					l = r.L[i]
				}
				term *= math.Exp(-r.Gamma[i] * math.Pow(delta, l))
			}
			sum += term
		}
		return sum, nil
	default:
		return 0, fmt.Errorf("unknown residual conductivity type: %s", r.Type)
	}
}
