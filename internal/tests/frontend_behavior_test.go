package tests

import (
	"os/exec"
	"testing"
)

func TestFrontendImageBehavior(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("node is not available: %v", err)
	}

	cmd := exec.Command("node", "./frontend_behavior_test.mjs")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("frontend behavior checks failed: %v\n%s", err, string(out))
	}
}
