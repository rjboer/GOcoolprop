package core

import (
	"math"
)

// ResidualHelmholtzPower: alpha = n * delta^d * tau^t * exp(-delta^l)
// If l == 0, exp term is 1.
type ResidualHelmholtzPower struct {
	N []float64
	D []float64
	T []float64
	L []float64
}

func (t *ResidualHelmholtzPower) Term(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		val := t.N[i] * math.Pow(delta, t.D[i]) * math.Pow(tau, t.T[i])
		if t.L[i] != 0 {
			val *= math.Exp(-math.Pow(delta, t.L[i]))
		}
		sum += val
	}
	return sum
}

func (t *ResidualHelmholtzPower) DDelta(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		n, d, _, l := t.N[i], t.D[i], t.T[i], t.L[i]
		term := n * math.Pow(delta, d-1) * math.Pow(tau, t.T[i])
		if l != 0 {
			expVal := math.Exp(-math.Pow(delta, l))
			// d/ddelta [ delta^d * exp(-delta^l) ] = delta^(d-1) * exp(...) * (d - l * delta^l)
			term *= expVal * (d - l*math.Pow(delta, l))
		} else {
			term *= d
		}

		sum += term
	}
	return sum
}

func (t *ResidualHelmholtzPower) DTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		n, d, ti, l := t.N[i], t.D[i], t.T[i], t.L[i]
		term := n * math.Pow(delta, d) * math.Pow(tau, ti-1) * ti
		if l != 0 {
			term *= math.Exp(-math.Pow(delta, l))
		}
		sum += term
	}
	return sum
}

func (t *ResidualHelmholtzPower) DDelta2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		n, d, _, l := t.N[i], t.D[i], t.T[i], t.L[i]
		if l == 0 {
			sum += n * d * (d - 1) * math.Pow(delta, d-2) * math.Pow(tau, t.T[i])
		} else {
			// Complex derivative
			// f = delta^d * exp(-delta^l)
			// f' = delta^(d-1) * exp * (d - l*delta^l)
			// f'' = ...
			// Let's use a simplified form or careful derivation.
			// f' = delta^(d-1) * exp * d - l * delta^(d+l-1) * exp
			// f'' = d(d-1)delta^(d-2)exp - d*l*delta^(d+l-2)exp - l(d+l-1)delta^(d+l-2)exp + l^2*delta^(d+2l-2)exp
			//     = delta^(d-2) * exp * [ d(d-1) - l(2d+l-1)delta^l + l^2 delta^(2l) ]
			expVal := math.Exp(-math.Pow(delta, l))
			deltaL := math.Pow(delta, l)
			bracket := d*(d-1) - l*(2*d+l-1)*deltaL + l*l*deltaL*deltaL
			sum += n * math.Pow(delta, d-2) * math.Pow(tau, t.T[i]) * expVal * bracket
		}
	}
	return sum
}

func (t *ResidualHelmholtzPower) DTau2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		n, d, ti, l := t.N[i], t.D[i], t.T[i], t.L[i]
		term := n * math.Pow(delta, d) * ti * (ti - 1) * math.Pow(tau, ti-2)
		if l != 0 {
			term *= math.Exp(-math.Pow(delta, l))
		}
		sum += term
	}
	return sum
}

func (t *ResidualHelmholtzPower) DDeltaTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		n, d, ti, l := t.N[i], t.D[i], t.T[i], t.L[i]
		// d/dtau [ d/ddelta ]
		// d/ddelta = delta^(d-1) * tau^t * exp * (d - l*delta^l)
		// d/dtau ... * tau^(t-1) * t
		term := n * math.Pow(delta, d-1) * ti * math.Pow(tau, ti-1)
		if l != 0 {
			expVal := math.Exp(-math.Pow(delta, l))
			term *= expVal * (d - l*math.Pow(delta, l))
		} else {
			term *= d
		}
		sum += term
	}
	return sum
}

