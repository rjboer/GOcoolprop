package fluid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadFluid(path string) (*FluidData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fluid file: %w", err)
	}

	var fluid FluidData
	if err := json.Unmarshal(data, &fluid); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fluid data: %w", err)
	}

	return &fluid, nil
}

// LoadFluidByName loads a fluid by name from a directory
// Uses the fluid registry to resolve aliases
func LoadFluidByName(name, dataDir string) (*FluidData, error) {
	// Try to get filename from registry
	filename, err := GetFluidFilename(name)
	if err != nil {
		// Fallback: try direct filename
		filename = name + ".json"
	}

	path := filepath.Join(dataDir, filename)
	return LoadFluid(path)
}
