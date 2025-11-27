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

	// Target state: 300K, 1 atm
	T_expected := 300.0
	// P_target := 101325.0 // Unused

	// Calculate H at this state
	// We need rho first
	// Ideal gas rho = P/RT = 101325 / (8.314 * 300) = 40.6
	rho_setup := 40.6
	state := core.NewState(f)
	state.Update(T_expected, rho_setup)
	// Refine P to be exactly P_target by adjusting rho?
	// Actually, let's just use the P calculated from T, rho_setup as our target P.
	P_actual := state.Pressure()
	H_target := state.MolarEnthalpy()

	t.Logf("Target: P=%v Pa, H=%v J/mol (at T=%v, rho=%v)", P_actual, H_target, T_expected, rho_setup)

	// FlashPH
	T_calc, Rho_calc, err := FlashPH(f, P_actual, H_target)
	if err != nil {
		t.Fatalf("FlashPH failed: %v", err)
	}

	t.Logf("Result: T=%v K, Rho=%v mol/m³", T_calc, Rho_calc)

	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}

	if math.Abs(Rho_calc-rho_setup) > 0.1 {
		t.Errorf("Density mismatch: got %v, expected %v", Rho_calc, rho_setup)
	}
}

func TestFlashPH_Water_Liquid(t *testing.T) {
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
	H_target := state.MolarEnthalpy()

	t.Logf("Target: P=%v Pa, H=%v J/mol (at T=%v, rho=%v)", P_actual, H_target, T_expected, rho_setup)

	// FlashPH
	T_calc, Rho_calc, err := FlashPH(f, P_actual, H_target)
	if err != nil {
		t.Fatalf("FlashPH failed: %v", err)
	}

	t.Logf("Result: T=%v K, Rho=%v mol/m³", T_calc, Rho_calc)

	if math.Abs(T_calc-T_expected) > 0.1 {
		t.Errorf("Temperature mismatch: got %v, expected %v", T_calc, T_expected)
	}

	if math.Abs(Rho_calc-rho_setup) > 100.0 { // Liquid density is large, allow larger abs error
		t.Errorf("Density mismatch: got %v, expected %v", Rho_calc, rho_setup)
	}
}
