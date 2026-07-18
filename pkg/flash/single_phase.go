package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/saturation"
	"GOcoolprop/pkg/solver"
	"fmt"
	"math"
)

type phasePreference int

const (
	phaseAny phasePreference = iota
	phaseVapor
	phaseLiquid
)

func bracketedRoots(minX, maxX float64, n int, value func(float64) float64, tol float64) ([]float64, error) {
	if minX <= 0 || maxX <= minX {
		return nil, fmt.Errorf("invalid scan range [%g, %g]", minX, maxX)
	}

	logMin := math.Log(minX)
	logMax := math.Log(maxX)
	dlog := (logMax - logMin) / float64(n)
	roots := make([]float64, 0, 4)

	addRoot := func(root float64) {
		if root <= 0 || math.IsNaN(root) || math.IsInf(root, 0) {
			return
		}
		for _, existing := range roots {
			if math.Abs(root-existing) <= 1e-7*math.Max(1, math.Abs(existing)) {
				return
			}
		}
		roots = append(roots, root)
	}

	prevX := minX
	prevVal := value(prevX)
	for i := 1; i <= n; i++ {
		x := math.Exp(logMin + float64(i)*dlog)
		val := value(x)
		if math.IsNaN(val) || math.IsInf(val, 0) {
			prevX, prevVal = x, val
			continue
		}
		if val == 0 {
			addRoot(x)
		}
		if !math.IsNaN(prevVal) && !math.IsInf(prevVal, 0) && prevVal*val < 0 {
			root, err := solver.Brent(value, prevX, x, tol)
			if err == nil {
				addRoot(root)
			}
		}
		prevX, prevVal = x, val
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no roots found in [%g, %g]", minX, maxX)
	}
	return roots, nil
}

func inferPhaseTP(fluidData *fluid.FluidData, T, P float64) (phasePreference, error) {
	if T <= 0 || P <= 0 {
		return phaseAny, fmt.Errorf("invalid T,P state T=%g P=%g", T, P)
	}
	if T >= fluidData.States.Critical.T || P >= fluidData.States.Critical.P {
		return phaseAny, nil
	}
	psat, err := saturation.Psat(fluidData, T)
	if err != nil {
		return phaseAny, nil
	}
	if math.Abs(P-psat)/math.Max(1, math.Abs(psat)) <= 1e-5 {
		return phaseAny, fmt.Errorf("state lies on saturation boundary for fluid=%s T=%g P=%g; specify Q=0 or Q=1", fluidData.Info.Name, T, P)
	}
	if P > psat {
		return phaseLiquid, nil
	}
	return phaseVapor, nil
}

func endpointPhase(lowPref, highPref phasePreference, lowVal, highVal, target float64) phasePreference {
	if target <= math.Min(lowVal, highVal) {
		if lowVal <= highVal {
			return lowPref
		}
		return highPref
	}
	if lowVal >= highVal {
		return lowPref
	}
	return highPref
}

func inferPhaseFromSaturationAtPressure(fluidData *fluid.FluidData, P, target float64, property func(*core.State) float64, quantity string) (phasePreference, error) {
	if P <= 0 {
		return phaseAny, fmt.Errorf("invalid pressure %g", P)
	}
	if P >= fluidData.States.Critical.P {
		return phaseAny, nil
	}

	Tsat, err := saturation.Tsat(fluidData, P)
	if err != nil {
		return phaseAny, nil
	}
	rhoL, err := saturation.RhoL(fluidData, Tsat)
	if err != nil {
		return phaseAny, err
	}
	rhoV, err := saturation.RhoV(fluidData, Tsat)
	if err != nil {
		return phaseAny, err
	}
	state, err := core.NewState(fluidData)
	if err != nil {
		return phaseAny, err
	}
	state.Update(Tsat, rhoL)
	propL := property(state)
	state.Update(Tsat, rhoV)
	propV := property(state)
	low := math.Min(propL, propV)
	high := math.Max(propL, propV)
	scale := math.Max(1, math.Max(math.Abs(target), math.Max(math.Abs(propL), math.Abs(propV))))
	if target > low+1e-8*scale && target < high-1e-8*scale {
		return phaseAny, fmt.Errorf("two-phase %s flash unsupported for fluid=%s at P=%g", quantity, fluidData.Info.Name, P)
	}
	return endpointPhase(phaseLiquid, phaseVapor, propL, propV, target), nil
}

