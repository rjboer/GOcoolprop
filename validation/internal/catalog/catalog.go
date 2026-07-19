package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type Entry struct {
	Name string
	Path string
}

func Discover(dir string) ([]Entry, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(files))
	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var v struct {
			Info struct {
				Name string `json:"NAME"`
			} `json:"INFO"`
		}
		if json.Unmarshal(b, &v) == nil {
			n := v.Info.Name
			if n == "" {
				n = filepath.Base(p)
				n = n[:len(n)-5]
			}
			out = append(out, Entry{Name: n, Path: p})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
