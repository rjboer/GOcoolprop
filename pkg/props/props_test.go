package props

import (
	"math"
	"testing"
)

func TestPropSI_Water(t *testing.T) {
	// T=300K, P=101325 Pa
	// Expected Rho ~ 55317

	val, err := PropSI("D", "T", 300, "P", 101325, "Water")
	if err != nil {
		t.Fatalf("PropSI failed: %v", err)
	}

	expected := 55317.0
	if math.Abs(val-expected) > 1000 {
		t.Errorf("Water Density mismatch: got %v, expected %v", val, expected)
	}

	// Check Enthalpy
	h, err := PropSI("H", "T", 300, "P", 101325, "Water")
	if err != nil {
		t.Fatalf("PropSI H failed: %v", err)
	}
	if h == 0 {
		t.Error("Enthalpy is 0")
	}
}

func TestPropSI_Nitrogen(t *testing.T) {
	// Nitrogen at 300K, 1 atm (Gas)
	// Ideal gas density: P/RT = 101325 / (8.314 * 300) = 40.6

	val, err := PropSI("D", "T", 300, "P", 101325, "Nitrogen")
	if err != nil {
		t.Fatalf("PropSI Nitrogen failed: %v", err)
	}

	expected := 40.6
	if math.Abs(val-expected) > 1 {
		t.Errorf("Nitrogen Density mismatch: got %v, expected %v", val, expected)
	}
}

func TestPropSI_Hydrogen(t *testing.T) {
	// Hydrogen at 300K, 1 atm (Gas)

	val, err := PropSI("D", "T", 300, "P", 101325, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI Hydrogen failed: %v", err)
	}

	expected := 40.6
	if math.Abs(val-expected) > 1 {
		t.Errorf("Hydrogen Density mismatch: got %v, expected %v", val, expected)
	}
}
