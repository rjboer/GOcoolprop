package solver

import (
	"errors"
	"math"
)

const MachineEpsilon = 2.220446049250313e-16

// Brent finds the root of f(x) = 0 in the interval [a, b].
func Brent(f func(float64) float64, a, b float64, tol float64) (float64, error) {
	fa := f(a)
	fb := f(b)

	if fa*fb > 0 {
		return 0, errors.New("root not bracketed")
	}

	c := a
	fc := fa
	d := b - a
	e := d

	for {
		// Ensure fb is closer to zero than fc
		if math.Abs(fc) < math.Abs(fb) {
			// Swap a and b
			a, b = b, a
			fa, fb = fb, fa
		}

		tol1 := 2.0*MachineEpsilon*math.Abs(b) + 0.5*tol
		xm := 0.5 * (c - b)

		// Check convergence
		if math.Abs(xm) <= tol1 || math.Abs(fb) < tol {
			return b, nil
		}

		// Decide between interpolation and bisection
		if math.Abs(e) >= tol1 && math.Abs(fa) > math.Abs(fb) {
			s := fb / fa
			var p, q float64
			if a == c {
				// Linear interpolation
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

			// Accept interpolation if it's good enough
			if 2.0*p < math.Min(min1, min2) {
				e = d
				d = p / q
			} else {
				// Fall back to bisection
				d = xm
				e = d
			}
		} else {
			// Bisection
			d = xm
			e = d
		}

		// Update a and fa to previous b and fb
		a = b
		fa = fb

		// Compute new b
		if math.Abs(d) > tol1 {
			b += d
		} else {
			b += math.Copysign(tol1, xm)
		}

		// Evaluate function at new b
		fb = f(b)
	}
}
