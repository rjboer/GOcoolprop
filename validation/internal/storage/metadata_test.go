package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteManifestPreservesReproducibilityFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	want := Manifest{
		RunID:             "20260719T120000Z",
		GoVersion:         "go1.23.6",
		PythonVersion:     "Python 3.13.3",
		CoolPropVersion:   "7.2.0",
		CoolPropGit:       "abc123",
		Backend:           "HEOS",
		ReferenceState:    "DEF",
		PythonExecutable:  "python313.exe",
		Seed:              20260719,
		CandidateWorkers:  8,
		ReferenceWorkers:  4,
		BatchSize:         512,
		Fluids:            []string{"Hydrogen", "Water"},
		GitRevision:       "candidate456",
		GitDirty:          true,
	}
	if err := WriteManifest(path, want); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Manifest
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.RunID != want.RunID || got.CoolPropVersion != want.CoolPropVersion || got.Seed != want.Seed || !got.GitDirty {
		t.Fatalf("manifest lost reproducibility fields: got %+v", got)
	}
}

func TestWriteResolvedConfigIsDeterministicYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.resolved.yaml")
	config := ConfigSnapshot{
		ReferenceImplementation: "python-coolprop",
		ReferenceVersion:        "7.2.0",
		Backend:                 "HEOS",
		ReferenceState:          "DEF",
		ReferenceWorkers:        4,
		BatchSize:               512,
		CandidateWorkers:        8,
		ActiveFluids:            2,
		Seed:                    20260719,
		TDTemperaturePoints:     64,
		TDDensityPoints:         64,
		PTTemperaturePoints:     64,
		PTPressurePoints:        64,
		QuasiRandomPoints:       2048,
	}
	if err := WriteResolvedConfig(path, config); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteResolvedConfig(path, config); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("resolved config changed between identical writes")
	}
	for _, required := range []string{"reference:", "version: \"7.2.0\"", "seed: 20260719", "td_temperature_points: 64"} {
		if !strings.Contains(string(first), required) {
			t.Fatalf("resolved config missing %q:\n%s", required, first)
		}
	}
}
