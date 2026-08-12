package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SnakePilot10/venom/internal/detect"
)

func Executable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	return filepath.Abs(executable)
}

func Hook(sh, executable string) (string, error) {
	if executable == "" {
		return "", fmt.Errorf("venom executable path is empty")
	}
	command := Quote(executable)
	switch sh {
	case "zsh":
		return `_venom_chpwd() {
  if [ "${VENOM_LAST_PWD:-}" = "$PWD" ]; then return; fi
  if _venom_code="$(` + command + ` export zsh "$PWD")"; then
    if eval "$_venom_code"; then
      VENOM_LAST_PWD="$PWD"
      export VENOM_LAST_PWD
    fi
  fi
  unset _venom_code
}

autoload -Uz add-zsh-hook
add-zsh-hook chpwd _venom_chpwd
_venom_chpwd
`, nil
	case "bash":
		return `_venom_chpwd() {
  if [ "${VENOM_LAST_PWD:-}" = "$PWD" ]; then return; fi
  if _venom_code="$(` + command + ` export bash "$PWD")"; then
    if eval "$_venom_code"; then
      VENOM_LAST_PWD="$PWD"
      export VENOM_LAST_PWD
    fi
  fi
  unset _venom_code
}

_venom_prompt_command() {
  _venom_chpwd
}

case ";${PROMPT_COMMAND:-};" in
  *";_venom_prompt_command;"*) ;;
  *) PROMPT_COMMAND="_venom_prompt_command${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
esac
_venom_chpwd
`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q", sh)
	}
}

func Export(sh string, project detect.Project, active bool) (string, error) {
	if sh != "zsh" && sh != "bash" {
		return "", fmt.Errorf("unsupported shell %q", sh)
	}

	var b strings.Builder
	b.WriteString(cleanup())
	if !active {
		return b.String(), nil
	}

	b.WriteString("export VENOM_ACTIVE_ROOT=")
	b.WriteString(Quote(project.Root))
	b.WriteByte('\n')
	b.WriteString("export VENOM_ACTIVE_KIND=")
	b.WriteString(Quote(detect.Kind(project.Detected)))
	b.WriteByte('\n')
	b.WriteString("if [ -n \"${VIRTUAL_ENV+x}\" ]; then export VENOM_HAD_VIRTUAL_ENV=1; export VENOM_OLD_VIRTUAL_ENV=\"$VIRTUAL_ENV\"; else export VENOM_HAD_VIRTUAL_ENV=0; unset VENOM_OLD_VIRTUAL_ENV; fi\n")
	if venv := project.Env["VIRTUAL_ENV"]; venv != "" {
		b.WriteString("export VENOM_ACTIVE_VENV=")
		b.WriteString(Quote(venv))
		b.WriteByte('\n')
		b.WriteString("export VIRTUAL_ENV=")
		b.WriteString(Quote(venv))
		b.WriteByte('\n')
	}

	if len(project.PathPrepend) > 0 {
		parts := make([]string, len(project.PathPrepend))
		for i, path := range project.PathPrepend {
			parts[i] = Quote(path)
		}
		b.WriteString("export PATH=")
		b.WriteString(strings.Join(parts, ":"))
		b.WriteString(":\"$PATH\"\n")
		b.WriteString(fmt.Sprintf("export VENOM_PATH_COUNT=%d\n", len(project.PathPrepend)))
		for i, path := range project.PathPrepend {
			b.WriteString(fmt.Sprintf("export VENOM_PATH_%d=", i))
			b.WriteString(Quote(path))
			b.WriteByte('\n')
		}
	}

	keys := make([]string, 0, len(project.Env))
	for key := range project.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key == "VIRTUAL_ENV" {
			continue
		}
		value := project.Env[key]
		b.WriteString("export ")
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(Quote(value))
		b.WriteByte('\n')
	}

	return b.String(), nil
}

func cleanup() string {
	return `if [ -n "${VENOM_ACTIVE_ROOT:-}" ]; then
  if [ "${VENOM_PATH_COUNT:-0}" -gt 0 ]; then
    _venom_path_remaining="$PATH"
    PATH=""
    while [ -n "$_venom_path_remaining" ]; do
      case "$_venom_path_remaining" in
        *:*) _venom_path_entry=${_venom_path_remaining%%:*}; _venom_path_remaining=${_venom_path_remaining#*:} ;;
        *) _venom_path_entry=$_venom_path_remaining; _venom_path_remaining="" ;;
      esac
      _venom_path_match=0
      _venom_path_index=0
      while [ "$_venom_path_index" -lt "${VENOM_PATH_COUNT:-0}" ]; do
        eval "_venom_managed_path=\${VENOM_PATH_${_venom_path_index}-}"
        if [ "$_venom_path_entry" = "$_venom_managed_path" ]; then _venom_path_match=1; break; fi
        _venom_path_index=$((_venom_path_index + 1))
      done
      if [ "$_venom_path_match" = "0" ]; then
        if [ -n "$PATH" ]; then PATH="$PATH:$_venom_path_entry"; else PATH=$_venom_path_entry; fi
      fi
    done
    export PATH
    unset _venom_path_remaining _venom_path_entry _venom_path_match _venom_path_index _venom_managed_path
    _venom_path_index=0
    while [ "$_venom_path_index" -lt "${VENOM_PATH_COUNT:-0}" ]; do
      unset "VENOM_PATH_$_venom_path_index"
      _venom_path_index=$((_venom_path_index + 1))
    done
    unset _venom_path_index
  fi
  if [ "${VENOM_ACTIVE_VENV:-}" != "" ] && [ "${VIRTUAL_ENV:-}" = "$VENOM_ACTIVE_VENV" ]; then
    if [ "${VENOM_HAD_VIRTUAL_ENV:-0}" = "1" ]; then export VIRTUAL_ENV="$VENOM_OLD_VIRTUAL_ENV"; else unset VIRTUAL_ENV; fi
  fi
fi
unset VENOM_ACTIVE_ROOT VENOM_ACTIVE_KIND VENOM_ACTIVE_VENV VENOM_PATH_COUNT VENOM_HAD_VIRTUAL_ENV VENOM_OLD_VIRTUAL_ENV
`
}

func Quote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
