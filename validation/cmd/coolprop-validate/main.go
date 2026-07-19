package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"GOcoolprop/validation/internal/candidate"
	"GOcoolprop/validation/internal/catalog"
	"GOcoolprop/validation/internal/compare"
	"GOcoolprop/validation/internal/generator"
	"GOcoolprop/validation/internal/reference"
	"GOcoolprop/validation/internal/report"
	"GOcoolprop/validation/internal/stats"
	"GOcoolprop/validation/internal/storage"
)

type options struct {
	All       bool
	Preflight bool
	Fluids    []string
	Root      string
	Python    string
	Generator generator.Config
}

func parseOptions(args []string) (options, error) {
	opts := options{
		Root:   "results",
		Python: "python",
		Generator: generator.Config{
			TDTemperaturePoints: 64,
			TDDensityPoints:     64,
			PTTemperaturePoints: 64,
			PTPressurePoints:    64,
			QuasiRandomPoints:   2048,
			Seed:                20260719,
		},
	}
	var fluidsFlag string
	seed := int(opts.Generator.Seed)
	fs := flag.NewFlagSet("coolprop-validate", flag.ContinueOnError)
	fs.BoolVar(&opts.All, "all", false, "validate every JSON fluid")
	fs.BoolVar(&opts.Preflight, "preflight", false, "check Go, Python, reference script, and CoolProp availability")
	fs.StringVar(&fluidsFlag, "fluids", "Water,Nitrogen,Hydrogen", "comma-separated fluids")
	fs.StringVar(&opts.Root, "results", opts.Root, "result root")
	fs.StringVar(&opts.Python, "python", opts.Python, "Python executable")
	fs.IntVar(&opts.Generator.TDTemperaturePoints, "td-t", opts.Generator.TDTemperaturePoints, "T-D screening temperature count")
	fs.IntVar(&opts.Generator.TDDensityPoints, "td-rho", opts.Generator.TDDensityPoints, "T-D screening density count")
	fs.IntVar(&opts.Generator.PTTemperaturePoints, "pt-t", opts.Generator.PTTemperaturePoints, "P-T screening temperature count")
	fs.IntVar(&opts.Generator.PTPressurePoints, "pt-p", opts.Generator.PTPressurePoints, "P-T screening pressure count")
	fs.IntVar(&opts.Generator.QuasiRandomPoints, "quasi", opts.Generator.QuasiRandomPoints, "quasi-random screening case count")
	fs.IntVar(&seed, "seed", seed, "deterministic generation seed")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	opts.Generator.Seed = int64(seed)
	opts.Fluids = strings.Split(fluidsFlag, ",")
	return opts, nil
}

func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		fatal(err)
	}
	ctx := context.Background()
	if opts.Preflight {
		if err := preflight(ctx, opts); err != nil {
			fatal(err)
		}
		return
	}
	if err := run(ctx, opts); err != nil {
		fatal(err)
	}
}

func run(ctx context.Context, opts options) error {
	fluids := opts.Fluids
	if opts.All {
		entries, err := catalog.Discover("data")
		if err != nil {
			return err
		}
		fluids = fluids[:0]
		for _, e := range entries {
			fluids = append(fluids, e.Name)
		}
	}
	var ref *reference.Worker
	script := filepath.Join("validation", "reference", "coolprop_reference.py")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("reference script unavailable: %w", err)
	}
	ref, err := reference.Start(ctx, opts.Python, script)
	if err != nil {
		return fmt.Errorf("reference unavailable: %w", err)
	}
	defer func() { _ = ref.Close() }()
	run, err := storage.New(opts.Root)
	if err != nil {
		return err
	}
	if err := writeRunMetadata(run, opts, fluids, ref.Startup); err != nil {
		return err
	}
	engine := candidate.Engine{DataDir: "data"}
	suites := []string{"td_grid", "pt_grid", "quasi_random", "saturation", "phase_boundary", "critical_boundary", "validity_boundary", "flash", "round_trip", "invalid_input"}
	statPlan, err := stats.BuildPlan(fluids, suites, engine.Capabilities().InputPairs, engine.Capabilities().Outputs, 5000, 0.99, 0.001)
	if err != nil {
		return err
	}
	if err := storage.WriteJSON(filepath.Join(opts.Root, run.ID, "statistical-plan.json"), statPlan); err != nil {
		return err
	}
	rows := make([]report.Fluid, 0, len(fluids))
	var mu sync.Mutex
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for _, fluidName := range fluids {
		fluidName = strings.TrimSpace(fluidName)
		if fluidName == "" {
			continue
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			row := runFluid(ctx, run, name, engine, ref, opts.Generator)
			mu.Lock()
			rows = append(rows, row)
			mu.Unlock()
		}(fluidName)
	}
	wg.Wait()
	_ = report.IndexWithPlan(filepath.Join(opts.Root, run.ID, "index.md"), run.ID, rows, statPlan)
	fmt.Printf("run=%s fluids=%d workers=%d\n", run.ID, len(rows), runtime.NumCPU())
	return nil
}

