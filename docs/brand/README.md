# Venom visual identity

Venom's emblem combines an emerald-green snake with the `>_` terminal prompt. The artwork used in the README is based on the standalone Venom isotipo approved for the project. It is a **256 × 256 PNG with transparency**, optimized to keep the README lightweight. The full-resolution source and icon variants are maintained separately in the Venom Brand Kit.

## Brand colors

| Role | Hex |
| --- | --- |
| Primary green | `#22C55E` |
| Secondary green | `#16A34A` |
| Primary background | `#0D1117` |
| Secondary background | `#161B22` |
| Primary text | `#E5E7EB` |
| Muted text | `#9CA3AF` |

These are **design reference colors** from the approved brand guide; they are not requirements for terminal output. In particular, command-line output must remain readable with `NO_COLOR` and `TERM=dumb`.

## Usage

- Main README: use `venom-isotipo.png` as the header emblem, with accessible alternative text.
- Documentation and demos: keep existing terminal recordings and examples intact.
- Small terminal icons: prefer a simplified emblem when the detailed snake becomes hard to recognize at 16 px.
- Other Venom-family tools: reuse the colors and emblem as a visual reference, adapting their UI to their own platform.

This identity work is **visual-only**. It does not change Venom's behavior, security model, installation, or dependencies. This optimized PNG is raster art, not an editable SVG.
