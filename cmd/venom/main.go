package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SnakePilot10/venom/internal/detect"
	"github.com/SnakePilot10/venom/internal/shell"
	"github.com/SnakePilot10/venom/internal/ui"
	"github.com/SnakePilot10/venom/internal/version"
)

func main() {
	args, mode, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := run(args, mode); err != nil {
		printError(err, mode)
		os.Exit(1)
	}
}

func run(args []string, mode ui.ColorMode) error {
	cmd := "status"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}
	if cmd != "help" && hasHelp(args) && isKnownCommand(cmd) {
		printCommandHelp(cmd, mode)
		return nil
	}

	switch cmd {
	case "detect":
		start := argOrPWD(args)
		project, ok := detect.Find(start)
		printDetect(ui.New(os.Stdout, mode), start, project, ok)
		return nil
	case "export":
		sh := "zsh"
		if len(args) > 0 {
			sh = args[0]
			args = args[1:]
		}
		start := argOrPWD(args)
		project, ok := detect.Find(start)
		code, err := shell.Export(sh, project, ok)
		if err != nil {
			return err
		}
		fmt.Print(code)
		return nil
	case "hook":
		sh := argOr(args, "zsh")
		executable, err := shell.Executable()
		if err != nil {
			return fmt.Errorf("resolve venom executable: %w", err)
		}
		code, err := shell.Hook(sh, executable)
		if err != nil {
			return err
		}
		fmt.Print(code)
		return nil
	case "init":
		sh := argOr(args, "zsh")
		executable, err := shell.Executable()
		if err != nil {
			return fmt.Errorf("resolve venom executable: %w", err)
		}
		fmt.Printf("eval \"$(%s hook %s)\"\n", shell.Quote(executable), sh)
		return nil
	case "status":
		return status(mode)
	case "doctor":
		return doctor(mode)
	case "version", "--version":
		u := ui.New(os.Stdout, mode)
		u.Banner("venom", "v"+version.Version)
		return nil
	case "help", "-h", "--help":
		printHelp(mode)
		return nil
	default:
		return unknownCommandError{command: cmd}
	}
}

func status(mode ui.ColorMode) error {
	u := ui.New(os.Stdout, mode)
	if root := os.Getenv("VENOM_ACTIVE_ROOT"); root != "" {
		if !dirExists(root) {
			u.Header("venom", "stale")
			u.Field("project", "missing "+ui.DisplayPath(root))
			return nil
		}
		project, ok := detect.Find(root)
		printStatus(u, root, project, ok, true)
		return nil
	}

	project, ok := detect.Find(mustPWD())
	printStatus(u, mustPWD(), project, ok, false)
	return nil
}

func doctor(mode ui.ColorMode) error {
	u := ui.New(os.Stdout, mode)
	u.Title("venom doctor")
	healthy := true
	hookHealthy := true
	u.Check("version", "v"+version.Version)
	executable, executableErr := shell.Executable()
	if executableErr != nil {
		u.Warning("binary", "unable to resolve executable: "+executableErr.Error())
		healthy = false
	} else {
		u.Check("binary", ui.DisplayPath(executable))
	}

	if resolved, err := exec.LookPath("venom"); err != nil {
		u.Warning("PATH", "venom is not available through PATH")
		healthy = false
	} else if executableErr == nil && sameFile(resolved, executable) {
		u.Check("PATH", "resolves correctly")
	} else {
		u.Warning("PATH", "resolves a different executable: "+ui.DisplayPath(resolved))
		healthy = false
	}

	shellName := filepath.Base(os.Getenv("SHELL"))
	if shellName == "bash" || shellName == "zsh" {
		u.Check("shell", shellName)
	} else {
		u.Warning("shell", "unsupported or unknown shell")
		healthy = false
	}

	if root := os.Getenv("VENOM_ACTIVE_ROOT"); root != "" {
		if dirExists(root) {
			u.Check("project", ui.DisplayPath(root))
		} else {
			u.Warning("project", "missing directory "+ui.DisplayPath(root))
			healthy = false
		}
	} else {
		project, ok := detect.Find(mustPWD())
		if ok {
			u.Check("project", ui.DisplayPath(project.Root))
		} else {
			u.Check("project", "no recognized environment")
		}
	}
	if os.Getenv("VENOM_LAST_PWD") == "" {
		u.Warning("hook", "not loaded")
		healthy = false
		hookHealthy = false
	} else {
		u.Check("hook", "active")
	}

	project, ok := detect.Find(mustPWD())
	if ok {
		if project.ConfigError != "" {
			u.Warning("config", project.ConfigError)
			healthy = false
		} else {
			u.Check("config", configState(project))
		}
	}
	if healthy {
		u.Success("Everything looks healthy.")
	} else if !hookHealthy {
		u.Hint("hint", "Add the shell hook with:")
		u.Indented(`eval "$(venom hook ` + shellNameOrDefault(shellName) + `)"`)
	} else {
		u.Hint("hint", "Review the warnings above.")
	}
	return nil
}

