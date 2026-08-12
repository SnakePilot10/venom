package shell

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/SnakePilot10/venom/internal/detect"
)

func TestQuote(t *testing.T) {
	tests := map[string]string{
		"":                        "''",
		"simple":                  "'simple'",
		"with space":              "'with space'",
		"mi proyecto $(rm -rf ~)": "'mi proyecto $(rm -rf ~)'",
		"can't":                   "'can'\\''t'",
	}

	for input, want := range tests {
		if got := Quote(input); got != want {
			t.Fatalf("Quote(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExportNeverSourcesProjectFiles(t *testing.T) {
	project := detect.Project{
		Root:        "/tmp/project",
		Detected:    []string{"Python .venv"},
		PathPrepend: []string{"/tmp/project/.venv/bin"},
		Env:         map[string]string{"VIRTUAL_ENV": "/tmp/project/.venv"},
	}

	code, err := Export("bash", project, true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(code, "source ") || strings.Contains(code, ". ") || strings.Contains(code, "activate") {
		t.Fatalf("export contains project script execution: %s", code)
	}
	if !strings.Contains(code, "VIRTUAL_ENV='/tmp/project/.venv'") {
		t.Fatalf("export does not set VIRTUAL_ENV: %s", code)
	}
}

func TestExportPreservesUserPathChanges(t *testing.T) {
	project := detect.Project{
		Root:        "/tmp/project",
		Detected:    []string{"Go go.mod"},
		PathPrepend: []string{"/tmp/project/bin"},
		Env:         map[string]string{},
	}
	first, err := Export("bash", project, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Export("bash", project, true)
	if err != nil {
		t.Fatal(err)
	}

	script := `PATH='/user/bin'; ` + first + `PATH="$PATH:/user-added"; ` + second + `printf '%s' "$PATH"`
	output, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		t.Fatalf("run export script: %v", err)
	}
	want := "/tmp/project/bin:/user/bin:/user-added"
	if got := string(output); got != want {
		t.Fatalf("PATH after reactivation = %q, want %q", got, want)
	}
}

func TestHookUsesAbsoluteExecutableAndCachesPWD(t *testing.T) {
	code, err := Hook("bash", "/usr/local/bin/venom")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "'/usr/local/bin/venom' export bash") {
		t.Fatalf("hook does not use absolute executable: %s", code)
	}
	if !strings.Contains(code, `VENOM_LAST_PWD`) {
		t.Fatalf("hook does not cache PWD: %s", code)
	}
}

func TestHookDoesNotCachePWDWhenExportFails(t *testing.T) {
	executable := t.TempDir() + "/venom"
	if err := os.WriteFile(executable, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, err := Hook("bash", executable)
	if err != nil {
		t.Fatal(err)
	}

	output, err := exec.Command("bash", "-c", code+`printf '%s' "${VENOM_LAST_PWD:-}"`).Output()
	if err != nil {
		t.Fatalf("run failing hook: %v", err)
	}
	if got := string(output); got != "" {
		t.Fatalf("VENOM_LAST_PWD after failed export = %q, want empty", got)
	}
}

func TestExportInactiveOnlyCleansVenomState(t *testing.T) {
	code, err := Export("bash", detect.Project{}, false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(code, "export PATH=") {
		t.Fatalf("inactive export unexpectedly changes PATH: %s", code)
	}
}
