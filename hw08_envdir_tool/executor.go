package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	commandPath, err := exec.LookPath(cmd[0])
	if err != nil {
		return 1
	}

	command := exec.Command(commandPath, cmd[1:]...)

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	osEnv := os.Environ()
	newEnv := make([]string, 0, len(osEnv)+len(env))

	for _, envVar := range osEnv {
		name := strings.SplitN(envVar, "=", 2)[0]
		if _, exists := env[name]; !exists {
			newEnv = append(newEnv, envVar)
		}
	}

	for name, envVal := range env {
		if !envVal.NeedRemove {
			newEnv = append(newEnv, fmt.Sprintf("%s=%s", name, envVal.Value))
		}
	}
	command.Env = newEnv

	err = command.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return 1
	}
	return 0
}
