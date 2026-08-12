package detect

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFindUsesStrongProjectMarkers(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "nested", "bin")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, ok := Find(child); ok {
		t.Fatal("Find detected a project from a generic bin directory")
	}

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	project, ok := Find(child)
	if !ok {
		t.Fatal("Find did not detect the Go project")
	}
	if project.Root != root {
		t.Fatalf("Find root = %q, want %q", project.Root, root)
	}
	if !slices.Contains(project.Detected, "Go go.mod") {
		t.Fatalf("detected = %v, want Go marker", project.Detected)
	}
}

func TestFindDoesNotTreatFilesystemBinAsProject(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "work")
	if err := os.MkdirAll(filepath.Join(child, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, ok := Find(child); ok {
		t.Fatal("Find detected a project without a strong marker")
	}
}

func TestFindDetectsVirtualEnvironmentWithoutSourcingActivation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\nname = 'demo'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	venv := filepath.Join(root, ".venv")
	if err := os.MkdirAll(filepath.Join(venv, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(venv, "bin", "activate"), []byte("rm -rf /"), 0o644); err != nil {
		t.Fatal(err)
	}

	project, ok := Find(root)
	if !ok {
		t.Fatal("Find did not detect the Python project")
	}
	if project.Env["VIRTUAL_ENV"] != venv {
		t.Fatalf("VIRTUAL_ENV = %q, want %q", project.Env["VIRTUAL_ENV"], venv)
	}
	if !slices.Contains(project.PathPrepend, filepath.Join(venv, "bin")) {
		t.Fatalf("PATH entries = %v, want virtualenv bin", project.PathPrepend)
	}
}

func TestFindDetectsVenomConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".venom.toml"), []byte("# configuration\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	project, ok := Find(root)
	if !ok {
		t.Fatal("Find did not detect .venom.toml")
	}
	if !slices.Contains(project.Detected, "Venom .venom.toml") {
		t.Fatalf("detected = %v, want Venom marker", project.Detected)
	}
}

func TestFindDetectsMultipleProjectTools(t *testing.T) {
	root := t.TempDir()
	for _, marker := range []string{"package.json", "go.mod", "Cargo.toml", "composer.json"} {
		if err := os.WriteFile(filepath.Join(root, marker), []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"node_modules/.bin", "target/debug", "vendor/bin", "bin"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	project, ok := Find(root)
	if !ok {
		t.Fatal("Find did not detect the polyglot project")
	}
	for _, marker := range []string{
		"Node package.json",
		"Go go.mod",
		"Rust Cargo.toml",
		"PHP composer.json",
		"Generic bin",
	} {
		if !slices.Contains(project.Detected, marker) {
			t.Fatalf("detected = %v, missing %q", project.Detected, marker)
		}
	}
	if len(project.PathPrepend) != 4 {
		t.Fatalf("PATH entries = %v, want four unique resource directories", project.PathPrepend)
	}
}

func TestFindAppliesVenomConfigWithoutExecutingIt(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".venom.toml"), []byte(`[env]
APP_MODE = "development"

[path]
prepend = ["tools/bin"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	project, ok := Find(root)
	if !ok {
		t.Fatal("Find did not detect configured project")
	}
	if project.ConfigError != "" {
		t.Fatalf("ConfigError = %q", project.ConfigError)
	}
	if project.Env["APP_MODE"] != "development" {
		t.Fatalf("APP_MODE = %q, want development", project.Env["APP_MODE"])
	}
	if !slices.Contains(project.PathPrepend, filepath.Join(root, "tools", "bin")) {
		t.Fatalf("PATH entries = %v, want configured path", project.PathPrepend)
	}
}
