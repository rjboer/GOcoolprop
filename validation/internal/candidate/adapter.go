package candidate

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"GOcoolprop/pkg/fluid"
	"GOcoolprop/pkg/props"
)

type Request struct {
	ID, Fluid, Output, Input1, Input2 string
	Value1, Value2                    float64
}
type Result struct {
	Value         float64
	Error         string
	ErrorCategory string
	Phase         string
}
type CapabilityManifest struct {
	Fluids     []string
	Outputs    []string
	InputPairs []string
	Units      map[string]string
	Aliases    map[string][]string
	Saturation bool
	Phase      bool
	Transport  []string
}
type FluidMetadata struct {
	Name                                                                     string
	MolarMass, Tmin, Tmax, Pmin, Pmax, Tcrit, Pcrit, RhoCrit, RhoTripleVapor float64
}
type PropertyEngine interface {
	Evaluate(context.Context, Request) Result
	Capabilities() CapabilityManifest
	Metadata(string) FluidMetadata
}
type Engine struct{ DataDir string }

func (e Engine) Evaluate(ctx context.Context, req Request) (result Result) {
	select {
	case <-ctx.Done():
		return Result{Error: ctx.Err().Error(), ErrorCategory: "timeout"}
	default:
	}
	defer func() {
		if r := recover(); r != nil {
			result = Result{Error: fmt.Sprintf("panic: %v", r), ErrorCategory: "panic"}
		}
	}()
	o, massOutput := canonicalOutput(req.Output)
	i1, value1 := e.canonicalInput(req.Fluid, req.Input1, req.Value1)
	i2, value2 := e.canonicalInput(req.Fluid, req.Input2, req.Value2)
	v, err := props.PropSI(o, i1, value1, i2, value2, req.Fluid)
	if err != nil {
		return Result{Error: err.Error(), ErrorCategory: normalize(err.Error())}
	}
	if massOutput {
		if strings.EqualFold(o, "Dmolar") {
			v = e.toMass(req.Fluid, v)
		} else {
			v = e.toSpecific(req.Fluid, v)
		}
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return Result{Error: "non-finite result", ErrorCategory: "internal_error"}
	}
	result.Value = v
	result.Phase = e.phase(req)
	return result
}

func (e Engine) phase(req Request) string {
	var temperature, pressure float64
	switch {
	case strings.EqualFold(req.Input1, "T") && strings.EqualFold(req.Input2, "P"):
		temperature, pressure = req.Value1, req.Value2
	case strings.EqualFold(req.Input1, "P") && strings.EqualFold(req.Input2, "T"):
		pressure, temperature = req.Value1, req.Value2
	default:
		return ""
	}
	meta := e.Metadata(req.Fluid)
	if temperature >= meta.Tcrit {
		if pressure >= meta.Pcrit {
			return "supercritical"
		}
		return "supercritical_gas"
	}
	if pressure >= meta.Pcrit {
		return "supercritical_liquid"
	}
	psat, err := props.PropSI("P_SAT", "T", temperature, "Q", 0, req.Fluid)
	if err != nil {
		return ""
	}
	if math.Abs(pressure-psat)/math.Max(1, math.Abs(psat)) <= 1e-5 {
		return "two_phase"
	}
	if pressure > psat {
		return "liquid"
	}
	return "gas"
}
func (e Engine) toMolar(name string, mass float64) float64 {
	m := e.Metadata(name).MolarMass
	if m == 0 {
		return mass
	}
	return mass / m
}
func (e Engine) toMass(name string, molar float64) float64 { return molar * e.Metadata(name).MolarMass }
func (e Engine) toSpecific(name string, molar float64) float64 {
	m := e.Metadata(name).MolarMass
	if m == 0 {
		return molar
	}
	return molar / m
}
func (e Engine) Capabilities() CapabilityManifest {
	return CapabilityManifest{
		Outputs:    []string{"T", "P", "Dmass", "Dmolar", "Hmass", "Hmolar", "Smass", "Smolar", "Umass", "Umolar", "Cpmass", "Cpmolar", "Cvmass", "Cvmolar", "Q", "P_SAT", "T_SAT", "V", "L", "I"},
		InputPairs: []string{"T,Dmass", "T,Dmolar", "T,P", "T,Hmass", "T,Hmolar", "P,Hmass", "P,Hmolar", "P,Smass", "P,Smolar", "P,Q", "T,Q"},
		Units: map[string]string{
			"T": "K", "P": "Pa", "Dmass": "kg/m3", "Dmolar": "mol/m3",
			"Hmass": "J/kg", "Hmolar": "J/mol", "Smass": "J/(kg*K)", "Smolar": "J/(mol*K)",
			"Umass": "J/kg", "Umolar": "J/mol", "Cpmass": "J/(kg*K)", "Cpmolar": "J/(mol*K)",
			"Cvmass": "J/(kg*K)", "Cvmolar": "J/(mol*K)", "V": "Pa*s", "L": "W/(m*K)", "I": "N/m",
		},
		Aliases: map[string][]string{
			"Dmass": {"D", "DMASS", "Dmass"}, "Dmolar": {"DMOLAR", "Dmolar"},
			"Hmass": {"H", "HMASS", "Hmass"}, "Hmolar": {"HMOLAR", "Hmolar"},
			"Smass": {"S", "SMASS", "Smass"}, "Smolar": {"SMOLAR", "Smolar"},
			"Umass": {"U", "UMASS", "Umass"}, "Umolar": {"UMOLAR", "Umolar"},
		},
		Saturation: true,
		Phase:      true,
		Transport:  []string{"V", "L", "I"},
	}
}

