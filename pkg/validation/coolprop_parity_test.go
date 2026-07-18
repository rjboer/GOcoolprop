package validation

import (
	"GOcoolprop/pkg/core"
	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/props"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type referenceDataset struct {
	StatePoints      []referenceStatePoint `json:"state_points"`
	TPPoints         []referenceTPPoint    `json:"tp_points"`
	SaturationPoints []referenceSatPoint   `json:"saturation_points"`
}

type referenceStatePoint struct {
	Name  string  `json:"name"`
	Fluid string  `json:"fluid"`
	T     float64 `json:"T"`
	Rho   float64 `json:"rho"`
	P     float64 `json:"P"`
	H     float64 `json:"H"`
	S     float64 `json:"S"`
	U     float64 `json:"U"`
	Cv    float64 `json:"Cv"`
	Cp    float64 `json:"Cp"`
}

type referenceTPPoint struct {
	Name  string  `json:"name"`
	Fluid string  `json:"fluid"`
	T     float64 `json:"T"`
	P     float64 `json:"P"`
	Rho   float64 `json:"rho"`
	H     float64 `json:"H"`
	S     float64 `json:"S"`
	U     float64 `json:"U"`
	Cv    float64 `json:"Cv"`
	Cp    float64 `json:"Cp"`
}

type referenceSatPoint struct {
	Name string  `json:"name"`
	Fluid string `json:"fluid"`
	T    float64 `json:"T,omitempty"`
	P    float64 `json:"P,omitempty"`
	Q    float64 `json:"Q"`
	PRef float64 `json:"P_ref,omitempty"`
	TRef float64 `json:"T_ref,omitempty"`
	Rho  float64 `json:"rho"`
	H    float64 `json:"H"`
	S    float64 `json:"S"`
}

func loadReferenceDataset(t *testing.T) referenceDataset {
	t.Helper()
	path := filepath.Join("testdata", "coolprop_core_reference.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read reference dataset: %v", err)
	}
	var ds referenceDataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatalf("decode reference dataset: %v", err)
	}
	return ds
}

func requireRel(t *testing.T, got, want, tol float64, label string) {
	t.Helper()
	if relErr(got, want) > tol {
		t.Fatalf("%s mismatch: got=%g want=%g rel=%g tol=%g", label, got, want, relErr(got, want), tol)
	}
}

func requireAbs(t *testing.T, got, want, tol float64, label string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s mismatch: got=%g want=%g abs=%g tol=%g", label, got, want, math.Abs(got-want), tol)
	}
}

func TestStateUpdateMatchesCoolPropReferencePoints(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.StatePoints {
		t.Run(tt.Name, func(t *testing.T) {
			f, s := buildState(t, tt.Fluid)
			s.Update(tt.T, tt.Rho)

			wantTau := f.EOS[0].States.Reducing.T / tt.T
			wantDelta := tt.Rho / f.EOS[0].States.Reducing.RhoMolar
			requireRel(t, s.Tau, wantTau, 1e-12, "tau")
			requireRel(t, s.Delta, wantDelta, 1e-12, "delta")
			requireRel(t, s.Pressure(), tt.P, 3e-4, "pressure")
			requireRel(t, s.MolarEnthalpy(), tt.H, 3e-4, "enthalpy")
			requireRel(t, s.MolarEntropy(), tt.S, 3e-4, "entropy")
			requireRel(t, s.MolarInternalEnergy(), tt.U, 3e-4, "internal energy")
			requireRel(t, s.Cv(), tt.Cv, 5e-4, "cv")
			requireRel(t, s.Cp(), tt.Cp, 5e-4, "cp")
		})
	}
}

func TestPropSITDMatchesCoolPropReferencePoints(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.StatePoints {
		t.Run(tt.Name, func(t *testing.T) {
			p, err := props.PropSI("P", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI P from T,D failed: %v", err)
			}
			requireRel(t, p, tt.P, 3e-4, "P(T,D)")

			h, err := props.PropSI("H", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI H from T,D failed: %v", err)
			}
			requireRel(t, h, tt.H, 3e-4, "H(T,D)")

			s, err := props.PropSI("S", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI S from T,D failed: %v", err)
			}
			requireRel(t, s, tt.S, 3e-4, "S(T,D)")

			u, err := props.PropSI("U", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI U from T,D failed: %v", err)
			}
			requireRel(t, u, tt.U, 3e-4, "U(T,D)")

			cv, err := props.PropSI("CV", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI Cv from T,D failed: %v", err)
			}
			requireRel(t, cv, tt.Cv, 5e-4, "Cv(T,D)")

			cp, err := props.PropSI("CP", "T", tt.T, "D", tt.Rho, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI Cp from T,D failed: %v", err)
			}
			requireRel(t, cp, tt.Cp, 5e-4, "Cp(T,D)")
		})
	}
}

