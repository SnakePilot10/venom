# Examples

These examples describe the marker files Venom recognizes. They are intentionally
documentation-only so this repository does not vendor generated environments or
package-manager state.

## Python

```text
python-app/
├── pyproject.toml
└── .venv/
    └── bin/
```

Venom sets `VIRTUAL_ENV` and prepends `.venv/bin` without sourcing `activate`.

## Node.js

```text
node-app/
├── package.json
└── node_modules/.bin/
```

`node_modules/.bin` is prepended when it exists.

## Go

```text
go-app/
├── go.mod
└── bin/
```

The project `bin/` directory is treated as a PATH resource after `go.mod`
identifies the project root.

## Rust

```text
rust-app/
├── Cargo.toml
└── target/debug/
```

Existing `target/debug` and `target/release` directories are added to PATH.

## Polyglot projects

A project can contain multiple recognized markers. Venom reports every detector
and combines their known PATH resources deterministically.

## Explicit configuration

Use `.venom.toml` for project-specific, literal settings:

```toml
[env]
APP_ENV = "development"

[path]
prepend = ["tools/bin"]
```

No example executes project files. This is a core safety property of Venom.
