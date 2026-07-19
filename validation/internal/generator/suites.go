package generator

import (
	"fmt"
	"math"
	"sort"
)

var qualities = []float64{0, 0.01, 0.1, 0.25, 0.5, 0.75, 0.9, 0.99, 1}

func Anchors(fluid string, e Envelope) []Case {
	t := func(f float64) float64 { return e.TMin + (e.TMax-e.TMin)*f }
	p := func(f float64) float64 { return loglerp(e.PMin, e.PMax, int(math.Round(f*10)), 11) }
	values := [][2]float64{{0.01, 0.01}, {0.25, 0.01}, {0.5, 0.1}, {0.75, 0.5}, {0.99, 0.5}, {0.5, 0.99}, {0.99, 0.99}, {0.01, 0.99}, {0.25, 0.9}, {0.75, 0.1}, {0.99, 0.1}, {0.5, 0.01}}
	out := make([]Case, 0, len(values))
	for i, value := range values {
		out = append(out, Case{ID: fmt.Sprintf("%s/anchor/PT/Dmass/%04d", fluid, i), Fluid: fluid, Stage: "anchor", Input1: "T", Input2: "P", Output: "Dmass", Value1: t(value[0]), Value2: p(value[1])})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func Saturation(fluid string, e Envelope, temperaturePoints int) []Case {
	if temperaturePoints < 1 {
		return nil
	}
	out := make([]Case, 0, temperaturePoints*len(qualities))
	for i := 0; i < temperaturePoints; i++ {
		f := 0.001 + 0.998*float64(i)/math.Max(1, float64(temperaturePoints-1))
		t := e.TMin + (e.TMax-e.TMin)*f
		for j, q := range qualities {
			out = append(out, Case{ID: fmt.Sprintf("%s/saturation/TQ/%04d-%02d", fluid, i, j), Fluid: fluid, Stage: "saturation", Input1: "T", Input2: "Q", Output: "P", Value1: t, Value2: q})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func InvalidInputs(fluid string, e Envelope) []Case {
	out := []Case{
		{ID: fluid + "/invalid/nan", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "P", Output: "Dmass", Value1: math.NaN(), Value2: e.PMin},
		{ID: fluid + "/invalid/positive-infinity", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "P", Output: "Dmass", Value1: math.Inf(1), Value2: e.PMin},
		{ID: fluid + "/invalid/negative-infinity", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "P", Output: "Dmass", Value1: math.Inf(-1), Value2: e.PMin},
		{ID: fluid + "/invalid/negative-pressure", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "P", Output: "Dmass", Value1: e.TMin, Value2: -1},
		{ID: fluid + "/invalid/zero-density", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "Dmolar", Output: "P", Value1: e.TMin, Value2: 0},
		{ID: fluid + "/invalid/negative-density", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "Dmolar", Output: "P", Value1: e.TMin, Value2: -1},
		{ID: fluid + "/invalid/quality-low", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "Q", Output: "P", Value1: e.TMin, Value2: -0.1},
		{ID: fluid + "/invalid/quality-high", Fluid: fluid, Stage: "invalid", Input1: "T", Input2: "Q", Output: "P", Value1: e.TMin, Value2: 1.1},
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func Adaptive(fluid string, base Case, grid int) []Case {
	if grid < 1 {
		return nil
	}
	out := make([]Case, 0, grid*grid)
	for i := 0; i < grid; i++ {
		for j := 0; j < grid; j++ {
			di := (float64(i) - float64(grid-1)/2) / math.Max(1, float64(grid-1))
			dj := (float64(j) - float64(grid-1)/2) / math.Max(1, float64(grid-1))
			c := base
			c.Stage = "adaptive"
			c.ID = fmt.Sprintf("%s/adaptive/%02d-%02d", fluid, i, j)
			c.Value1 = base.Value1 * (1 + 1e-3*di)
			c.Value2 = base.Value2 * (1 + 1e-3*dj)
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