func TestPropSITPMatchesCoolPropReferencePoints(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.TPPoints {
		t.Run(tt.Name, func(t *testing.T) {
			rho, err := props.PropSI("D", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI D from T,P failed: %v", err)
			}
			requireRel(t, rho, tt.Rho, 3e-4, "rho(T,P)")

			h, err := props.PropSI("H", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI H from T,P failed: %v", err)
			}
			requireRel(t, h, tt.H, 3e-4, "h(T,P)")

			s, err := props.PropSI("S", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI S from T,P failed: %v", err)
			}
			requireRel(t, s, tt.S, 3e-4, "s(T,P)")

			u, err := props.PropSI("U", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI U from T,P failed: %v", err)
			}
			requireRel(t, u, tt.U, 5e-4, "u(T,P)")

			cv, err := props.PropSI("CV", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI Cv from T,P failed: %v", err)
			}
			requireRel(t, cv, tt.Cv, 7e-4, "cv(T,P)")

			cp, err := props.PropSI("CP", "T", tt.T, "P", tt.P, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI Cp from T,P failed: %v", err)
			}
			requireRel(t, cp, tt.Cp, 7e-4, "cp(T,P)")

			rhoFromTH, err := props.PropSI("D", "T", tt.T, "H", tt.H, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI D from T,H failed: %v", err)
			}
			requireRel(t, rhoFromTH, tt.Rho, 5e-4, "rho(T,H)")

			tempFromPH, err := props.PropSI("T", "P", tt.P, "H", tt.H, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI T from P,H failed: %v", err)
			}
			requireAbs(t, tempFromPH, tt.T, 2e-3, "T(P,H)")

			rhoFromPH, err := props.PropSI("D", "P", tt.P, "H", tt.H, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI D from P,H failed: %v", err)
			}
			requireRel(t, rhoFromPH, tt.Rho, 7e-4, "rho(P,H)")

			tempFromPS, err := props.PropSI("T", "P", tt.P, "S", tt.S, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI T from P,S failed: %v", err)
			}
			requireAbs(t, tempFromPS, tt.T, 2e-3, "T(P,S)")

			rhoFromPS, err := props.PropSI("D", "P", tt.P, "S", tt.S, tt.Fluid)
			if err != nil {
				t.Fatalf("PropSI D from P,S failed: %v", err)
			}
			requireRel(t, rhoFromPS, tt.Rho, 7e-4, "rho(P,S)")
		})
	}
}

func TestPropSISaturationEndpointsMatchCoolPropReferencePoints(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.SaturationPoints {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.T != 0 {
				p, err := props.PropSI("P", "T", tt.T, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI P from T,Q failed: %v", err)
				}
				requireRel(t, p, tt.PRef, 8e-3, "P(T,Q)")

				rho, err := props.PropSI("D", "T", tt.T, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI D from T,Q failed: %v", err)
				}
				requireRel(t, rho, tt.Rho, 8e-3, "rho(T,Q)")

				h, err := props.PropSI("H", "T", tt.T, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI H from T,Q failed: %v", err)
				}
				requireRel(t, h, tt.H, 8e-3, "h(T,Q)")

				s, err := props.PropSI("S", "T", tt.T, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI S from T,Q failed: %v", err)
				}
				requireRel(t, s, tt.S, 8e-3, "s(T,Q)")

				q, err := props.PropSI("Q", "T", tt.T, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI Q from T,Q failed: %v", err)
				}
				requireAbs(t, q, tt.Q, 0, "Q(T,Q)")
			}

			if tt.P != 0 {
				temp, err := props.PropSI("T", "P", tt.P, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI T from P,Q failed: %v", err)
				}
				requireAbs(t, temp, tt.TRef, 2e-1, "T(P,Q)")

				rho, err := props.PropSI("D", "P", tt.P, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI D from P,Q failed: %v", err)
				}
				requireRel(t, rho, tt.Rho, 1e-2, "rho(P,Q)")

				h, err := props.PropSI("H", "P", tt.P, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI H from P,Q failed: %v", err)
				}
				requireRel(t, h, tt.H, 1e-2, "h(P,Q)")

				s, err := props.PropSI("S", "P", tt.P, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI S from P,Q failed: %v", err)
				}
				requireRel(t, s, tt.S, 1e-2, "s(P,Q)")

				q, err := props.PropSI("Q", "P", tt.P, "Q", tt.Q, tt.Fluid)
				if err != nil {
					t.Fatalf("PropSI Q from P,Q failed: %v", err)
				}
				requireAbs(t, q, tt.Q, 0, "Q(P,Q)")
			}
		})
	}
}

func TestPropSIRejectsTPAtExactSaturationBoundary(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.SaturationPoints {
		if tt.T == 0 || tt.PRef == 0 {
			continue
		}
		t.Run(tt.Name, func(t *testing.T) {
			if _, err := props.PropSI("D", "T", tt.T, "P", tt.PRef, tt.Fluid); err == nil {
				t.Fatalf("expected T,P saturation-boundary request to fail for %s", tt.Name)
			}
		})
	}
}

func TestStateUpdateInvalidInputsProduceUnusableState(t *testing.T) {
	_, s := buildState(t, "Water")
	s.Update(0, 10)
	if !math.IsNaN(s.Tau) || !math.IsNaN(s.Delta) || !math.IsNaN(s.Pressure()) {
		t.Fatalf("expected invalid temperature update to poison state")
	}

	_, s = buildState(t, "Water")
	s.Update(300, -1)
	if !math.IsNaN(s.Tau) || !math.IsNaN(s.Delta) || !math.IsNaN(s.Pressure()) {
		t.Fatalf("expected invalid density update to poison state")
	}
}

func TestCoreStateReferencePointsStillConstruct(t *testing.T) {
	ds := loadReferenceDataset(t)
	for _, tt := range ds.StatePoints {
		t.Run(tt.Fluid, func(t *testing.T) {
			f, err := fluid.LoadFluidByName(tt.Fluid, "../../data")
			if err != nil {
				t.Fatalf("load %s: %v", tt.Fluid, err)
			}
			if _, err := core.NewState(f); err != nil {
				t.Fatalf("new state %s: %v", tt.Fluid, err)
			}
		})
	}
}
