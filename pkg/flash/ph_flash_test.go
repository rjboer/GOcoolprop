package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestFlashPH_Nitrogen_Gas(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	T_expected := 300.0
	rho_setup := 40.6

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T_expected, rho_setup)
	P_actual := state.Pressure()
	H_target := state.MolarEnthalpy()

	T_calc, rhoCalc, err := FlashPH(f, P_actual, H_target)
	if err != nil {
		t.Fatalf("FlashPH failed: %v", err)
	}
	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}
	if math.Abs(rhoCalc-rho_setup) > 0.1 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoCalc, rho_setup)
	}
}

func TestFlashPH_Water_Liquid(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	T_expected := 300.0
	P_target := 101325.0
	rhoSetup, err := DensityTP(f, T_expected, P_target)
	if err != nil {
		t.Fatalf("DensityTP failed: %v", err)
	}

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T_expected, rhoSetup)
	H_target := state.MolarEnthalpy()

	T_calc, rhoCalc, err := FlashPH(f, P_target, H_target)
	if err != nil {
		t.Fatalf("FlashPH failed: %v", err)
	}
	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}
	if math.Abs(rhoCalc-rhoSetup) > 1.0 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoCalc, rhoSetup)
	}
}
