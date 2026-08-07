package domain

import "testing"

func TestInterpretationAlertEligible(t *testing.T) {
	tests := []struct {
		name       string
		severity   string
		actionable bool
		want       bool
	}{
		{"critical noise", "critical", false, true},
		{"high noise", "high", false, true},
		{"medium actionable", "medium", true, true},
		{"medium noise", "medium", false, false},
		{"low actionable", "low", true, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := (Interpretation{Severity: test.severity, Actionable: test.actionable}).AlertEligible()
			if got != test.want {
				t.Fatalf("AlertEligible() = %v, want %v", got, test.want)
			}
		})
	}
}
