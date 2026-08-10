package domain

import "testing"

func TestAlertPolicyEligibility(t *testing.T) {
	policy := AlertPolicy{Enabled: true, AlwaysAlertSeverities: []string{"critical", "high"}, ActionableAlertSeverities: []string{"medium"}}
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
			got := policy.Eligible(Interpretation{Severity: test.severity, Actionable: test.actionable})
			if got != test.want {
				t.Fatalf("Eligible() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestDisabledAlertPolicy(t *testing.T) {
	if (AlertPolicy{}).Eligible(Interpretation{Severity: "critical"}) {
		t.Fatal("disabled policy must reject every interpretation")
	}
}
