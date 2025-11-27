package saturation

import (
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/solver"
	"fmt"
)

// Psat returns the saturation pressure at temperature T.
func Psat(f *fluid.FluidData, T float64) (float64, error) {
	// Check bounds
	if T < f.Ancillaries.PS.TMin || T > f.Ancillaries.PS.TMax {
		// Allow small tolerance
		if T < f.Ancillaries.PS.TMin-0.1 || T > f.Ancillaries.PS.TMax+0.1 {
			return 0, fmt.Errorf("temperature %v K out of range for Psat [%v, %v]", T, f.Ancillaries.PS.TMin, f.Ancillaries.PS.TMax)
		}
	}

	return f.Ancillaries.PS.Evaluate(T), nil
}

// Tsat returns the saturation temperature at pressure P.
func Tsat(f *fluid.FluidData, P float64) (float64, error) {
	// Inverse of Psat(T) = P
	// Objective: Psat(T) - P = 0

	// Bounds for T
	minT := f.Ancillaries.PS.TMin
	maxT := f.Ancillaries.PS.TMax

	// Check if P is within range
	minP := f.Ancillaries.PS.Evaluate(minT)
	maxP := f.Ancillaries.PS.Evaluate(maxT)

	if P < minP || P > maxP {
		// Allow small tolerance
		if P < minP*0.99 || P > maxP*1.01 {
			return 0, fmt.Errorf("pressure %v Pa out of range for Tsat [%v, %v]", P, minP, maxP)
		}
	}

	obj := func(T float64) float64 {
		return f.Ancillaries.PS.Evaluate(T) - P
	}

	// Solve for T
	T, err := solver.Brent(obj, minT, maxT, 1e-6)
	if err != nil {
		return 0, fmt.Errorf("failed to solve for Tsat: %v", err)
	}

	return T, nil
}

// RhoL returns the saturated liquid density at temperature T.
func RhoL(f *fluid.FluidData, T float64) (float64, error) {
	return f.Ancillaries.RhoL.Evaluate(T), nil
}

// RhoV returns the saturated vapor density at temperature T.
func RhoV(f *fluid.FluidData, T float64) (float64, error) {
	return f.Ancillaries.RhoV.Evaluate(T), nil
}
