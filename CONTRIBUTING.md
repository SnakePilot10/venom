# Contributing

Issues and pull requests are welcome. Please describe the user-visible
problem or behavior change before proposing an implementation.

## Development setup

Requirements:

- Go 1.26 or newer
- Bash or Zsh for shell integration checks

Run the standard checks before opening a pull request:

```sh
gofmt -w ./cmd ./internal
go test ./...
go build ./cmd/venom
```

The GitHub Actions workflow also runs the race detector and `govulncheck`.

## Design principles

- Keep the core implementation in the Go standard library.
- Never execute or source files from a detected project automatically.
- Keep `export`, `hook`, and `init` free of human-facing presentation output.
- Add regression tests for security-sensitive shell behavior.

## Pull requests

- Keep changes focused and explain the compatibility impact.
- Update the README or changelog for user-visible behavior.
- Include test coverage for new branches and error cases.
- Do not include generated binaries, local tool metadata, or secrets.
