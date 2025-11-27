package core

import (
	"GOcoolprop/pkg/fluid"
)

type State struct {
	Fluid *fluid.FluidData
	HE    *HelmholtzEnergy

	T   float64
	Rho float64
	P   float64

	// Cache
	Tau   float64
	Delta float64

	// Derivatives
	Alpha         float64
	DaDDelta      float64
	DaDTau        float64
	D2aDDelta2    float64
	D2aDTau2      float64
	D2aDDeltaDTau float64
}

func NewState(f *fluid.FluidData) *State {
	// Build HelmholtzEnergy from FluidData
	he := &HelmholtzEnergy{}

	// Alpha0
	for _, term := range f.EOS[0].Alpha0 {
		switch term.Type {
		case "IdealGasHelmholtzLead":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzLead{A1: term.A1, A2: term.A2})
		case "IdealGasHelmholtzLogTau":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzLogTau{A: term.A})
		case "IdealGasHelmholtzPlanckEinstein":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzPlanckEinstein{N: term.N, T: term.T})
		}
	}

	// AlphaR
	for _, term := range f.EOS[0].AlphaR {
		switch term.Type {
		case "ResidualHelmholtzPower":
			// Handle L if missing (default 0)
			l := term.L
			if len(l) == 0 {
				l = make([]float64, len(term.N))
			}
			he.AlphaR = append(he.AlphaR, &ResidualHelmholtzPower{N: term.N, D: term.D, T: term.T, L: l})
		case "ResidualHelmholtzGaussian":
			he.AlphaR = append(he.AlphaR, &ResidualHelmholtzGaussian{
				N: term.N, D: term.D, T: term.T,
				Eta: term.Eta, Epsilon: term.Epsilon, Beta: term.Beta, Gamma: term.Gamma,
			})
		}
	}

	return &State{Fluid: f, HE: he}
}

func (s *State) Update(T, Rho float64) {
	s.T = T
	s.Rho = Rho

	// Critical values
	Tc := s.Fluid.EOS[0].States.Critical.T
	if Tc == 0 {
		// Fallback if not in EOS.States
		Tc = s.Fluid.States.Critical.T
	}
	Rhoc := s.Fluid.EOS[0].States.Critical.RhoMolar
	if Rhoc == 0 {
		Rhoc = s.Fluid.States.Critical.RhoMolar
	}

	s.Tau = Tc / T
	s.Delta = Rho / Rhoc

	s.Alpha, s.DaDDelta, s.DaDTau, s.D2aDDelta2, s.D2aDTau2, s.D2aDDeltaDTau = s.HE.Update(s.Tau, s.Delta)

	// Calculate P immediately? Or on demand.
	// Let's calculate P.
	R := s.Fluid.EOS[0].GasConstant
	// P = rho * R * T * (1 + delta * alphar_delta)
	// Note: DaDDelta contains both alpha0 and alphar derivatives.
	// alpha0_delta is 1/delta.
	// So alpha_delta = 1/delta + alphar_delta
	// alphar_delta = alpha_delta - 1/delta
	// 1 + delta * alphar_delta = 1 + delta * (alpha_delta - 1/delta) = delta * alpha_delta

	// Wait, let's verify alpha0_delta.
	// Lead: 1/delta. LogTau: 0. Planck: 0.
	// So yes, alpha0_delta = 1/delta.

	s.P = s.Rho * R * s.T * s.Delta * s.DaDDelta
}

func (s *State) Pressure() float64 {
	return s.P
}

func (s *State) MolarEntropy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	// S = R * (tau * alpha_tau - alpha)
	return R * (s.Tau*s.DaDTau - s.Alpha)
}

func (s *State) MolarEnthalpy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	// H = R * T * (tau * alpha_tau + delta * alpha_delta)
	return R * s.T * (s.Tau*s.DaDTau + s.Delta*s.DaDDelta)
}

func (s *State) MolarInternalEnergy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	// U = R * T * tau * alpha_tau
	return R * s.T * s.Tau * s.DaDTau
}

func (s *State) Cv() float64 {
	R := s.Fluid.EOS[0].GasConstant
	// Cv = -R * tau^2 * alpha_tau2
	return -R * s.Tau * s.Tau * s.D2aDTau2
}

