package fluid

import (
	"fmt"
	"strings"
)

// FluidRegistry maps fluid names and aliases to their JSON filenames
var FluidRegistry = map[string]string{
	// Common gases
	"air":           "Air.json",
	"nitrogen":      "Nitrogen.json",
	"n2":            "Nitrogen.json",
	"oxygen":        "Oxygen.json",
	"o2":            "Oxygen.json",
	"argon":         "Argon.json",
	"ar":            "Argon.json",
	"helium":        "Helium.json",
	"he":            "Helium.json",
	"neon":          "Neon.json",
	"ne":            "Neon.json",
	"krypton":       "Krypton.json",
	"kr":            "Krypton.json",
	"xenon":         "Xenon.json",
	"xe":            "Xenon.json",
	"hydrogen":      "Hydrogen.json",
	"h2":            "Hydrogen.json",
	"parahydrogen":  "ParaHydrogen.json",
	"orthohydrogen": "OrthoHydrogen.json",
	"deuterium":     "Deuterium.json",
	"d2":            "Deuterium.json",

	// Water
	"water":      "Water.json",
	"h2o":        "Water.json",
	"heavywater": "HeavyWater.json",
	"d2o":        "HeavyWater.json",

	// Carbon compounds
	"carbondioxide":  "CarbonDioxide.json",
	"co2":            "CarbonDioxide.json",
	"carbonmonoxide": "CarbonMonoxide.json",
	"co":             "CarbonMonoxide.json",

	// Hydrocarbons
	"methane":    "Methane.json",
	"ch4":        "Methane.json",
	"ethane":     "Ethane.json",
	"c2h6":       "Ethane.json",
	"propane":    "n-Propane.json",
	"n-propane":  "n-Propane.json",
	"c3h8":       "n-Propane.json",
	"butane":     "n-Butane.json",
	"n-butane":   "n-Butane.json",
	"isobutane":  "IsoButane.json",
	"pentane":    "n-Pentane.json",
	"n-pentane":  "n-Pentane.json",
	"isopentane": "Isopentane.json",
	"hexane":     "n-Hexane.json",
	"n-hexane":   "n-Hexane.json",
	"isohexane":  "Isohexane.json",
	"heptane":    "n-Heptane.json",
	"n-heptane":  "n-Heptane.json",
	"octane":     "n-Octane.json",
	"n-octane":   "n-Octane.json",
	"nonane":     "n-Nonane.json",
	"n-nonane":   "n-Nonane.json",
	"decane":     "n-Decane.json",
	"n-decane":   "n-Decane.json",

	// Refrigerants (R-series)
	"r11":        "R11.json",
	"r12":        "R12.json",
	"r13":        "R13.json",
	"r14":        "R14.json",
	"r21":        "R21.json",
	"r22":        "R22.json",
	"r23":        "R23.json",
	"r32":        "R32.json",
	"r40":        "R40.json",
	"r41":        "R41.json",
	"r113":       "R113.json",
	"r114":       "R114.json",
	"r115":       "R115.json",
	"r116":       "R116.json",
	"r123":       "R123.json",
	"r124":       "R124.json",
	"r125":       "R125.json",
	"r134a":      "R134a.json",
	"r-134a":     "R134a.json",
	"r141b":      "R141b.json",
	"r142b":      "R142b.json",
	"r143a":      "R143a.json",
	"r152a":      "R152A.json",
	"r161":       "R161.json",
	"r218":       "R218.json",
	"r227ea":     "R227EA.json",
	"r236ea":     "R236EA.json",
	"r236fa":     "R236FA.json",
	"r245ca":     "R245ca.json",
	"r245fa":     "R245fa.json",
	"r365mfc":    "R365MFC.json",
	"r404a":      "R404A.json",
	"r-404a":     "R404A.json",
	"r407c":      "R407C.json",
	"r-407c":     "R407C.json",
	"r410a":      "R410A.json",
	"r-410a":     "R410A.json",
	"r507a":      "R507A.json",
	"r-507a":     "R507A.json",
	"r1233zd(e)": "R1233zd(E).json",
	"r1234yf":    "R1234yf.json",
	"r-1234yf":   "R1234yf.json",
	"r1234ze(e)": "R1234ze(E).json",
	"r1234ze(z)": "R1234ze(Z).json",
	"r1243zf":    "R1243zf.json",

	// Ammonia and other inorganics
	"ammonia":         "Ammonia.json",
	"nh3":             "Ammonia.json",
	"sulfurdioxide":   "SulfurDioxide.json",
	"so2":             "SulfurDioxide.json",
	"hydrogensulfide": "HydrogenSulfide.json",
	"h2s":             "HydrogenSulfide.json",
	"nitrousoxide":    "NitrousOxide.json",
	"n2o":             "NitrousOxide.json",

	// Alcohols
	"methanol": "Methanol.json",
	"ethanol":  "Ethanol.json",

	// Aromatics
	"benzene":      "Benzene.json",
	"toluene":      "Toluene.json",
	"ethylbenzene": "EthylBenzene.json",
	"m-xylene":     "m-Xylene.json",
	"o-xylene":     "o-Xylene.json",
	"p-xylene":     "p-Xylene.json",

	// Others
	"acetone":   "Acetone.json",
	"ethylene":  "Ethylene.json",
	"propylene": "Propylene.json",
}

// GetFluidFilename returns the JSON filename for a given fluid name or alias
// Returns error if fluid is not found in registry
func GetFluidFilename(name string) (string, error) {
	// Normalize name: lowercase and remove spaces/dashes
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " ", ""), "-", ""))

	if filename, ok := FluidRegistry[normalized]; ok {
		return filename, nil
	}

	// Try exact match with .json extension
	if strings.HasSuffix(normalized, ".json") {
		return name, nil
	}

	return "", fmt.Errorf("fluid '%s' not found in registry", name)
}

// ListAvailableFluids returns a sorted list of all available fluid names
func ListAvailableFluids() []string {
	seen := make(map[string]bool)
	var fluids []string

	for alias, filename := range FluidRegistry {
		if !seen[filename] {
			seen[filename] = true
			fluids = append(fluids, alias)
		}
	}

	return fluids
}
