# Security Policy

## Supported versions

Security fixes target the latest release on the `master` branch. Venom is a
0.x project, so behavior and configuration may change between releases.

## Reporting a vulnerability

Please do not open a public issue for a suspected vulnerability. Use GitHub's
private vulnerability reporting if it is enabled for this repository:

https://github.com/SnakePilot10/venom/security/advisories/new

Include the affected version, operating system and shell, reproduction steps,
and the smallest project layout or configuration needed to reproduce the issue.
Do not include secrets or destructive payloads in a report.

Venom's core security boundary is that it inspects project files but never
automatically executes or sources them. Reports involving shell output,
`PATH`, hooks, `.venom.toml`, or command shadowing are especially valuable.
