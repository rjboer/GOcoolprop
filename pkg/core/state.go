package core

import (
	"GOcoolprop/pkg/fluid"
	"fmt"
	"math"
)

type State struct {
	Fluid *fluid.FluidData
	HE    *HelmholtzEnergy

	T   float64
	Rho float64
	P   float64

	Tau   float64
	Delta float64

	ReducingT   float64
	ReducingRho float64

	Alpha0         float64
	Alpha0DDelta   float64
	Alpha0DTau     float64
	Alpha0DDelta2  float64
	Alpha0DTau2    float64
	Alpha0DDeltaTau float64

	AlphaR         float64
	AlphaRDDelta   float64
	AlphaRDTau     float64
	AlphaRDDelta2  float64
	AlphaRDTau2    float64
	AlphaRDDeltaTau float64

	Alpha         float64
	DaDDelta      float64
	DaDTau        float64
	D2aDDelta2    float64
	D2aDTau2      float64
	D2aDDeltaDTau float64
}

func NewState(f *fluid.FluidData) (*State, error) {
	if f == nil {
		return nil, fmt.Errorf("nil fluid")
	}
	if len(f.EOS) == 0 {
		return nil, fmt.Errorf("fluid %s has no EOS data", f.Info.Name)
	}

	he := &HelmholtzEnergy{}
	eos := f.EOS[0]

	for _, term := range eos.Alpha0 {
		switch term.Type {
		case "IdealGasHelmholtzLead":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzLead{A1: term.A1, A2: term.A2})
		case "IdealGasHelmholtzLogTau":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzLogTau{A: term.A})
		case "IdealGasHelmholtzPlanckEinstein":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzPlanckEinstein{N: term.N, T: term.T})
		case "IdealGasHelmholtzPower":
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzPower{N: term.N, T: term.T})
		case "IdealGasHelmholtzPlanckEinsteinFunctionT":
			if term.Tcrit == 0 {
				return nil, fmt.Errorf("fluid %s term %s missing Tcrit", f.Info.Name, term.Type)
			}
			theta := make([]float64, len(term.V))
			c := make([]float64, len(theta))
			d := make([]float64, len(theta))
			for i := range theta {
				theta[i] = -term.V[i] / term.Tcrit
				c[i] = 1
				d[i] = -1
			}
			he.Alpha0 = append(he.Alpha0, &IdealGasHelmholtzPlanckEinsteinGeneralized{N: term.N, Theta: theta, C: c, D: d})
		default:
			return nil, fmt.Errorf("fluid %s has unsupported alpha0 term %s", f.Info.Name, term.Type)
		}
	}

	for _, term := range eos.AlphaR {
		switch term.Type {
		case "ResidualHelmholtzPower":
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
		case "ResidualHelmholtzNonAnalytic":
			he.AlphaR = append(he.AlphaR, &ResidualHelmholtzNonAnalytic{
				N: term.N, A: term.A, B: term.B, Beta: term.Beta, ABig: term.ABig, C: term.C, DBig: term.DBig,
			})
		default:
			return nil, fmt.Errorf("fluid %s has unsupported alphar term %s", f.Info.Name, term.Type)
		}
	}

	reducingT := eos.States.Reducing.T
	reducingRho := eos.States.Reducing.RhoMolar
	if reducingT == 0 {
		reducingT = eos.States.Critical.T
	}
	if reducingT == 0 {
		reducingT = f.States.Critical.T
	}
	if reducingRho == 0 {
		reducingRho = eos.States.Critical.RhoMolar
	}
	if reducingRho == 0 {
		reducingRho = f.States.Critical.RhoMolar
	}
	if reducingT <= 0 || reducingRho <= 0 {
		return nil, fmt.Errorf("fluid %s is missing valid reducing state", f.Info.Name)
	}

	return &State{
		Fluid:       f,
		HE:          he,
		ReducingT:   reducingT,
		ReducingRho: reducingRho,
	}, nil
}

