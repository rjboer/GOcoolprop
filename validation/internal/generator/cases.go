package generator

import (
	"fmt"
	"math"
)

type Case struct {
	ID, Fluid, Stage, Input1, Input2, Output string
	Value1, Value2                           float64
}
type Envelope struct{ TMin, TMax, PMin, PMax, RhoMin, RhoMax float64 }
type Config struct {
	TDTemperaturePoints, TDDensityPoints, PTTemperaturePoints, PTPressurePoints, QuasiRandomPoints int
	Seed                                                                                           int64
}

func Screening(fluid string, e Envelope, c Config) []Case {
	if c.Seed == 0 {
		c.Seed = 1
	}
	var out []Case
	for i := 0; i < c.TDTemperaturePoints; i++ {
		t := lerp(e.TMin, e.TMax, i, c.TDTemperaturePoints)
		for j := 0; j < c.TDDensityPoints; j++ {
			rho := loglerp(e.RhoMin, e.RhoMax, j, c.TDDensityPoints)
			out = append(out, Case{ID: fmt.Sprintf("%s/screen/TD/%04d-%04d", fluid, i, j), Fluid: fluid, Stage: "screen", Input1: "T", Input2: "Dmolar", Output: "P", Value1: t, Value2: rho})
		}
	}
	for i := 0; i < c.PTTemperaturePoints; i++ {
		t := lerp(e.TMin, e.TMax, i, c.PTTemperaturePoints)
		for j := 0; j < c.PTPressurePoints; j++ {
			p := loglerp(e.PMin, e.PMax, j, c.PTPressurePoints)
			out = append(out, Case{ID: fmt.Sprintf("%s/screen/PT/%04d-%04d", fluid, i, j), Fluid: fluid, Stage: "screen", Input1: "T", Input2: "P", Output: "Dmass", Value1: t, Value2: p})
		}
	}
	for i := 0; i < c.QuasiRandomPoints; i++ {
		x := halton(i+1, 2)
		y := halton(i+1, 3)
		out = append(out, Case{ID: fmt.Sprintf("%s/screen/quasi/%04d", fluid, i), Fluid: fluid, Stage: "screen", Input1: "T", Input2: "P", Output: "Dmass", Value1: e.TMin + (e.TMax-e.TMin)*x, Value2: math.Exp(math.Log(e.PMin) + (math.Log(e.PMax)-math.Log(e.PMin))*y)})
	}
	return out
}
func lerp(a, b float64, i, n int) float64 {
	if n <= 1 {
		return (a + b) / 2
	}
	return a + (b-a)*float64(i)/float64(n-1)
}
func loglerp(a, b float64, i, n int) float64 {
	if a <= 0 || b <= 0 {
		return lerp(a, b, i, n)
	}
	return math.Exp(lerp(math.Log(a), math.Log(b), i, n))
}
func halton(index, base int) float64 {
	f, r := 1.0, 0.0
	for index > 0 {
		f /= float64(base)
		r += f * float64(index%base)
		index /= base
	}
	return r
}
