package fluid

type FluidData struct {
	Ancillaries Ancillaries `json:"ANCILLARIES"`
	EOS         []EOS       `json:"EOS"`
	Info        Info        `json:"INFO"`
	States      States      `json:"STATES"`
	Transport   Transport   `json:"TRANSPORT"`
}

type Transport struct {
	Viscosity      ViscosityData      `json:"viscosity"`
	Conductivity   ConductivityData   `json:"conductivity"`
	SurfaceTension SurfaceTensionData `json:"surface_tension"`
}

type ViscosityData struct {
	BibTeX       string           `json:"BibTeX"`
	Hardcoded    string           `json:"hardcoded"`
	Dilute       *ViscosityDilute `json:"dilute"`
	HigherOrder  *ViscosityHigher `json:"higher_order"`
	SigmaEta     float64          `json:"sigma_eta"`
	EpsilonOverK float64          `json:"epsilon_over_k"`
}

type ViscosityDilute struct {
	Type      string    `json:"type"`
	C         float64   `json:"C"`
	A         []float64 `json:"a"`
	MolarMass float64   `json:"molar_mass"`
	T         []float64 `json:"t"`
}

type ViscosityHigher struct {
	Type      string    `json:"type"`
	TReduce   float64   `json:"T_reduce"`
	RhoReduce float64   `json:"rhomolar_reduce"`
	A         []float64 `json:"a"`
	D1        []float64 `json:"d1"`
	D2        []float64 `json:"d2"`
	T1        []float64 `json:"t1"`
	T2        []float64 `json:"t2"`
	F         []float64 `json:"f"`
	G         []float64 `json:"g"`
	Gamma     []float64 `json:"gamma"`
	H         []float64 `json:"h"`
	L         []float64 `json:"l"`
	P         []float64 `json:"p"`
	Q         []float64 `json:"q"`
}

type ConductivityData struct {
	BibTeX    string              `json:"BibTeX"`
	Hardcoded string              `json:"hardcoded"`
	Dilute    *ConductivityDilute `json:"dilute"`
	Residual  *ConductivityResid  `json:"residual"`
	Critical  *ConductivityCrit   `json:"critical"`
}

type ConductivityDilute struct {
	Type string    `json:"type"`
	A    []float64 `json:"A"`
	B    []float64 `json:"B"`
	T    []float64 `json:"t"` // Sometimes used
}

type ConductivityResid struct {
	Type  string    `json:"type"`
	A     []float64 `json:"A"`
	D     []float64 `json:"d"`
	Gamma []float64 `json:"gamma"`
	L     []float64 `json:"l"`
	T     []float64 `json:"t"`
}

type ConductivityCrit struct {
	// Simplified for now
	Hardcoded string `json:"hardcoded"`
}

type SurfaceTensionData struct {
	BibTeX string    `json:"BibTeX"`
	Type   string    `json:"type"`
	A      []float64 `json:"a"`
	N      []float64 `json:"n"`
	Tc     float64   `json:"Tc"`
}

type Ancillaries struct {
	// We might not need all ancillaries for the core PropSI, but good to have.
	PS             AncillaryCurve     `json:"pS"`
	RhoL           AncillaryCurve     `json:"rhoL"`
	RhoV           AncillaryCurve     `json:"rhoV"`
	SurfaceTension SurfaceTensionData `json:"surface_tension"`
}

type AncillaryCurve struct {
	Type          string    `json:"type"`
	TMax          float64   `json:"Tmax"`
	TMin          float64   `json:"Tmin"`
	ReducingValue float64   `json:"reducing_value"`
	N             []float64 `json:"n"`
	T             []float64 `json:"t"`
	TR            float64   `json:"T_r"` // Reducing temperature
	UsingTauR     bool      `json:"using_tau_r"`
}

type EOS struct {
	BibTeXEOS      string          `json:"BibTeX_EOS"`
	States         EOSStates       `json:"STATES"`
	TMax           float64         `json:"T_max"`
	TTriple        float64         `json:"Ttriple"`
	Acentric       float64         `json:"acentric"`
	Alpha0         []Alpha0Term    `json:"alpha0"`
	AlphaR         []AlphaRTerm    `json:"alphar"`
	GasConstant    float64         `json:"gas_constant"`
	MolarMass      float64         `json:"molar_mass"`
	PMax           float64         `json:"p_max"`
	CriticalRegion *CriticalRegion `json:"critical_region_splines,omitempty"`
}

type EOSStates struct {
	Reducing StatePoint `json:"reducing"`
	Critical StatePoint `json:"critical"` // Sometimes in EOS, sometimes in top-level STATES
}

type StatePoint struct {
	T        float64 `json:"T"`
	P        float64 `json:"p"`
	RhoMolar float64 `json:"rhomolar"`
	HMolar   float64 `json:"hmolar"`
	SMolar   float64 `json:"smolar"`
}

type Alpha0Term struct {
	Type string    `json:"type"`
	A1   float64   `json:"a1,omitempty"`
	A2   float64   `json:"a2,omitempty"`
	A    float64   `json:"a,omitempty"` // For LogTau
	N    []float64 `json:"n,omitempty"`
	T    []float64 `json:"t,omitempty"`
}

type AlphaRTerm struct {
	Type    string    `json:"type"`
	N       []float64 `json:"n,omitempty"`
	T       []float64 `json:"t,omitempty"`
	D       []float64 `json:"d,omitempty"`
	L       []float64 `json:"l,omitempty"` // For Power
	P       []float64 `json:"p,omitempty"` // Sometimes used?
	Gamma   []float64 `json:"gamma,omitempty"`
	Epsilon []float64 `json:"epsilon,omitempty"`
	Beta    []float64 `json:"beta,omitempty"`
	Eta     []float64 `json:"eta,omitempty"`
}

type CriticalRegion struct {
	// Simplified for now
}

type Info struct {
	Name    string `json:"NAME"`
	Formula string `json:"FORMULA"`
}

type States struct {
	Critical     StatePoint `json:"critical"`
	TripleLiquid StatePoint `json:"triple_liquid"`
	TripleVapor  StatePoint `json:"triple_vapor"`
}
