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

func LoadFluidByName(name string, dataDir string) (*FluidData, error) {
	// Try name.json
	path := filepath.Join(dataDir, name+".json")
	return LoadFluid(path)
}
