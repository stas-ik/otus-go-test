package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	env := make(Environment)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			continue
		}

		filePath := filepath.Join(dir, name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		lines := bytes.Split(data, []byte("\n"))
		var value string
		needRemove := len(data) == 0

		if len(lines) > 0 && len(lines[0]) > 0 {
			value = string(bytes.ReplaceAll(lines[0], []byte{0}, []byte("\n")))
			value = strings.TrimRight(value, " \t")
		} else {
			needRemove = true
		}

		env[name] = EnvValue{
			Value:      value,
			NeedRemove: needRemove,
		}
	}

	return env, nil
}
