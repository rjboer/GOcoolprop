package transport

import (
	"GOcoolprop/pkg/fluid"
	"math"
	"testing"
)

func TestViscosity_Nitrogen(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	// Test point: 300 K, 1 atm (Gas)
	T := 300.0
	P := 101325.0

	// Calculate density first
	// Ideal gas approx for density setup
	R := f.EOS[0].GasConstant
	rho := P / (R * T) // approx 40.6

	// Refine density using PropSI logic (or just use ideal guess for this test if robust enough)
	// Let's use the ideal guess, the viscosity shouldn't be super sensitive to small rho errors in gas.

	mu, err := Viscosity(f, T, rho)
	if err != nil {
		t.Fatalf("Viscosity failed: %v", err)
	}

	// Expected: ~17.8 microPa*s (0.0000178 Pa*s)
	expected := 1.78e-5

	t.Logf("Nitrogen Viscosity at 300K, 1atm: %v Pa*s (Expected ~%v)", mu, expected)

	if math.Abs(mu-expected)/expected > 0.05 {
		t.Errorf("Viscosity mismatch: got %v, expected %v", mu, expected)
	}
}

func TestConductivity_Nitrogen(t *testing.T) {
	f, err := fluid.LoadFluidByName("Nitrogen", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Nitrogen: %v", err)
	}

	// Test point: 300 K, 1 atm
	T := 300.0
	P := 101325.0
	R := f.EOS[0].GasConstant
	rho := P / (R * T)

	k, err := Conductivity(f, T, rho)
	if err != nil {
		t.Fatalf("Conductivity failed: %v", err)
	}

	// Expected: ~0.026 W/m/K
	expected := 0.026

	t.Logf("Nitrogen Conductivity at 300K, 1atm: %v W/m/K (Expected ~%v)", k, expected)

	if math.IsNaN(k) {
		t.Errorf("Conductivity is NaN")
	}

	if math.Abs(k-expected)/expected > 0.05 {
		t.Errorf("Conductivity mismatch: got %v, expected %v", k, expected)
	}
}

func TestSurfaceTension_Water(t *testing.T) {
	f, err := fluid.LoadFluidByName("Water", "../../data")
	if err != nil {
		t.Fatalf("Failed to load Water: %v", err)
	}

	// Test point: 300 K
	T := 300.0

	sigma, err := SurfaceTension(f, T)
	if err != nil {
		t.Fatalf("SurfaceTension failed: %v", err)
	}

	// Expected: ~0.072 N/m
	expected := 0.07197

	t.Logf("Water Surface Tension at 300K: %v N/m (Expected ~%v)", sigma, expected)

	if math.Abs(sigma-expected)/expected > 0.01 {
		t.Errorf("Surface Tension mismatch: got %v, expected %v", sigma, expected)
	}
}
