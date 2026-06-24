package domain

import "testing"

func TestSandboxTaskStateAllowsSmokeLoop(t *testing.T) {
	path := []SandboxTaskState{
		SandboxTaskStateQueued,
		SandboxTaskStateStarting,
		SandboxTaskStateRunning,
		SandboxTaskStateSucceeded,
	}

	for i := 0; i < len(path)-1; i++ {
		if err := ValidateSandboxTaskTransition(path[i], path[i+1]); err != nil {
			t.Fatalf("transition %s -> %s rejected: %v", path[i], path[i+1], err)
		}
	}
}

func TestSandboxTaskStateRejectsTerminalRestart(t *testing.T) {
	if err := ValidateSandboxTaskTransition(SandboxTaskStateSucceeded, SandboxTaskStateRunning); err == nil {
		t.Fatal("succeeded -> running transition was allowed")
	}
}
