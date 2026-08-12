package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

type UI struct {
	w     io.Writer
	color bool
}

func New(w io.Writer, mode ColorMode) *UI {
	color := mode == ColorAlways || mode == ColorAuto && colorAllowed(w)
	return &UI{w: w, color: color}
}

func (u *UI) ColorEnabled() bool {
	return u.color
}

func (u *UI) Title(name string) {
	u.printf("%s\n", u.brand(name))
}

func (u *UI) Line(format string, args ...any) {
	u.printf(format+"\n", args...)
}

func (u *UI) Header(name, state string) {
	u.printf("%s  %s\n", u.brand(name), u.strong(state))
}

func (u *UI) Banner(name, state string) {
	u.printf("%s %s\n", u.brand(name), u.strong(state))
}

func (u *UI) Field(label, value string) {
	u.printf("  %s %s\n", u.muted(fmt.Sprintf("%-14s", label)), value)
}

func (u *UI) Check(label, value string) {
	u.printf("  %s %s %s\n", u.success("✓"), u.muted(fmt.Sprintf("%-10s", label)), value)
}

func (u *UI) Warning(label, value string) {
	u.printf("  %s %s %s\n", u.warning("!"), u.muted(fmt.Sprintf("%-10s", label)), value)
}

func (u *UI) Failure(label, value string) {
	u.printf("  %s %s %s\n", u.failure("×"), u.muted(fmt.Sprintf("%-10s", label)), value)
}

func (u *UI) Added(value string) {
	u.printf("    %s %s\n", u.info("+"), value)
}

func (u *UI) Section(name string) {
	u.printf("\n  %s\n", u.strong(name))
}

func (u *UI) Hint(label, value string) {
	u.printf("\n  %s %s\n", u.warning(label), value)
}

func (u *UI) Indented(value string) {
	for _, line := range strings.Split(value, "\n") {
		u.printf("        %s\n", line)
	}
}

func (u *UI) Success(message string) {
	u.printf("\n  %s\n", u.success(message))
}

func (u *UI) Error(message string) {
	u.printf("%s %s\n", u.failure("error:"), message)
}

func (u *UI) MutedText(message string) {
	u.printf("%s\n", u.muted(message))
}

func DisplayPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		if path == home {
			return "~"
		}
		prefix := home + string(filepath.Separator)
		if strings.HasPrefix(path, prefix) {
			return "~" + strings.TrimPrefix(path, home)
		}
	}
	return path
}

func colorAllowed(w io.Writer) bool {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func (u *UI) brand(value string) string   { return u.paint("\033[32;1m", value) }
func (u *UI) strong(value string) string  { return u.paint("\033[1m", value) }
func (u *UI) muted(value string) string   { return u.paint("\033[2m", value) }
func (u *UI) success(value string) string { return u.paint("\033[32m", value) }
func (u *UI) info(value string) string    { return u.paint("\033[36m", value) }
func (u *UI) warning(value string) string { return u.paint("\033[33m", value) }
func (u *UI) failure(value string) string { return u.paint("\033[31m", value) }

func (u *UI) paint(code, value string) string {
	if !u.color {
		return value
	}
	return code + value + "\033[0m"
}

func (u *UI) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(u.w, format, args...)
}
