package compare

import "testing"

func TestMetricUsesAbsoluteToleranceNearZero(t *testing.T) {
	m := Metric(1e-10, 0, Tolerance{Absolute: 1e-9, Relative: 1e-6})
	if !m.Pass || m.RelativeMeaningful {
		t.Fatalf("near-zero metric = %+v", m)
	}
}

func TestNormalizeErrorClassifiesMassMolarMismatch(t *testing.T) {
	if got := Classify(1000, 18, 0.5); got != ClassMassMolar {
		t.Fatalf("classification = %q", got)
	}
}