func printStatus(u *ui.UI, root string, project detect.Project, ok, active bool) {
	state := "inactive"
	if active {
		state = "active"
	}
	u.Header("venom", state)
	u.Field("project", ui.DisplayPath(root))
	if !ok {
		if active {
			u.Warning("stack", "environment details unavailable")
		} else {
			u.Field("stack", "no recognized environment")
		}
		return
	}
	u.Field("stack", strings.Join(stackNames(project.Detected), " · "))
	u.Field("config", configState(project))
	if venv := project.Env["VIRTUAL_ENV"]; venv != "" {
		u.Field("venv", relativePath(project.Root, venv))
	}
	printPathSection(u, project)
}

func printDetect(u *ui.UI, start string, project detect.Project, ok bool) {
	u.Header("venom detect", ui.DisplayPath(start))
	if !ok {
		u.Warning("project", "no recognized environment")
		return
	}
	for _, item := range project.Detected {
		label, value := detectorParts(item)
		u.Check(label, value)
	}
	if project.ConfigError != "" {
		u.Warning("config", project.ConfigError)
	}
	if len(project.Env) > 0 {
		u.Section("environment")
		keys := make([]string, 0, len(project.Env))
		for key := range project.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := project.Env[key]
			if key == "VIRTUAL_ENV" {
				value = relativePath(project.Root, value)
			}
			u.Field(key, value)
		}
	}
	printPathSection(u, project)
}

func printPathSection(u *ui.UI, project detect.Project) {
	if len(project.PathPrepend) == 0 {
		return
	}
	u.Section("path")
	for _, path := range project.PathPrepend {
		u.Added(relativePath(project.Root, path))
	}
}

func detectorParts(item string) (string, string) {
	parts := strings.SplitN(item, " ", 2)
	if len(parts) == 1 {
		return item, "detected"
	}
	return parts[0], parts[1]
}

func stackNames(detected []string) []string {
	seen := map[string]bool{}
	var names []string
	for _, item := range detected {
		label, _ := detectorParts(item)
		if label == "Venom" || seen[label] {
			continue
		}
		seen[label] = true
		names = append(names, label)
	}
	sort.Strings(names)
	return names
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return rel
	}
	return ui.DisplayPath(path)
}

func configState(project detect.Project) string {
	if project.ConfigError != "" {
		return "invalid"
	}
	if _, err := os.Stat(filepath.Join(project.Root, ".venom.toml")); err == nil {
		return ".venom.toml"
	}
	return "none"
}

func shellNameOrDefault(name string) string {
	if name == "bash" || name == "zsh" {
		return name
	}
	return "zsh"
}

