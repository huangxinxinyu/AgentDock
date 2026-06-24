package domain

import "testing"

func TestSandboxStateAllowsManualLifecycle(t *testing.T) {
	path := []SandboxState{
		SandboxStateCreating,
		SandboxStateReady,
		SandboxStatePaused,
		SandboxStateReady,
		SandboxStateClosing,
		SandboxStateClosed,
	}

	for i := 0; i < len(path)-1; i++ {
		if err := ValidateSandboxTransition(path[i], path[i+1]); err != nil {
			t.Fatalf("transition %s -> %s rejected: %v", path[i], path[i+1], err)
		}
	}
}

func TestSandboxStateAllowsIdempotentLifecycleActions(t *testing.T) {
	for _, state := range []SandboxState{SandboxStateReady, SandboxStatePaused, SandboxStateClosed} {
		if err := ValidateSandboxTransition(state, state); err != nil {
			t.Fatalf("idempotent transition for %s rejected: %v", state, err)
		}
	}
}

func TestSandboxStateRejectsTerminalResume(t *testing.T) {
	if err := ValidateSandboxTransition(SandboxStateClosed, SandboxStateReady); err == nil {
		t.Fatal("closed -> ready transition was allowed")
	}
}