func (s *State) Cp() float64 {
	R := s.Fluid.EOS[0].GasConstant
	// Cp = Cv + R * (1 + delta*alphar_delta - delta*tau*alphar_delta_tau)^2 / (1 + 2*delta*alphar_delta + delta^2*alphar_delta2)

	// We need alphar derivatives.
	// alphar_delta = alpha_delta - 1/delta
	// alphar_delta2 = alpha_delta2 - (-1/delta^2) = alpha_delta2 + 1/delta^2
	// alphar_delta_tau = alpha_delta_tau - 0 = alpha_delta_tau

	ar_d := s.DaDDelta - 1.0/s.Delta
	ar_d2 := s.D2aDDelta2 + 1.0/(s.Delta*s.Delta)
	ar_dt := s.D2aDDeltaDTau

	num := 1 + s.Delta*ar_d - s.Delta*s.Tau*ar_dt
	den := 1 + 2*s.Delta*ar_d + s.Delta*s.Delta*ar_d2

	return s.Cv() + R*num*num/den
}

// Property derivatives for flash algorithms

// DPdT returns ∂P/∂T at constant ρ
func (s *State) DPdT() float64 {
	R := s.Fluid.EOS[0].GasConstant
	Tc := s.Fluid.EOS[0].States.Critical.T
	if Tc == 0 {
		Tc = s.Fluid.States.Critical.T
	}

	ar_dt := s.D2aDDeltaDTau

	// ∂P/∂T = P/T - ρRT·δ·αʳ_δτ·Tc/T²
	return s.P/s.T - s.Rho*R*s.T*s.Delta*ar_dt*Tc/(s.T*s.T)
}

// DPdRho returns ∂P/∂ρ at constant T
func (s *State) DPdRho() float64 {
	R := s.Fluid.EOS[0].GasConstant
	Rhoc := s.Fluid.EOS[0].States.Critical.RhoMolar
	if Rhoc == 0 {
		Rhoc = s.Fluid.States.Critical.RhoMolar
	}

	// P = ρRT·δ·α_δ
	// ∂P/∂ρ = RT·δ·α_δ + ρRT·∂(δ·α_δ)/∂ρ
	// ∂(δ·α_δ)/∂ρ = ∂(δ·α_δ)/∂δ · ∂δ/∂ρ = (α_δ + δ·α_δδ) · (1/ρc)

	return R*s.T*s.Delta*s.DaDDelta + s.Rho*R*s.T*(s.DaDDelta+s.Delta*s.D2aDDelta2)/Rhoc
}

// DHdT returns ∂H/∂T at constant ρ
func (s *State) DHdT() float64 {
	// This is actually Cp!
	return s.Cp()
}

// DHdRho returns ∂H/∂ρ at constant T
func (s *State) DHdRho() float64 {
	// H = RT(τ·α_τ + δ·α_δ)
	// ∂H/∂ρ = RT·∂(δ·α_δ)/∂ρ = RT·(α_δ + δ·α_δδ)·(1/ρc)

	R := s.Fluid.EOS[0].GasConstant
	Rhoc := s.Fluid.EOS[0].States.Critical.RhoMolar
	if Rhoc == 0 {
		Rhoc = s.Fluid.States.Critical.RhoMolar
	}

	return R * s.T * (s.DaDDelta + s.Delta*s.D2aDDelta2) / Rhoc
}

// DSdT returns ∂S/∂T at constant ρ
func (s *State) DSdT() float64 {
	// S = R(τ·α_τ - α)
	// ∂S/∂T = R·∂(τ·α_τ - α)/∂T
	// ∂(τ·α_τ)/∂T = (α_τ + τ·α_ττ)·(-Tc/T²)
	// ∂α/∂T = α_τ·(-Tc/T²)
	// ∂S/∂T = R·[-(α_τ + τ·α_ττ)·Tc/T² + α_τ·Tc/T²] = -R·τ·α_ττ·Tc/T²
	// = -R·τ²·α_ττ/T = Cv/T

	return s.Cv() / s.T
}

// DSdRho returns ∂S/∂ρ at constant T
func (s *State) DSdRho() float64 {
	// S = R(τ·α_τ - α)
	// ∂S/∂ρ = R·∂(τ·α_τ - α)/∂ρ
	// ∂(τ·α_τ)/∂ρ = τ·α_τδ·(1/ρc)
	// ∂α/∂ρ = α_δ·(1/ρc)
	// ∂S/∂ρ = R·(τ·α_τδ - α_δ)/ρc

	R := s.Fluid.EOS[0].GasConstant
	Rhoc := s.Fluid.EOS[0].States.Critical.RhoMolar
	if Rhoc == 0 {
		Rhoc = s.Fluid.States.Critical.RhoMolar
	}

	return R * (s.Tau*s.D2aDDeltaDTau - s.DaDDelta) / Rhoc
}
