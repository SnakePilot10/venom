# Venom Demo

`venom.tape` is the reproducible source for the README demo. It expects `vhs`,
`ttyd`, a Chromium-compatible browser, and an installed `venom` command, then
writes:

- `venom-demo.gif`
- `venom-detect.png`

The demo project lives at `~/venom-demo` and contains Python, Node.js, and
`.venom.toml` markers. Recreate it with the files documented in the repository
setup instructions or use the prepared local fixture when working from the
author's Termux workspace.

Render from the repository root:

```sh
vhs docs/demo/venom.tape
```

If VHS cannot find Venom, install the latest release or put the compiled binary
on `PATH` before rendering:

```sh
export PATH="$HOME/Projects/Go/venom:$PATH"
vhs docs/demo/venom.tape
```

On Termux, `vhs` and `ttyd` can be installed independently, but the Chromium
package may require a compatible repository mirror and dependencies. VHS cannot
render the GIF or screenshot until a browser binary is available.
