package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env         map[string]string
	PathPrepend []string
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	cfg := Config{Env: map[string]string{}}
	section := ""
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			if section != "env" && section != "path" {
				return Config{}, fmt.Errorf("line %d: unsupported section %q", lineNumber, section)
			}
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("line %d: expected key = value", lineNumber)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return Config{}, fmt.Errorf("line %d: invalid key or value", lineNumber)
		}

		switch section {
		case "env":
			if !validEnvKey(key) {
				return Config{}, fmt.Errorf("line %d: invalid environment variable %q", lineNumber, key)
			}
			parsed, err := parseString(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			cfg.Env[key] = parsed
		case "path":
			if key != "prepend" {
				return Config{}, fmt.Errorf("line %d: unsupported path key %q", lineNumber, key)
			}
			parsed, err := parseStringArray(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			cfg.PathPrepend = append(cfg.PathPrepend, parsed...)
		default:
			return Config{}, fmt.Errorf("line %d: keys must be inside [env] or [path]", lineNumber)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validEnvKey(key string) bool {
	if key == "PATH" || key == "VIRTUAL_ENV" || strings.HasPrefix(key, "VENOM_") {
		return false
	}
	for i, r := range key {
		if i == 0 && r != '_' && !asciiLetter(r) {
			return false
		}
		if i > 0 && r != '_' && !asciiLetter(r) && !asciiDigit(r) {
			return false
		}
	}
	return key != ""
}

func asciiLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func asciiDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func parseString(value string) (string, error) {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("expected a double-quoted string")
	}
	parsed, err := strconv.Unquote(value)
	if err != nil {
		return "", fmt.Errorf("invalid string: %w", err)
	}
	if parsed == "" {
		return "", fmt.Errorf("empty values are not allowed")
	}
	return parsed, nil
}

func parseStringArray(value string) ([]string, error) {
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return nil, fmt.Errorf("expected an array of double-quoted strings")
	}
	contents := strings.TrimSpace(value[1 : len(value)-1])
	if contents == "" {
		return nil, nil
	}
	var values []string
	for len(contents) > 0 {
		contents = strings.TrimSpace(contents)
		if contents[0] != '"' {
			return nil, fmt.Errorf("array values must be double-quoted strings")
		}
		end := 1
		escaped := false
		for end < len(contents) {
			if contents[end] == '"' && !escaped {
				break
			}
			escaped = contents[end] == '\\' && !escaped
			if contents[end] != '\\' {
				escaped = false
			}
			end++
		}
		if end >= len(contents) {
			return nil, fmt.Errorf("unterminated string in array")
		}
		parsed, err := parseString(contents[:end+1])
		if err != nil {
			return nil, err
		}
		values = append(values, parsed)
		contents = strings.TrimSpace(contents[end+1:])
		if contents == "" {
			break
		}
		if contents[0] != ',' {
			return nil, fmt.Errorf("expected comma between array values")
		}
		contents = contents[1:]
	}
	return values, nil
}

func stripComment(line string) string {
	quoted := false
	escaped := false
	for i, r := range line {
		if r == '"' && !escaped {
			quoted = !quoted
		}
		if r == '#' && !quoted {
			return line[:i]
		}
		escaped = r == '\\' && !escaped
		if r != '\\' {
			escaped = false
		}
	}
	return line
}
