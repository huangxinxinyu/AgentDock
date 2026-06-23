package runtime

import (
	"context"
	"testing"
)

func TestNoopRunnerSatisfiesRunnerBoundary(t *testing.T) {
	var runner Runner = NoopRunner{}
	result, err := runner.Start(context.Background(), StartRequest{
		RunID:  "run_123",
		Prompt: "summarize repository",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if result.Runtime != "noop" {
		t.Fatalf("Runtime = %q, want noop", result.Runtime)
	}
}