func sameFile(first, second string) bool {
	first, err := filepath.EvalSymlinks(first)
	if err != nil {
		return false
	}
	second, err = filepath.EvalSymlinks(second)
	if err != nil {
		return false
	}
	return first == second
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func printHelp(mode ui.ColorMode) {
	u := ui.New(os.Stdout, mode)
	u.Title("venom")
	u.Line("Automatic project environments without executing project code.")
	u.Section("Usage")
	u.Line("  venom <command> [options]")
	u.Section("Commands")
	u.Field("status", "Show the current environment")
	u.Field("detect", "Inspect a project")
	u.Field("doctor", "Diagnose your setup")
	u.Field("version", "Show the installed version")
	u.Section("Shell")
	u.Field("init", "Print the shell setup command")
	u.Field("hook", "Generate the shell hook")
	u.Field("export", "Generate environment changes")
	u.Section("Options")
	u.Field("--color", "auto, always, or never")
	u.MutedText("Run 'venom <command> --help' for more information.")
}

func printCommandHelp(command string, mode ui.ColorMode) {
	u := ui.New(os.Stdout, mode)
	u.Title("venom " + command)
	switch command {
	case "status":
		u.Line("Show the current activation state.")
		u.Section("Usage")
		u.Line("  venom status [--color=auto|always|never]")
	case "detect":
		u.Line("Inspect project markers and the environment plan.")
		u.Section("Usage")
		u.Line("  venom detect [dir] [--color=auto|always|never]")
	case "doctor":
		u.Line("Diagnose the executable, shell hook, project, and configuration.")
		u.Section("Usage")
		u.Line("  venom doctor [--color=auto|always|never]")
	case "version":
		u.Line("Show the installed Venom version.")
		u.Section("Usage")
		u.Line("  venom version")
	case "init":
		u.Line("Print the command that installs the shell hook.")
		u.Section("Usage")
		u.Line("  venom init zsh|bash")
	case "hook":
		u.Line("Generate a shell hook for automatic activation.")
		u.Section("Usage")
		u.Line("  venom hook zsh|bash")
	case "export":
		u.Line("Generate shell code for the environment at a directory.")
		u.Section("Usage")
		u.Line("  venom export zsh|bash [dir]")
	}
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func isKnownCommand(command string) bool {
	switch command {
	case "status", "detect", "doctor", "version", "init", "hook", "export":
		return true
	default:
		return false
	}
}

func parseOptions(args []string) ([]string, ui.ColorMode, error) {
	mode := ui.ColorAuto
	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--color" {
			if i+1 >= len(args) {
				return nil, mode, fmt.Errorf("--color requires auto, always, or never")
			}
			i++
			arg = "--color=" + args[i]
		}
		if strings.HasPrefix(arg, "--color=") {
			value := strings.TrimPrefix(arg, "--color=")
			switch ui.ColorMode(value) {
			case ui.ColorAuto, ui.ColorAlways, ui.ColorNever:
				mode = ui.ColorMode(value)
				continue
			default:
				return nil, mode, fmt.Errorf("invalid color mode %q", value)
			}
		}
		filtered = append(filtered, arg)
	}
	return filtered, mode, nil
}

type unknownCommandError struct {
	command string
}

func (e unknownCommandError) Error() string {
	return fmt.Sprintf("unknown command %q", e.command)
}

func printError(err error, mode ui.ColorMode) {
	u := ui.New(os.Stderr, mode)
	if unknown, ok := err.(unknownCommandError); ok {
		u.Error(fmt.Sprintf("unknown command '%s'", unknown.command))
		if suggestion := suggestCommand(unknown.command); suggestion != "" {
			u.Hint("hint", fmt.Sprintf("Did you mean 'venom %s'?", suggestion))
		}
		u.Line("  Run 'venom help' to see available commands.")
		return
	}
	u.Error(err.Error())
}

func suggestCommand(input string) string {
	commands := []string{"status", "detect", "doctor", "version", "init", "hook", "export", "help"}
	best, distance := "", 3
	for _, command := range commands {
		if d := levenshtein(input, command); d < distance {
			best, distance = command, d
		}
	}
	return best
}

func levenshtein(a, b string) int {
	previous := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i, ra := range a {
		current := make([]int, len(b)+1)
		current[0] = i + 1
		for j, rb := range b {
			cost := 0
			if ra != rb {
				cost = 1
			}
			current[j+1] = min(current[j]+1, previous[j+1]+1, previous[j]+cost)
		}
		previous = current
	}
	return previous[len(b)]
}

func min(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func argOrPWD(args []string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return mustPWD()
}

func argOr(args []string, fallback string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return fallback
}

func mustPWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return pwd
}
