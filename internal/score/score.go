package score

// Input holds factors for DORA readiness scoring.
type Input struct {
	CriticalPolicies int
	HighPolicies     int
	DriftExitCode    int
	GitDirty         bool
}

const PassThreshold = 50

// Compute returns DORA readiness score clamped to [0, 100].
// Formula: 100 - (10*critical) - (5*high) - (30 if drift exit!=0) - (10 if dirty git).
func Compute(in Input) int {
	s := 100
	s -= 10 * in.CriticalPolicies
	s -= 5 * in.HighPolicies
	if in.DriftExitCode != 0 {
		s -= 30
	}
	if in.GitDirty {
		s -= 10
	}
	if s < 0 {
		return 0
	}
	if s > 100 {
		return 100
	}
	return s
}
