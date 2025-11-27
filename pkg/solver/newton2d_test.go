package solver

import (
	"math"
	"testing"
)

func TestNewton2D_SimpleSystem(t *testing.T) {
	// Solve system:
	// x + y - 3 = 0  => f1
	// x - y - 1 = 0  => f2
	// Solution: x=2, y=1

	// Jacobian:
	// J11=1, J12=1
	// J21=1, J22=-1

	funcJS := func(x, y float64) (f1, f2, J11, J12, J21, J22 float64) {
		f1 = x + y - 3
		f2 = x - y - 1
		J11 = 1
		J12 = 1
		J21 = 1
		J22 = -1
		return
	}

	x, y, err := Newton2D(funcJS, 0, 0, 1e-8, 100)
	if err != nil {
		t.Fatalf("Newton2D failed: %v", err)
	}

	if math.Abs(x-2) > 1e-6 || math.Abs(y-1) > 1e-6 {
		t.Errorf("Expected (2, 1), got (%v, %v)", x, y)
	}
}

func TestNewton2D_NonLinear(t *testing.T) {
	// Solve system:
	// x^2 + y^2 - 4 = 0   (Circle radius 2)
	// x - y = 0           (Line y=x)
	// Solutions: (sqrt(2), sqrt(2)) and (-sqrt(2), -sqrt(2))

	funcJS := func(x, y float64) (f1, f2, J11, J12, J21, J22 float64) {
		f1 = x*x + y*y - 4
		f2 = x - y
		J11 = 2 * x
		J12 = 2 * y
		J21 = 1
		J22 = -1
		return
	}

	// Start near positive root
	x, y, err := Newton2D(funcJS, 1, 1, 1e-8, 100)
	if err != nil {
		t.Fatalf("Newton2D failed: %v", err)
	}

	expected := math.Sqrt(2)
	if math.Abs(x-expected) > 1e-6 || math.Abs(y-expected) > 1e-6 {
		t.Errorf("Expected (%v, %v), got (%v, %v)", expected, expected, x, y)
	}
}
