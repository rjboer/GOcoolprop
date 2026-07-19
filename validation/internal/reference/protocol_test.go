package reference

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartReturnsStartupError(t *testing.T) {
	python, err := exec.LookPath("py")
	if err != nil {
		python, err = exec.LookPath("python")
	}
	if err != nil {
		t.Skip("python executable not available")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "startup_error.py")
	body := "import json\nprint(json.dumps({'startup_error':'missing CoolProp'}), flush=True)\n"
	if err := os.WriteFile(script, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	worker, err := Start(ctx, python, script)

	if err == nil {
		if worker != nil {
			_ = worker.Close()
		}
		t.Fatal("expected startup error")
	}
	if !strings.Contains(err.Error(), "missing CoolProp") {
		t.Fatalf("expected startup error to include message, got %q", err)
	}
}

func TestCaseUsesReferenceJSONFieldNames(t *testing.T) {
	b, err := json.Marshal(Case{
		Output: "P",
		Input1: "T",
		Value1: 300,
		Input2: "Dmass",
		Value2: 1,
		Fluid:  "Water",
	})
	if err != nil {
		t.Fatal(err)
	}

	got := string(b)
	for _, key := range []string{`"output"`, `"input1"`, `"value1"`, `"input2"`, `"value2"`, `"fluid"`} {
		if !strings.Contains(got, key) {
			t.Fatalf("expected %s in JSON, got %s", key, got)
		}
	}
}
