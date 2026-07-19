package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Run struct{ Root, ID string }

func New(root string) (Run, error) {
	id := time.Now().UTC().Format("20060102T150405Z")
	r := Run{Root: root, ID: id}
	for _, d := range []string{"_running", "passed", "failed"} {
		if err := os.MkdirAll(filepath.Join(root, id, d), 0755); err != nil {
			return r, err
		}
	}
	return r, nil
}
func (r Run) FluidDir(fluid string) string {
	p := filepath.Join(r.Root, r.ID, "_running", fluid)
	_ = os.MkdirAll(p, 0755)
	return p
}
func (r Run) Finish(fluid string, passed bool) error {
	src := r.FluidDir(fluid)
	dst := filepath.Join(r.Root, r.ID, "failed", fluid)
	if passed {
		dst = filepath.Join(r.Root, r.ID, "passed", fluid)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}
func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func WriteSummary(dir, fluid, status string, total, passed, failed int) error {
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(fmt.Sprintf("# %s\n\nStatus: **%s**\n\n| Metric | Count |\n|---|---:|\n| Mandatory cases | %d |\n| Passed | %d |\n| Failed | %d |\n", fluid, status, total, passed, failed)), 0644)
}
