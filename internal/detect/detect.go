package detect

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SnakePilot10/venom/internal/config"
)

type Project struct {
	Root        string
	Detected    []string
	PathPrepend []string
	Env         map[string]string
	ConfigError string
}

func Find(start string) (Project, bool) {
	root, ok := findRoot(start)
	if !ok {
		return Project{}, false
	}

	project := Project{Root: root, Env: map[string]string{}}
	detectVenom(&project)
	detectPython(&project)
	detectNode(&project)
	detectGo(&project)
	detectRust(&project)
	detectPHP(&project)
	detectGeneric(&project)
	applyConfig(&project)
	project.PathPrepend = unique(project.PathPrepend)
	sort.Strings(project.Detected)
	return project, len(project.Detected) > 0
}

func applyConfig(project *Project) {
	path := filepath.Join(project.Root, ".venom.toml")
	cfg, err := config.Load(path)
	if err != nil {
		if !os.IsNotExist(err) {
			project.ConfigError = err.Error()
		}
		return
	}
	for key, value := range cfg.Env {
		project.Env[key] = value
	}
	for _, path := range cfg.PathPrepend {
		if !filepath.IsAbs(path) {
			path = filepath.Join(project.Root, path)
		}
		project.PathPrepend = append(project.PathPrepend, filepath.Clean(path))
	}
}

func detectVenom(project *Project) {
	if exists(filepath.Join(project.Root, ".venom.toml")) {
		project.Detected = append(project.Detected, "Venom .venom.toml")
	}
}

func findRoot(start string) (string, bool) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err == nil && !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	for {
		if hasMarker(abs) {
			return abs, true
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", false
		}
		abs = parent
	}
}

func hasMarker(dir string) bool {
	markers := []string{
		".git", ".venom.toml", "uv.lock", "pyproject.toml", "package.json",
		"go.mod", "Cargo.toml", "composer.json",
	}
	for _, marker := range markers {
		if exists(filepath.Join(dir, marker)) {
			return true
		}
	}
	return false
}

func detectPython(project *Project) {
	for _, name := range []string{".venv", "venv", "env"} {
		venv := filepath.Join(project.Root, name)
		if dirExists(venv) {
			project.Detected = append(project.Detected, "Python "+name)
			project.Env["VIRTUAL_ENV"] = venv
			bin := filepath.Join(venv, "bin")
			if dirExists(bin) {
				project.PathPrepend = append(project.PathPrepend, bin)
			}
			return
		}
	}
	if exists(filepath.Join(project.Root, "pyproject.toml")) {
		project.Detected = append(project.Detected, "Python pyproject.toml")
	}
	if exists(filepath.Join(project.Root, "uv.lock")) {
		project.Detected = append(project.Detected, "Python uv.lock")
	}
}

func detectNode(project *Project) {
	if !exists(filepath.Join(project.Root, "package.json")) {
		return
	}
	project.Detected = append(project.Detected, "Node package.json")
	bin := filepath.Join(project.Root, "node_modules", ".bin")
	if dirExists(bin) {
		project.PathPrepend = append(project.PathPrepend, bin)
	}
}

func detectGo(project *Project) {
	if !exists(filepath.Join(project.Root, "go.mod")) {
		return
	}
	project.Detected = append(project.Detected, "Go go.mod")
}

func detectRust(project *Project) {
	if !exists(filepath.Join(project.Root, "Cargo.toml")) {
		return
	}
	project.Detected = append(project.Detected, "Rust Cargo.toml")
	for _, dir := range []string{"target/debug", "target/release"} {
		path := filepath.Join(project.Root, filepath.FromSlash(dir))
		if dirExists(path) {
			project.PathPrepend = append(project.PathPrepend, path)
		}
	}
}

func detectPHP(project *Project) {
	if !exists(filepath.Join(project.Root, "composer.json")) {
		return
	}
	project.Detected = append(project.Detected, "PHP composer.json")
	bin := filepath.Join(project.Root, "vendor", "bin")
	if dirExists(bin) {
		project.PathPrepend = append(project.PathPrepend, bin)
	}
}

func detectGeneric(project *Project) {
	bin := filepath.Join(project.Root, "bin")
	if dirExists(bin) {
		project.Detected = append(project.Detected, "Generic bin")
		project.PathPrepend = append(project.PathPrepend, bin)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func Kind(detected []string) string {
	return strings.Join(detected, ",")
}
