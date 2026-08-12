package availability

import "testing"

func TestAggregate(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, "unknown"}, {[]string{"unknown", "up"}, "unknown"}, {[]string{"up", "up"}, "operational"},
		{[]string{"down", "up"}, "degraded"}, {[]string{"down", "down"}, "major_outage"},
	}
	for _, tc := range cases {
		if got := Aggregate(tc.in); got != tc.want {
			t.Fatalf("Aggregate(%v)=%s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestSanitizeError(t *testing.T) {
	if got := sanitizeError(fakeError("Get https://secret.example/token: connection refused")); got != "connection refused" {
		t.Fatal(got)
	}
}

type fakeError string

func (e fakeError) Error() string { return string(e) }
