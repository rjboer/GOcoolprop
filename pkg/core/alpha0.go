package core

import (
	"math"
)

// IdealGasHelmholtzLead: alpha = ln(delta) + a1 + a2*tau
type IdealGasHelmholtzLead struct {
	A1 float64
	A2 float64
}

func (t *IdealGasHelmholtzLead) Term(tau, delta float64) float64 {
	return math.Log(delta) + t.A1 + t.A2*tau
}
func (t *IdealGasHelmholtzLead) DDelta(tau, delta float64) float64 {
	return 1.0 / delta
}
func (t *IdealGasHelmholtzLead) DTau(tau, delta float64) float64 {
	return t.A2
}
func (t *IdealGasHelmholtzLead) DDelta2(tau, delta float64) float64 {
	return -1.0 / (delta * delta)
}
func (t *IdealGasHelmholtzLead) DTau2(tau, delta float64) float64 {
	return 0
}
func (t *IdealGasHelmholtzLead) DDeltaTau(tau, delta float64) float64 {
	return 0
}

// IdealGasHelmholtzLogTau: alpha = a * ln(tau)
type IdealGasHelmholtzLogTau struct {
	A float64
}

func (t *IdealGasHelmholtzLogTau) Term(tau, delta float64) float64 {
	return t.A * math.Log(tau)
}
func (t *IdealGasHelmholtzLogTau) DDelta(tau, delta float64) float64 {
	return 0
}
func (t *IdealGasHelmholtzLogTau) DTau(tau, delta float64) float64 {
	return t.A / tau
}
func (t *IdealGasHelmholtzLogTau) DDelta2(tau, delta float64) float64 {
	return 0
}
func (t *IdealGasHelmholtzLogTau) DTau2(tau, delta float64) float64 {
	return -t.A / (tau * tau)
}
func (t *IdealGasHelmholtzLogTau) DDeltaTau(tau, delta float64) float64 {
	return 0
}

// IdealGasHelmholtzPlanckEinstein: alpha = sum(n_i * ln(1 - exp(-t_i * tau)))
type IdealGasHelmholtzPlanckEinstein struct {
	N []float64
	T []float64
}

func (t *IdealGasHelmholtzPlanckEinstein) Term(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		sum += t.N[i] * math.Log(1-math.Exp(-t.T[i]*tau))
	}
	return sum
}
func (t *IdealGasHelmholtzPlanckEinstein) DDelta(tau, delta float64) float64 {
	return 0
}
func (t *IdealGasHelmholtzPlanckEinstein) DTau(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		expVal := math.Exp(-t.T[i] * tau)
		sum += t.N[i] * t.T[i] * expVal / (1 - expVal)
	}
	return sum
}
func (t *IdealGasHelmholtzPlanckEinstein) DDelta2(tau, delta float64) float64 {
	return 0
}
func (t *IdealGasHelmholtzPlanckEinstein) DTau2(tau, delta float64) float64 {
	sum := 0.0
	for i := range t.N {
		expVal := math.Exp(-t.T[i] * tau)
		denom := (1 - expVal)
		sum += -t.N[i] * t.T[i] * t.T[i] * expVal / (denom * denom)
	}
	return sum
}
func (t *IdealGasHelmholtzPlanckEinstein) DDeltaTau(tau, delta float64) float64 {
	return 0
}
