package main

import (
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		env := Environment{
			"TEST_VAR": EnvValue{Value: "test_value", NeedRemove: false},
		}
		cmd := []string{"bash", "-c", "echo $TEST_VAR"}

		tmpFile, err := os.CreateTemp("", "test_output")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())

		oldStdout := os.Stdout
		os.Stdout = tmpFile
		defer func() { os.Stdout = oldStdout }()

		returnCode := RunCmd(cmd, env)
		if returnCode != 0 {
			t.Errorf("Expected return code 0, got %d", returnCode)
		}

		tmpFile.Close()
		output, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}

		if string(output) != "test_value\n" {
			t.Errorf("Expected output 'test_value\n', got '%s'", string(output))
		}
	})

	t.Run("Command not found", func(t *testing.T) {
		cmd := []string{"nonexistent_command"}
		returnCode := RunCmd(cmd, Environment{})
		if returnCode == 0 {
			t.Error("Expected non-zero return code for nonexistent command")
		}
	})
}
