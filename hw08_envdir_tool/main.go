package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: go-envdir <envdir> <command> [args...]")
		os.Exit(1)
	}

	envDir := os.Args[1]
	cmd := os.Args[2:]

	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading env dir: %v\n", err)
		os.Exit(1)
	}

	returnCode := RunCmd(cmd, env)
	os.Exit(returnCode)
}
