# Changelog

## [0.1.1] - 2026-08-12

### Fixed

- Detect Python projects that use `requirements.txt`, `requirements-dev.txt`, `Pipfile`, `Pipfile.lock`, or `poetry.lock` without requiring `pyproject.toml` or `uv.lock`.
- Treat `.venv`, `venv`, and `env` as project-root markers only when `pyvenv.cfg` confirms a real Python virtual environment.

## [0.1.0] - 2026-08-11

Initial public release.

### Included

- Automatic Python, Node.js, Go, Rust, Composer, and generic `bin/` detection.
- Declarative `.venom.toml` support for literal environment values and PATH entries.
- Bash and Zsh hooks with machine-safe shell output.
- Human-friendly CLI output with adaptive color and command-specific help.
- Linux amd64 and arm64 release archives with `SHA256SUMS`.
- CI with formatting, tests, race detector, vet, build, and vulnerability scanning.

### Security

- Project files are inspected but never automatically sourced or executed.
- Configuration cannot execute commands or modify Venom's reserved state variables.
