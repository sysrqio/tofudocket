package score

import "testing"

func TestCompute(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want int
	}{
		{"clean", Input{}, 100},
		{"one critical", Input{CriticalPolicies: 1}, 90},
		{"two high", Input{HighPolicies: 2}, 90},
		{"drift", Input{DriftExitCode: 2}, 70},
		{"dirty", Input{GitDirty: true}, 90},
		{"fail combo", Input{CriticalPolicies: 6}, 40},
		{"clamp low", Input{CriticalPolicies: 20}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compute(tt.in); got != tt.want {
				t.Fatalf("Compute() = %d, want %d", got, tt.want)
			}
		})
	}
}
