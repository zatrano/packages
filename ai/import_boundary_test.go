package ai_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestAIDoesNotImportAgentRAGWorkflow(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", "github.com/zatrano/packages/ai").CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "github.com/zatrano/packages/rag",
			"github.com/zatrano/packages/agent",
			"github.com/zatrano/packages/workflow":
			t.Fatalf("forbidden import %s", line)
		}
	}
}