func inferPhaseFromSaturationAtTemperature(fluidData *fluid.FluidData, T, target float64, property func(*core.State) float64, quantity string) (phasePreference, error) {
	if T <= 0 {
		return phaseAny, fmt.Errorf("invalid temperature %g", T)
	}
	if T >= fluidData.States.Critical.T {
		return phaseAny, nil
	}

	rhoL, err := saturation.RhoL(fluidData, T)
	if err != nil {
		return phaseAny, err
	}
	rhoV, err := saturation.RhoV(fluidData, T)
	if err != nil {
		return phaseAny, err
	}
	state, err := core.NewState(fluidData)
	if err != nil {
		return phaseAny, err
	}
	state.Update(T, rhoL)
	propL := property(state)
	state.Update(T, rhoV)
	propV := property(state)
	low := math.Min(propL, propV)
	high := math.Max(propL, propV)
	scale := math.Max(1, math.Max(math.Abs(target), math.Max(math.Abs(propL), math.Abs(propV))))
	if target > low+1e-8*scale && target < high-1e-8*scale {
		return phaseAny, fmt.Errorf("two-phase %s flash unsupported for fluid=%s at T=%g", quantity, fluidData.Info.Name, T)
	}
	return endpointPhase(phaseLiquid, phaseVapor, propL, propV, target), nil
}

func densityTPWithPhase(fluidData *fluid.FluidData, T, P float64, pref phasePreference) (float64, error) {
	state, err := core.NewState(fluidData)
	if err != nil {
		return 0, err
	}

	R := fluidData.EOS[0].GasConstant
	ideal := math.Max(1e-9, P/(R*T))
	maxRho := math.Max(ideal*50, math.Max(fluidData.States.TripleLiquid.RhoMolar*3, math.Max(fluidData.States.Critical.RhoMolar*5, state.ReducingRho*5)))
	if maxRho <= ideal {
		maxRho = ideal * 100
	}

	obj := func(rho float64) float64 {
		state.Update(T, rho)
		return state.Pressure() - P
	}

	roots, err := bracketedRoots(math.Max(1e-12, ideal*1e-4), maxRho, 220, obj, 1e-10)
	if err != nil {
		return 0, fmt.Errorf("density solve failed for fluid=%s T=%g P=%g: %w", fluidData.Info.Name, T, P, err)
	}

	chosen := roots[0]
	for _, root := range roots[1:] {
		if pref == phaseLiquid && root > chosen {
			chosen = root
		}
		if pref == phaseVapor && root < chosen {
			chosen = root
		}
		if pref == phaseAny && math.Abs(root-ideal) < math.Abs(chosen-ideal) {
			chosen = root
		}
	}
	return chosen, nil
}

func DensityTP(fluidData *fluid.FluidData, T, P float64) (float64, error) {
	pref, err := inferPhaseTP(fluidData, T, P)
	if err != nil {
		return 0, err
	}
	return densityTPWithPhase(fluidData, T, P, pref)
}

func flashPressureProperty(fluidData *fluid.FluidData, P, target float64, phase phasePreference, property func(*core.State) float64, quantity string) (float64, float64, error) {
	state, err := core.NewState(fluidData)
	if err != nil {
		return 0, 0, err
	}

	minT := fluidData.EOS[0].TTriple
	if minT <= 0 {
		minT = fluidData.States.TripleLiquid.T
	}
	if minT <= 0 {
		minT = 20
	}
	minT *= 1.001

	maxT := fluidData.EOS[0].TMax
	if maxT <= minT {
		maxT = math.Max(fluidData.States.Critical.T*2.5, minT+200)
	}

	propertyAtT := func(T float64) float64 {
		rho, rhoErr := densityTPWithPhase(fluidData, T, P, phase)
		if rhoErr != nil {
			return math.NaN()
		}
		state.Update(T, rho)
		return property(state) - target
	}

	temperatures := make([]float64, 0, 120)
	for i := 0; i <= 120; i++ {
		f := float64(i) / 120.0
		temperatures = append(temperatures, minT+(maxT-minT)*f)
	}

	var a, b float64
	var fa, fb float64
	found := false
	for i := 1; i < len(temperatures); i++ {
		t0 := temperatures[i-1]
		t1 := temperatures[i]
		f0 := propertyAtT(t0)
		f1 := propertyAtT(t1)
		if math.IsNaN(f0) || math.IsNaN(f1) || math.IsInf(f0, 0) || math.IsInf(f1, 0) {
			continue
		}
		if f0 == 0 {
			a, b, fa, fb = t0, t0, f0, f0
			found = true
			break
		}
		if f0*f1 <= 0 {
			a, b, fa, fb = t0, t1, f0, f1
			found = true
			break
		}
	}
	if !found {
		return 0, 0, fmt.Errorf("%s flash failed to bracket temperature for fluid=%s at P=%g", quantity, fluidData.Info.Name, P)
	}

	var T float64
	if a == b || fa == 0 || fb == 0 {
		T = a
	} else {
		T, err = solver.Brent(propertyAtT, a, b, 1e-8)
		if err != nil {
			return 0, 0, fmt.Errorf("%s flash failed to solve temperature for fluid=%s at P=%g: %w", quantity, fluidData.Info.Name, P, err)
		}
	}
	rho, err := densityTPWithPhase(fluidData, T, P, phase)
	if err != nil {
		return 0, 0, err
	}
	return T, rho, nil
}