// ResidualHelmholtzGaussian: alpha = n * delta^d * tau^t * exp(-eta*(delta-epsilon)^2 - beta*(tau-gamma)^2)
type ResidualHelmholtzGaussian struct {
	N       []float64
	D       []float64
	T       []float64
	Eta     []float64
	Epsilon []float64
	Beta    []float64
	Gamma   []float64
}

func (t *ResidualHelmholtzGaussian) Term(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		deltaDiff := delta - t.Epsilon[i]
		tauDiff := tau - t.Gamma[i]
		expVal := math.Exp(-t.Eta[i]*deltaDiff*deltaDiff - t.Beta[i]*tauDiff*tauDiff)
		sum += t.N[i] * math.Pow(delta, t.D[i]) * math.Pow(tau, t.T[i]) * expVal
	}
	return sum
}

func (t *ResidualHelmholtzGaussian) DDelta(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		// f = delta^d * tau^t * exp(-eta*(delta-eps)^2 - ...)
		// f' = delta^(d-1) * tau^t * exp * [d - 2*eta*delta*(delta-eps)]
		deltaDiff := delta - t.Epsilon[i]
		tauDiff := tau - t.Gamma[i]
		expVal := math.Exp(-t.Eta[i]*deltaDiff*deltaDiff - t.Beta[i]*tauDiff*tauDiff)

		term := t.N[i] * math.Pow(delta, t.D[i]-1) * math.Pow(tau, t.T[i]) * expVal
		bracket := t.D[i] - 2*t.Eta[i]*delta*deltaDiff
		sum += term * bracket
	}
	return sum
}

func (t *ResidualHelmholtzGaussian) DTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		deltaDiff := delta - t.Epsilon[i]
		tauDiff := tau - t.Gamma[i]
		expVal := math.Exp(-t.Eta[i]*deltaDiff*deltaDiff - t.Beta[i]*tauDiff*tauDiff)

		term := t.N[i] * math.Pow(delta, t.D[i]) * math.Pow(tau, t.T[i]-1) * expVal
		bracket := t.T[i] - 2*t.Beta[i]*tau*tauDiff
		sum += term * bracket
	}
	return sum
}

func (t *ResidualHelmholtzGaussian) DDelta2(tau, delta float64) float64 {
	// Approximation or full derivation needed.
	// Let's implement full derivation later if needed, or now.
	// It's just math.
	// f' = A * exp * [d - 2*eta*delta*(delta-eps)]
	// f'' = ...
	// For now, return 0 to compile, but I should implement it.
	// I'll leave it as TODO or implement a simple numerical diff if lazy, but better to do analytic.
	// Given time constraints, I will implement it properly.

	sum := 0.0
	for i := range t.N {
		d, eta, eps := t.D[i], t.Eta[i], t.Epsilon[i]
		deltaDiff := delta - eps
		tauDiff := tau - t.Gamma[i]
		expVal := math.Exp(-eta*deltaDiff*deltaDiff - t.Beta[i]*tauDiff*tauDiff)

		// f = delta^d * ...
		// f_delta = f/delta * (d - 2*eta*delta*deltaDiff)
		// f_delta2 = f/delta^2 * [ (d - 2*eta*delta*deltaDiff)^2 - d - 2*eta*delta^2 + (d - 2*eta*delta*deltaDiff) * (-1) ?? No ]

		// Let's use the recurrence:
		// f_d = f * (d/delta - 2*eta*deltaDiff)
		// f_dd = f_d * (d/delta - 2*eta*deltaDiff) + f * (-d/delta^2 - 2*eta)

		term := t.N[i] * math.Pow(delta, d) * math.Pow(tau, t.T[i]) * expVal
		bracket1 := d/delta - 2*eta*deltaDiff
		bracket2 := -d/(delta*delta) - 2*eta

		sum += term * (bracket1*bracket1 + bracket2)
	}
	return sum
}

