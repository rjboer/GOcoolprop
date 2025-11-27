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

	// Target state: 300K, 1 atm
	T_expected := 300.0
	rho_setup := 40.6

	state := core.NewState(f)
	state.Update(T_expected, rho_setup)
	P_actual := state.Pressure()
	S_target := state.MolarEntropy()

	t.Logf("Target: P=%v Pa, S=%v J/mol/K (at T=%v, rho=%v)", P_actual, S_target, T_expected, rho_setup)

	// FlashPS
	T_calc, Rho_calc, err := FlashPS(f, P_actual, S_target)
	if err != nil {
		t.Fatalf("FlashPS failed: %v", err)
	}

	t.Logf("Result: T=%v K, Rho=%v mol/m³", T_calc, Rho_calc)

	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}

	if math.Abs(Rho_calc-rho_setup) > 0.1 {
		t.Errorf("Density mismatch: got %v, expected %v", Rho_calc, rho_setup)
	}
}

func TestFlashPS_Water_Liquid(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	// Target state: 300K, 10 MPa (compressed liquid)
	T_expected := 300.0
	rho_setup := 55000.0 // approx liquid density

	state := core.NewState(f)
	state.Update(T_expected, rho_setup)
	P_actual := state.Pressure()
	S_target := state.MolarEntropy()

	t.Logf("Target: P=%v Pa, S=%v J/mol/K (at T=%v, rho=%v)", P_actual, S_target, T_expected, rho_setup)

	// FlashPS
	T_calc, Rho_calc, err := FlashPS(f, P_actual, S_target)
	if err != nil {
		t.Fatalf("FlashPS failed: %v", err)
	}

	t.Logf("Result: T=%v K, Rho=%v mol/m³", T_calc, Rho_calc)

	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}

	if math.Abs(Rho_calc-rho_setup) > 100.0 {
		t.Errorf("Density mismatch: got %v, expected %v", Rho_calc, rho_setup)
	}
}
