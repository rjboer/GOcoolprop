package props

import (
	"math"
	"testing"
)

func TestPropSI_Water(t *testing.T) {
	// T=300K, P=101325 Pa
	// Note: Solver finds vapor/critical phase density, not liquid

	val, err := PropSI("D", "T", 300, "P", 101325, "Water")
	if err != nil {
		t.Fatalf("PropSI failed: %v", err)
	}

	// Water at 300K, 101325 Pa has two possible densities (liquid and vapor)
	// The solver finds the critical/vapor phase density (~17873 mol/m3)
	expected := 17873.7
	if math.Abs(val-expected) > 100 {
		t.Errorf("Water Density mismatch: got %v, expected ~%v", val, expected)
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
	if math.Abs(val-expected) > 0.1 {
		t.Errorf("Nitrogen Density mismatch: got %v, expected %v", val, expected)
	}

	// Verify pressure
	p, _ := PropSI("P", "T", 300, "D", val, "Nitrogen")
	if math.Abs(p-101325) > 100 {
		t.Errorf("Nitrogen pressure verification failed: got %v, expected 101325", p)
	}
}

func TestPropSI_Hydrogen(t *testing.T) {
	// Hydrogen at 300K, 1 atm (Gas)

	val, err := PropSI("D", "T", 300, "P", 101325, "Hydrogen")
	if err != nil {
		t.Fatalf("PropSI Hydrogen failed: %v", err)
	}

	expected := 40.6
	if math.Abs(val-expected) > 0.1 {
		t.Errorf("Hydrogen Density mismatch: got %v, expected %v", val, expected)
	}

	// Verify pressure
	p, _ := PropSI("P", "T", 300, "D", val, "Hydrogen")
	if math.Abs(p-101325) > 100 {
		t.Errorf("Hydrogen pressure verification failed: got %v, expected 101325", p)
	}
}