func (t *ResidualHelmholtzGaussian) DTau2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		ti, beta, gamma := t.T[i], t.Beta[i], t.Gamma[i]
		tauDiff := tau - gamma
		deltaDiff := delta - t.Epsilon[i]
		expVal := math.Exp(-t.Eta[i]*deltaDiff*deltaDiff - beta*tauDiff*tauDiff)

		term := t.N[i] * math.Pow(delta, t.D[i]) * math.Pow(tau, ti) * expVal
		bracket1 := ti/tau - 2*beta*tauDiff
		bracket2 := -ti/(tau*tau) - 2*beta

		sum += term * (bracket1*bracket1 + bracket2)
	}
	return sum
}

func (t *ResidualHelmholtzGaussian) DDeltaTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		d, eta, eps := t.D[i], t.Eta[i], t.Epsilon[i]
		ti, beta, gamma := t.T[i], t.Beta[i], t.Gamma[i]
		deltaDiff := delta - eps
		tauDiff := tau - gamma
		expVal := math.Exp(-eta*deltaDiff*deltaDiff - beta*tauDiff*tauDiff)

		term := t.N[i] * math.Pow(delta, d) * math.Pow(tau, ti) * expVal
		bracketDelta := d/delta - 2*eta*deltaDiff
		bracketTau := ti/tau - 2*beta*tauDiff

		sum += term * bracketDelta * bracketTau
	}
	return sum
}

// ResidualHelmholtzNonAnalytic matches the CoolProp non-analytic critical term.
type ResidualHelmholtzNonAnalytic struct {
	N    []float64
	A    []float64
	B    []float64
	Beta []float64
	ABig []float64
	C    []float64
	DBig []float64
}

func nonAnalyticShift(tauIn, deltaIn float64) (tau, delta float64) {
	tau, delta = tauIn, deltaIn
	eps := 10 * math.SmallestNonzeroFloat64
	if math.Abs(tau-1) < eps {
		tau = 1 + eps
	}
	if math.Abs(delta-1) < eps {
		delta = 1 + eps
	}
	return tau, delta
}

