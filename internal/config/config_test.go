package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLiteralEnvironmentAndPaths(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".venom.toml")
	contents := `[env]
APP_MODE = "development # local"

[path]
prepend = ["bin", "tools/bin"]
`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Env["APP_MODE"]; got != "development # local" {
		t.Fatalf("APP_MODE = %q, want literal value", got)
	}
	if len(cfg.PathPrepend) != 2 || cfg.PathPrepend[0] != "bin" || cfg.PathPrepend[1] != "tools/bin" {
		t.Fatalf("PathPrepend = %v, want configured paths", cfg.PathPrepend)
	}
}

func TestLoadRejectsReservedAndExecutableConfiguration(t *testing.T) {
	for name, contents := range map[string]string{
		"reserved environment": "[env]\nVENOM_ACTIVE_ROOT = \"/tmp\"\n",
		"unsupported section":  "[run]\ncommand = \"echo unsafe\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".venom.toml")
			if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("Load accepted unsafe configuration")
			}
		})
	}
}

func TestLoadRequiresExplicitSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".venom.toml")
	if err := os.WriteFile(path, []byte("FOO = \"bar\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load accepted a top-level key")
	}
}
