package flash

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestFlashPS_Nitrogen_Gas(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	T_expected := 300.0
	rhoSetup := 40.6

	state, err := core.NewState(f)
	if err != nil {
		t.Fatalf("Failed to build state: %v", err)
	}
	state.Update(T_expected, rhoSetup)
	P_actual := state.Pressure()
	S_target := state.MolarEntropy()

	T_calc, rhoCalc, err := FlashPS(f, P_actual, S_target)
	if err != nil {
		t.Fatalf("FlashPS failed: %v", err)
	}
	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}
	if math.Abs(rhoCalc-rhoSetup) > 0.1 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoCalc, rhoSetup)
	}
}

func TestFlashPS_Water_Liquid(t *testing.T) {
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
	S_target := state.MolarEntropy()

	T_calc, rhoCalc, err := FlashPS(f, P_target, S_target)
	if err != nil {
		t.Fatalf("FlashPS failed: %v", err)
	}
	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}
	if math.Abs(rhoCalc-rhoSetup) > 1.0 {
		t.Errorf("Density mismatch: got %v, expected %v", rhoCalc, rhoSetup)
	}
}
