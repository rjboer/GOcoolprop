package core

type HelmholtzTerm interface {
	Term(tau, delta float64) float64
	DDelta(tau, delta float64) float64
	DTau(tau, delta float64) float64
	DDelta2(tau, delta float64) float64
	DTau2(tau, delta float64) float64
	DDeltaTau(tau, delta float64) float64
}

type HelmholtzEnergy struct {
	Alpha0 []HelmholtzTerm
	AlphaR []HelmholtzTerm
}

func (h *HelmholtzEnergy) Update(tau, delta float64) (a, da_ddelta, da_dtau, d2a_ddelta2, d2a_dtau2, d2a_ddelta_dtau float64) {
	for _, term := range h.Alpha0 {
		a += term.Term(tau, delta)
		da_ddelta += term.DDelta(tau, delta)
		da_dtau += term.DTau(tau, delta)
		d2a_ddelta2 += term.DDelta2(tau, delta)
		d2a_dtau2 += term.DTau2(tau, delta)
		d2a_ddelta_dtau += term.DDeltaTau(tau, delta)
	}
	for _, term := range h.AlphaR {
		a += term.Term(tau, delta)
		da_ddelta += term.DDelta(tau, delta)
		da_dtau += term.DTau(tau, delta)
		d2a_ddelta2 += term.DDelta2(tau, delta)
		d2a_dtau2 += term.DTau2(tau, delta)
		d2a_ddelta_dtau += term.DDeltaTau(tau, delta)
	}
	return
}
