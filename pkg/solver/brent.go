package solver

import (
	"errors"
	"math"
)

const MachineEpsilon = 2.220446049250313e-16

// Brent finds the root of f(x) = 0 in the interval [a, b] using
// Brent's method. f(a) and f(b) must have opposite signs or one of
// them must be exactly zero.
func Brent(f func(float64) float64, a, b float64, tol float64) (float64, error) {
	fa := f(a)
	fb := f(b)

	// Handle exact roots at the endpoints
	if fa == 0.0 {
		return a, nil
	}
	if fb == 0.0 {
		return b, nil
	}

	// Root must be bracketed
	if fa*fb > 0 {
		return 0, errors.New("root not bracketed")
	}

	// Initialization (Cephes / Numerical Recipes style)
	c := b
	fc := fb
	d := b - a
	e := d

	const maxIter = 100

	for iter := 0; iter < maxIter; iter++ {
		// Ensure that [b, c] brackets the root: f(b) and f(c) have opposite sign.
		if fb*fc > 0 {
			c = a
			fc = fa
			d = b - a
			e = d
		}

		// Make sure that |f(b)| <= |f(c)| so that b is the best current approximation.
		if math.Abs(fc) < math.Abs(fb) {
			// Rotate (a, b, c) and (fa, fb, fc) so that b stays "best"
			a, b, c = b, c, b
			fa, fb, fc = fb, fc, fb
		}

		tol1 := 2.0*MachineEpsilon*math.Abs(b) + 0.5*tol
		xm := 0.5 * (c - b)

		// Convergence test: small bracket or small residual
		if math.Abs(xm) <= tol1 || fb == 0.0 {
			return b, nil
		}

		var s, p, q float64

		// Attempt inverse quadratic interpolation or secant step
		if math.Abs(e) >= tol1 && math.Abs(fa) > math.Abs(fb) {
			s = fb / fa
			if a == c {
				// Secant method
				p = 2.0 * xm * s
				q = 1.0 - s
			} else {
				// Inverse quadratic interpolation
				q = fa / fc
				r := fb / fc
				p = s * (2.0*xm*q*(q-r) - (b-a)*(r-1.0))
				q = (q - 1.0) * (r - 1.0) * (s - 1.0)
			}

			if p > 0.0 {
				q = -q
			}
			p = math.Abs(p)

			min1 := 3.0*xm*q - math.Abs(tol1*q)
			min2 := math.Abs(e * q)

			// Accept interpolation only if it's sufficiently small and safe
			if 2.0*p < math.Min(min1, min2) {
				e = d
				d = p / q
			} else {
				// Otherwise, fall back to bisection
				d = xm
				e = d
			}
		} else {
			// Bisection step
			d = xm
			e = d
		}

		// Move a to b and evaluate new b
		a = b
		fa = fb

		if math.Abs(d) > tol1 {
			b += d
		} else {
			b += math.Copysign(tol1, xm)
		}

		fb = f(b)
	}

	// If you hit this, the function is misbehaving or tol is unrealistically tight.
	return 0, errors.New("Brent: maximum iterations exceeded")
}
