package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestColorModes(t *testing.T) {
	var output bytes.Buffer
	New(&output, ColorAlways).Header("venom", "active")
	if !strings.Contains(output.String(), "\033[") {
		t.Fatal("ColorAlways did not emit ANSI color")
	}

	output.Reset()
	New(&output, ColorNever).Header("venom", "active")
	if strings.Contains(output.String(), "\033[") {
		t.Fatal("ColorNever emitted ANSI color")
	}
}

func TestColorDoesNotChangeVisualAlignment(t *testing.T) {
	var plain, colored bytes.Buffer
	withoutColor := New(&plain, ColorNever)
	withColor := New(&colored, ColorAlways)

	withoutColor.Field("project", "~/Projects/venom")
	withoutColor.Check("version", "v0.1.0")
	withoutColor.Warning("config", "invalid")
	withColor.Field("project", "~/Projects/venom")
	withColor.Check("version", "v0.1.0")
	withColor.Warning("config", "invalid")

	if got, want := stripANSI(colored.String()), plain.String(); got != want {
		t.Fatalf("colored output alignment differs after ANSI removal:\n got: %q\nwant: %q", got, want)
	}
}

func TestNoColorDisablesAutomaticColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var output bytes.Buffer
	New(&output, ColorAuto).Header("venom", "active")
	if strings.Contains(output.String(), "\033[") {
		t.Fatal("NO_COLOR did not disable automatic color")
	}
}

func TestDisplayPathUsesHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("home directory unavailable")
	}
	if got := DisplayPath(home + "/project"); got != "~/project" {
		t.Fatalf("DisplayPath = %q, want ~/project", got)
	}
}

func stripANSI(value string) string {
	var output strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\033' || i+1 >= len(value) || value[i+1] != '[' {
			output.WriteByte(value[i])
			continue
		}
		i += 2
		for i < len(value) && (value[i] < 'A' || value[i] > 'z') {
			i++
		}
	}
	return output.String()
}
