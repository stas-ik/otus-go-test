package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "envdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFiles := map[string]string{
		"FOO":   "123",
		"BAR":   "value\nsecond line",
		"EMPTY": "",
		"NULL":  "test\x00null",
	}

	for name, content := range testFiles {
		filePath := filepath.Join(tempDir, name)
		err := os.WriteFile(filePath, []byte(content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	env, err := ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]EnvValue{
		"FOO":   {Value: "123", NeedRemove: false},
		"BAR":   {Value: "value", NeedRemove: false},
		"EMPTY": {Value: "", NeedRemove: true},
		"NULL":  {Value: "test\nnull", NeedRemove: false},
	}

	if len(env) != len(expected) {
		t.Errorf("Expected %d env vars, got %d", len(expected), len(env))
	}

	for name, expectedVal := range expected {
		if val, exists := env[name]; !exists {
			t.Errorf("Expected env var %s not found", name)
		} else if val != expectedVal {
			t.Errorf("For %s, expected %+v, got %+v", name, expectedVal, val)
		}
	}
}