func (s *State) Update(T, Rho float64) {
	s.T = T
	s.Rho = Rho

	if T <= 0 || Rho <= 0 {
		s.Tau = math.NaN()
		s.Delta = math.NaN()
		s.P = math.NaN()
		return
	}

	s.Tau = s.ReducingT / T
	s.Delta = Rho / s.ReducingRho

	s.Alpha0, s.Alpha0DDelta, s.Alpha0DTau, s.Alpha0DDelta2, s.Alpha0DTau2, s.Alpha0DDeltaTau = updateTerms(s.HE.Alpha0, s.Tau, s.Delta)
	s.AlphaR, s.AlphaRDDelta, s.AlphaRDTau, s.AlphaRDDelta2, s.AlphaRDTau2, s.AlphaRDDeltaTau = updateTerms(s.HE.AlphaR, s.Tau, s.Delta)

	s.Alpha = s.Alpha0 + s.AlphaR
	s.DaDDelta = s.Alpha0DDelta + s.AlphaRDDelta
	s.DaDTau = s.Alpha0DTau + s.AlphaRDTau
	s.D2aDDelta2 = s.Alpha0DDelta2 + s.AlphaRDDelta2
	s.D2aDTau2 = s.Alpha0DTau2 + s.AlphaRDTau2
	s.D2aDDeltaDTau = s.Alpha0DDeltaTau + s.AlphaRDDeltaTau

	R := s.Fluid.EOS[0].GasConstant
	s.P = s.Rho * R * s.T * (1.0 + s.Delta*s.AlphaRDDelta)
}

func (s *State) Pressure() float64 {
	return s.P
}

func (s *State) MolarEntropy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * (s.Tau*(s.Alpha0DTau+s.AlphaRDTau) - (s.Alpha0 + s.AlphaR))
}

func (s *State) MolarEnthalpy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * s.T * (1.0 + s.Tau*(s.Alpha0DTau+s.AlphaRDTau) + s.Delta*s.AlphaRDDelta)
}

func (s *State) MolarInternalEnergy() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * s.T * s.Tau * (s.Alpha0DTau + s.AlphaRDTau)
}

func (s *State) Cv() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return -R * s.Tau * s.Tau * (s.Alpha0DTau2 + s.AlphaRDTau2)
}

func (s *State) Cp() float64 {
	R := s.Fluid.EOS[0].GasConstant
	num := 1.0 + s.Delta*s.AlphaRDDelta - s.Delta*s.Tau*s.AlphaRDDeltaTau
	den := 1.0 + 2.0*s.Delta*s.AlphaRDDelta + s.Delta*s.Delta*s.AlphaRDDelta2
	return s.Cv() + R*num*num/den
}

func (s *State) DPdT() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return s.Rho*R*(1.0+s.Delta*s.AlphaRDDelta-s.Delta*s.Tau*s.AlphaRDDeltaTau)
}

func (s *State) DPdRho() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * s.T * (1.0 + 2.0*s.Delta*s.AlphaRDDelta + s.Delta*s.Delta*s.AlphaRDDelta2)
}

func (s *State) DHdT() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R*(1.0+s.Delta*s.AlphaRDDelta-s.Delta*s.Tau*s.AlphaRDDeltaTau) - R*s.Tau*s.Tau*(s.Alpha0DTau2+s.AlphaRDTau2)
}

func (s *State) DHdRho() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * s.T / s.ReducingRho * (s.AlphaRDDelta + s.Delta*s.AlphaRDDelta2 + s.Tau*s.AlphaRDDeltaTau)
}

func (s *State) DSdT() float64 {
	return s.Cv() / s.T
}

func (s *State) DSdRho() float64 {
	R := s.Fluid.EOS[0].GasConstant
	return R * (s.Tau*s.AlphaRDDeltaTau - s.DaDDelta) / s.ReducingRho
}