func (t *ResidualHelmholtzNonAnalytic) eval(i int, tauIn, deltaIn float64) (term, dDelta, dTau, dDelta2, dTau2, dDeltaTau float64) {
	tau, delta := nonAnalyticShift(tauIn, deltaIn)
	n := t.N[i]
	ai := t.A[i]
	bi := t.B[i]
	betai := t.Beta[i]
	Ai := t.ABig[i]
	Ci := t.C[i]
	Di := t.DBig[i]

	dm1 := delta - 1.0
	tm1 := tau - 1.0
	dm1sq := dm1 * dm1
	absPow := math.Pow(dm1sq, 1.0/(2.0*betai))

	theta := (1.0 - tau) + Ai*absPow
	dthetaTau := -1.0
	dthetaDelta := Ai / betai * math.Pow(dm1sq, 1.0/(2.0*betai)-1.0) * dm1
	d2thetaDelta2 := Ai / betai * (1.0/betai - 1.0) * math.Pow(dm1sq, 1.0/(2.0*betai)-1.0)

	psi := math.Exp(-Ci*dm1sq - Di*tm1*tm1)
	dpsiDeltaOverPsi := -2.0 * Ci * dm1
	dpsiDelta := dpsiDeltaOverPsi * psi
	dpsiTauOverPsi := -2.0 * Di * tm1
	dpsiTau := dpsiTauOverPsi * psi
	d2psiDelta2OverPsi := (2.0*Ci*dm1sq - 1.0) * 2.0 * Ci
	d2psiDelta2 := d2psiDelta2OverPsi * psi
	d2psiTau2 := (2.0*Di*tm1*tm1 - 1.0) * 2.0 * Di * psi
	d2psiDeltaTau := dpsiDelta * dpsiTauOverPsi

	deltaTerm := theta*theta + t.B[i]*math.Pow(dm1sq, ai)
	dDeltaTauTerm := 2.0 * theta * dthetaTau
	dDeltaDeltaTerm := 2.0*theta*dthetaDelta + 2.0*t.B[i]*ai*math.Pow(dm1sq, ai-1.0)*dm1
	d2DeltaTau2Term := 2.0
	d2DeltaDeltaTauTerm := 2.0 * dthetaTau * dthetaDelta
	d2DeltaDelta2Term := 2.0 * (theta*d2thetaDelta2 + dthetaDelta*dthetaDelta + t.B[i]*(2.0*ai*ai-ai)*math.Pow(dm1sq, ai-1.0))

	deltaPowBi := math.Pow(deltaTerm, bi)
	dDeltaPowBiDelta := bi * math.Pow(deltaTerm, bi-1.0) * dDeltaDeltaTerm
	dDeltaPowBiTau := bi * math.Pow(deltaTerm, bi-1.0) * dDeltaTauTerm
	d2DeltaPowBiDelta2 := bi * (math.Pow(deltaTerm, bi-1.0)*d2DeltaDelta2Term + (bi-1.0)*math.Pow(deltaTerm, bi-2.0)*dDeltaDeltaTerm*dDeltaDeltaTerm)
	d2DeltaPowBiTau2 := bi*math.Pow(deltaTerm, bi-1.0)*d2DeltaTau2Term + bi*(bi-1.0)*math.Pow(deltaTerm, bi-2.0)*dDeltaTauTerm*dDeltaTauTerm
	d2DeltaPowBiDeltaTau := bi * (math.Pow(deltaTerm, bi-1.0)*d2DeltaDeltaTauTerm + (bi-1.0)*math.Pow(deltaTerm, bi-2.0)*dDeltaTauTerm*dDeltaDeltaTerm)

	term = n * delta * deltaPowBi * psi
	dTau = n * delta * (deltaPowBi*dpsiTau + dDeltaPowBiTau*psi)
	dDelta = n * (deltaPowBi*(psi+delta*dpsiDelta) + dDeltaPowBiDelta*delta*psi)
	dTau2 = n * delta * (d2DeltaPowBiTau2*psi + 2.0*dDeltaPowBiTau*dpsiTau + deltaPowBi*d2psiTau2)
	dDeltaTau = n * (deltaPowBi*(dpsiTau+delta*d2psiDeltaTau) +
		delta*dDeltaPowBiDelta*dpsiTau +
		dDeltaPowBiTau*(psi+delta*dpsiDelta) +
		d2DeltaPowBiDeltaTau*delta*psi)
	dDelta2 = n * (deltaPowBi*(2.0*dpsiDelta+delta*d2psiDelta2) +
		2.0*dDeltaPowBiDelta*(psi+delta*dpsiDelta) +
		d2DeltaPowBiDelta2*delta*psi)
	return
}

func (t *ResidualHelmholtzNonAnalytic) Term(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		term, _, _, _, _, _ := t.eval(i, tau, delta)
		sum += term
	}
	return sum
}

func (t *ResidualHelmholtzNonAnalytic) DDelta(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		_, dd, _, _, _, _ := t.eval(i, tau, delta)
		sum += dd
	}
	return sum
}

func (t *ResidualHelmholtzNonAnalytic) DTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		_, _, dt, _, _, _ := t.eval(i, tau, delta)
		sum += dt
	}
	return sum
}

func (t *ResidualHelmholtzNonAnalytic) DDelta2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		_, _, _, dd2, _, _ := t.eval(i, tau, delta)
		sum += dd2
	}
	return sum
}

func (t *ResidualHelmholtzNonAnalytic) DTau2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		_, _, _, _, dt2, _ := t.eval(i, tau, delta)
		sum += dt2
	}
	return sum
}

func (t *ResidualHelmholtzNonAnalytic) DDeltaTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		_, _, _, _, _, ddt := t.eval(i, tau, delta)
		sum += ddt
	}
	return sum
}