func writeRunMetadata(run storage.Run, opts options, fluids []string, startup map[string]any) error {
	gitRevision := ""
	gitDirty := false
	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		gitRevision = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		gitDirty = strings.TrimSpace(string(out)) != ""
	}
	pythonVersion := startupString(startup, "python_version")
	if pythonVersion == "" {
		pythonVersion = startupString(startup, "python")
	}
	manifest := storage.Manifest{
		RunID:            run.ID,
		StartedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		GoVersion:        runtime.Version(),
		PythonVersion:    pythonVersion,
		CoolPropVersion:  startupString(startup, "coolprop_version"),
		CoolPropGit:      startupString(startup, "gitrevision"),
		Backend:          startupString(startup, "backend"),
		ReferenceState:   startupString(startup, "reference_state"),
		PythonExecutable: opts.Python,
		Seed:             opts.Generator.Seed,
		CandidateWorkers: runtime.NumCPU(),
		ReferenceWorkers: 1,
		BatchSize:        512,
		Fluids:           append([]string(nil), fluids...),
		GitRevision:      gitRevision,
		GitDirty:         gitDirty,
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,
	}
	config := storage.ConfigSnapshot{
		ReferenceImplementation: "python-coolprop",
		ReferenceVersion:        manifest.CoolPropVersion,
		Backend:                 manifest.Backend,
		ReferenceState:          manifest.ReferenceState,
		ReferenceWorkers:        manifest.ReferenceWorkers,
		BatchSize:               manifest.BatchSize,
		CandidateWorkers:        manifest.CandidateWorkers,
		ActiveFluids:            len(fluids),
		Seed:                    opts.Generator.Seed,
		TDTemperaturePoints:     opts.Generator.TDTemperaturePoints,
		TDDensityPoints:         opts.Generator.TDDensityPoints,
		PTTemperaturePoints:     opts.Generator.PTTemperaturePoints,
		PTPressurePoints:        opts.Generator.PTPressurePoints,
		QuasiRandomPoints:       opts.Generator.QuasiRandomPoints,
		StatisticsEnabled:       true,
		Confidence:              0.99,
		DetectablePrevalence:    0.001,
		MinimumAuditSamples:     5000,
		FamilyWiseAdjustment:    "bonferroni",
		SafetyMultiplier:        1.5,
	}
	runDir := filepath.Join(run.Root, run.ID)
	if err := storage.WriteManifest(filepath.Join(runDir, "manifest.json"), manifest); err != nil {
		return err
	}
	return storage.WriteResolvedConfig(filepath.Join(runDir, "config.resolved.yaml"), config)
}

func startupString(startup map[string]any, key string) string {
	value, _ := startup[key].(string)
	return value
}

func preflight(ctx context.Context, opts options) error {
	fmt.Printf("go=%s\n", runtime.Version())
	out, err := exec.CommandContext(ctx, opts.Python, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("python unavailable: %w: %s", err, strings.TrimSpace(string(out)))
	}
	fmt.Printf("python=%s\n", strings.TrimSpace(string(out)))
	script := filepath.Join("validation", "reference", "coolprop_reference.py")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("reference script unavailable: %w", err)
	}
	ref, err := reference.Start(ctx, opts.Python, script)
	if err != nil {
		return err
	}
	defer func() { _ = ref.Close() }()
	fmt.Printf("coolprop=%v backend=%v reference_state=%v\n", ref.Startup["coolprop_version"], ref.Startup["backend"], ref.Startup["reference_state"])
	return nil
}

