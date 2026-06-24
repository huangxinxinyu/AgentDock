package domain

import "testing"

func TestRunStateAllowsWorkerHappyPath(t *testing.T) {
	path := []RunState{
		RunStateQueued,
		RunStateProvisioning,
		RunStatePreparingWorkspace,
		RunStateRunning,
		RunStateCompleted,
	}

	for i := 0; i < len(path)-1; i++ {
		if err := ValidateRunTransition(path[i], path[i+1]); err != nil {
			t.Fatalf("transition %s -> %s rejected: %v", path[i], path[i+1], err)
		}
	}
}

func TestRunStateRejectsTerminalTransition(t *testing.T) {
	if err := ValidateRunTransition(RunStateCompleted, RunStateRunning); err == nil {
		t.Fatal("completed -> running transition was allowed")
	}
}

func TestRunStateRejectsPatchReviewAsRunState(t *testing.T) {
	if IsRunState("awaiting_review") {
		t.Fatal("awaiting_review should not be a Sprint 2 run state")
	}
}
