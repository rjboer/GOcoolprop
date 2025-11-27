package solver

import (
	"fmt"
	"math"
)

// Newton2D solves a system of 2 equations with 2 unknowns using Newton-Raphson method.
// funcJS returns the residuals (f1, f2) and the Jacobian matrix elements (J11, J12, J21, J22)
// at a given point (x, y).
//
// The system is:
// f1(x, y) = 0
// f2(x, y) = 0
//
// The Jacobian is:
// J = [ ∂f1/∂x  ∂f1/∂y ] = [ J11 J12 ]
//
//	[ ∂f2/∂x  ∂f2/∂y ]   [ J21 J22 ]
//
// The update step is:
// [ Δx ] = -J^-1 * [ f1 ]
// [ Δy ]           [ f2 ]
func Newton2D(funcJS func(x, y float64) (f1, f2, J11, J12, J21, J22 float64), x0, y0 float64, tol float64, maxIter int) (x, y float64, err error) {
	x = x0
	y = y0

	for i := 0; i < maxIter; i++ {
		f1, f2, J11, J12, J21, J22 := funcJS(x, y)

		// Check convergence on residuals
		if math.Abs(f1) < tol && math.Abs(f2) < tol {
			return x, y, nil
		}

		// Calculate determinant
		det := J11*J22 - J12*J21
		if math.Abs(det) < 1e-20 {
			return x, y, fmt.Errorf("singular Jacobian at iter %d (x=%v, y=%v)", i, x, y)
		}

		// Calculate inverse Jacobian * residuals
		// [ Δx ] = - (1/det) * [  J22  -J12 ] * [ f1 ]
		// [ Δy ]               [ -J21   J11 ]   [ f2 ]

		dx := -(J22*f1 - J12*f2) / det
		dy := -(-J21*f1 + J11*f2) / det

		// Limit step size if needed (damping could be added here)

		x += dx
		y += dy

		// Check for NaN/Inf
		if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
			return x, y, fmt.Errorf("solver diverged to NaN/Inf at iter %d", i)
		}
	}

	return x, y, fmt.Errorf("max iterations (%d) reached without convergence", maxIter)
}