func runFluid(ctx context.Context, run storage.Run, name string, engine candidate.Engine, ref *reference.Worker, gen generator.Config) report.Fluid {
	dir := run.FluidDir(name)
	meta := engine.Metadata(name)
	env := screeningEnvelope(meta)
	cases := generator.Screening(name, env, gen)
	passed, failed := 0, 0
	failures := make([]caseFailure, 0)
	for _, c := range cases {
		cr := engine.Evaluate(ctx, candidate.Request{ID: c.ID, Fluid: c.Fluid, Output: c.Output, Input1: c.Input1, Input2: c.Input2, Value1: c.Value1, Value2: c.Value2})
		ok := cr.Error == ""
		if ref != nil {
			rr, err := ref.Call(reference.Request{RequestID: c.ID, Fluid: c.Fluid, Cases: []reference.Case{{Output: c.Output, Input1: c.Input1, Input2: c.Input2, Fluid: c.Fluid, Value1: c.Value1, Value2: c.Value2}}})
			out := compareOutcome(cr, rr, err)
			ok = out.OK
			if !out.OK {
				out.Failure.ID = c.ID
				out.Failure.Output = c.Output
				out.Failure.Input1 = c.Input1
				out.Failure.Input2 = c.Input2
				out.Failure.Value1 = c.Value1
				out.Failure.Value2 = c.Value2
				failures = append(failures, out.Failure)
			}
		} else if !ok {
			failures = append(failures, caseFailure{ID: c.ID, Reason: "candidate_error", Output: c.Output, Input1: c.Input1, Input2: c.Input2, Value1: c.Value1, Value2: c.Value2, CandidateError: cr.Error})
		}
		if ok {
			passed++
		} else {
			failed++
		}
	}
	status := "passed"
	if failed > 0 {
		status = "failed"
	}
	_ = storage.WriteSummary(dir, name, status, len(cases), passed, failed)
	_ = storage.WriteJSON(filepath.Join(dir, "summary.json"), map[string]any{"fluid": name, "status": status, "total": len(cases), "passed": passed, "failed": failed, "completed_at": time.Now().UTC()})
	_ = storage.WriteJSON(filepath.Join(dir, "failures.json"), failures)
	_ = run.Finish(name, failed == 0)
	return report.Fluid{Name: name, Status: status, Total: len(cases), Passed: passed, Failed: failed, Report: filepath.Join(status, name, "summary.md")}
}

func screeningEnvelope(meta candidate.FluidMetadata) generator.Envelope {
	env := generator.Envelope{TMin: meta.Tmin, TMax: meta.Tmax, PMin: meta.Pmin, PMax: meta.Pmax, RhoMin: 1, RhoMax: meta.RhoCrit * 20}
	if env.TMax <= env.TMin {
		env.TMin = 100
		env.TMax = 500
	} else {
		env.TMin += (env.TMax - env.TMin) * 1e-6
	}
	if env.PMax <= env.PMin {
		env.PMax = 1e8
	}
	if meta.Pcrit > env.PMin && meta.Pcrit < env.PMax {
		env.PMax = meta.Pcrit
	}
	if env.PMin <= 0 {
		env.PMin = 1
	}
	if meta.RhoTripleVapor > 0 {
		env.RhoMin = meta.RhoTripleVapor / 10
	}
	if env.RhoMax <= env.RhoMin {
		env.RhoMax = 1000
	}
	return env
}

type caseOutcome struct {
	OK      bool
	Failure caseFailure
}

type caseFailure struct {
	ID             string  `json:"id,omitempty"`
	Reason         string  `json:"reason"`
	Output         string  `json:"output,omitempty"`
	Input1         string  `json:"input1,omitempty"`
	Input2         string  `json:"input2,omitempty"`
	Value1         float64 `json:"value1,omitempty"`
	Value2         float64 `json:"value2,omitempty"`
	Candidate      float64 `json:"candidate,omitempty"`
	Reference      float64 `json:"reference,omitempty"`
	AbsError       float64 `json:"abs_error,omitempty"`
	RelError       float64 `json:"rel_error,omitempty"`
	NormError      float64 `json:"normalized_error,omitempty"`
	Outcome        string  `json:"outcome,omitempty"`
	Tolerance      float64 `json:"relative_tolerance,omitempty"`
	CandidateError string  `json:"candidate_error,omitempty"`
	ReferenceError string  `json:"reference_error,omitempty"`
}

func compareOutcome(cr candidate.Result, rr reference.Response, refErr error) caseOutcome {
	if refErr != nil {
		return caseOutcome{Failure: caseFailure{Reason: "reference_error", Outcome: compare.OutcomeValidatorError, Candidate: cr.Value, ReferenceError: refErr.Error()}}
	}
	if len(rr.Results) == 0 {
		return caseOutcome{Failure: caseFailure{Reason: "reference_error", Outcome: compare.OutcomeValidatorError, Candidate: cr.Value, ReferenceError: "empty response"}}
	}
	item := rr.Results[0]
	result := compare.Compare(cr.Value, cr.Error, "", item.Value, item.Error, item.Phase, compare.Tolerance{Absolute: 1e-9, Relative: 1e-8})
	if result.Outcome == compare.OutcomePassed || result.Outcome == compare.OutcomeConsistentError {
		return caseOutcome{OK: true}
	}
	reason := "tolerance"
	if result.Outcome == compare.OutcomeErrorMismatch {
		reason = "reference_error"
	}
	if cr.Error != "" {
		reason = "candidate_error"
	}
	return caseOutcome{
		Failure: caseFailure{
			Reason:         reason,
			Outcome:        result.Outcome,
			Candidate:      cr.Value,
			Reference:      item.Value,
			AbsError:       result.Metric.Absolute,
			RelError:       result.Metric.Relative,
			NormError:      result.Metric.Normalized,
			Tolerance:      result.Tolerance.Relative,
			CandidateError: cr.Error,
			ReferenceError: item.Error,
		},
	}
}

func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
