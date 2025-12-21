package main

import "fmt"

var (
	release   = "develop"
	buildDate = "unknown"
	gitHash   = "unknown"
)

func printVersion() {
	fmt.Printf("Release: %s\nBuild date: %s\nGit hash: %s\n", release, buildDate, gitHash)
}
