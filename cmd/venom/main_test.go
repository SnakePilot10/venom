package main

import (
	"testing"

	"github.com/SnakePilot10/venom/internal/ui"
)

func TestParseOptions(t *testing.T) {
	args, mode, err := parseOptions([]string{"status", "--color", "never"})
	if err != nil {
		t.Fatal(err)
	}
	if mode != ui.ColorNever || len(args) != 1 || args[0] != "status" {
		t.Fatalf("args = %v, mode = %q", args, mode)
	}
}

func TestSuggestCommand(t *testing.T) {
	if got := suggestCommand("staus"); got != "status" {
		t.Fatalf("suggestCommand(staus) = %q, want status", got)
	}
	if got := suggestCommand("completely-unknown"); got != "" {
		t.Fatalf("suggestCommand(unknown) = %q, want empty", got)
	}
}
