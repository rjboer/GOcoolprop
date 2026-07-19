package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"GOcoolprop/validation/internal/candidate"
	"GOcoolprop/validation/internal/generator"
	"GOcoolprop/validation/internal/reference"
	"GOcoolprop/validation/internal/storage"
)

func TestOptionsParseSmallScreeningFlags(t *testing.T) {
	opts, err := parseOptions([]string{
		"-fluids", "Water",
		"-td-t", "2",
		"-td-rho", "3",
		"-pt-t", "4",
		"-pt-p", "5",
		"-quasi", "6",
		"-preflight",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !opts.Preflight {
		t.Fatal("expected preflight to be enabled")
	}
	if opts.Generator.TDTemperaturePoints != 2 ||
		opts.Generator.TDDensityPoints != 3 ||
		opts.Generator.PTTemperaturePoints != 4 ||
		opts.Generator.PTPressurePoints != 5 ||
		opts.Generator.QuasiRandomPoints != 6 {
		t.Fatalf("unexpected generator config: %+v", opts.Generator)
	}
}

func TestCompareOutcomeRecordsReferenceCallFailure(t *testing.T) {
	out := compareOutcome(
		candidate.Result{Value: 1},
		reference.Response{},
		errors.New("worker closed"),
	)

	if out.OK {
		t.Fatal("expected failed outcome")
	}
	if out.Failure.Reason != "reference_error" {
		t.Fatalf("unexpected reason: %+v", out.Failure)
	}
}

func TestCompareOutcomeRecordsNumericalMismatch(t *testing.T) {
	out := compareOutcome(
		candidate.Result{Value: 10},
		reference.Response{Results: []reference.Item{{Value: 12}}},
		nil,
	)

	if out.OK {
		t.Fatal("expected failed outcome")
	}
	if out.Failure.Reason != "tolerance" {
		t.Fatalf("unexpected reason: %+v", out.Failure)
	}
	if out.Failure.Candidate != 10 || out.Failure.Reference != 12 {
		t.Fatalf("unexpected values: %+v", out.Failure)
	}
}

func TestScreeningEnvelopeUsesInteriorTemperatureAndMetadataPressure(t *testing.T) {
	env := screeningEnvelope(candidate.FluidMetadata{
		Tmin:           100,
		Tmax:           200,
		Pmin:           1234,
		Pmax:           1e9,
		Pcrit:          1e6,
		RhoCrit:        10,
		RhoTripleVapor: 0.5,
	})

	if env.TMin <= 100 {
		t.Fatalf("expected interior Tmin, got %+v", env)
	}
	if env.PMin != 1234 {
		t.Fatalf("expected metadata Pmin, got %+v", env)
	}
	if env.PMax != 1e6 {
		t.Fatalf("expected screening Pmax to use critical pressure, got %+v", env)
	}
	if env.RhoMin != 0.05 {
		t.Fatalf("expected screening RhoMin below triple vapor density, got %+v", env)
	}
}

func TestWriteRunMetadataCreatesReproducibleArtifacts(t *testing.T) {
	run, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	opts := options{Python: "python313.exe", Generator: generator.Config{Seed: 20260719}}
	startup := map[string]any{
		"python_version":  "Python 3.13.3",
		"coolprop_version": "7.2.0",
		"gitrevision":     "reference123",
		"backend":         "HEOS",
		"reference_state": "DEF",
	}
	if err := writeRunMetadata(run, opts, []string{"Water", "Hydrogen"}, startup); err != nil {
		t.Fatal(err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(run.Root, run.ID, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest["coolprop_version"] != "7.2.0" || manifest["seed"] != float64(20260719) {
		t.Fatalf("unexpected manifest: %s", manifestBytes)
	}
	configBytes, err := os.ReadFile(filepath.Join(run.Root, run.ID, "config.resolved.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(configBytes), "seed: 20260719") {
		t.Fatalf("resolved config missing seed: %s", configBytes)
	}
}