func canonicalOutput(output string) (string, bool) {
	switch strings.ToUpper(output) {
	case "D", "DMASS":
		return "Dmolar", true
	case "DMOLAR":
		return "Dmolar", false
	case "H", "HMASS":
		return "H", true
	case "HMOLAR":
		return "Hmolar", false
	case "S", "SMASS":
		return "S", true
	case "SMOLAR":
		return "Smolar", false
	case "U", "UMASS":
		return "U", true
	case "UMOLAR":
		return "Umolar", false
	case "CP", "CPMASS":
		return "Cp", true
	case "CPMOLAR":
		return "Cpmolar", false
	case "CV", "CVMASS":
		return "Cv", true
	case "CVMOLAR":
		return "Cvmolar", false
	default:
		return output, false
	}
}

func (e Engine) canonicalInput(fluidName, input string, value float64) (string, float64) {
	switch strings.ToUpper(input) {
	case "D", "DMASS":
		return "Dmolar", e.toMolar(fluidName, value)
	case "DMOLAR":
		return "Dmolar", value
	case "H", "HMASS":
		return "H", e.toMolar(fluidName, value)
	case "HMOLAR":
		return "Hmolar", value
	case "S", "SMASS":
		return "S", e.toMolar(fluidName, value)
	case "SMOLAR":
		return "Smolar", value
	case "U", "UMASS":
		return "U", e.toMolar(fluidName, value)
	case "UMOLAR":
		return "Umolar", value
	default:
		return input, value
	}
}
func (e Engine) Metadata(name string) FluidMetadata {
	f, err := e.load(name)
	if err != nil || len(f.EOS) == 0 {
		return FluidMetadata{Name: name}
	}
	x := f.EOS[0]
	return FluidMetadata{Name: name, MolarMass: x.MolarMass, Tmin: x.TTriple, Tmax: x.TMax, Pmin: f.States.TripleVapor.P, Pmax: x.PMax, Tcrit: f.States.Critical.T, Pcrit: f.States.Critical.P, RhoCrit: f.States.Critical.RhoMolar, RhoTripleVapor: f.States.TripleVapor.RhoMolar}
}

func (e Engine) load(name string) (*fluid.FluidData, error) {
	paths := []string{e.DataDir, "data", "../../data", "../../../data", filepath.Join("..", "..", "..", "..", "data")}
	var last error
	for _, p := range paths {
		if p == "" {
			continue
		}
		f, err := fluid.LoadFluidByName(name, p)
		if err == nil {
			return f, nil
		}
		last = err
	}
	return nil, last
}
func normalize(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "not supported"):
		return "not_implemented"
	case strings.Contains(s, "out of range"):
		return "out_of_range"
	case strings.Contains(s, "converg"):
		return "no_convergence"
	case strings.Contains(s, "saturation"):
		return "phase_ambiguity"
	default:
		return "internal_error"
	}
}
