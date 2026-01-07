package main

import (
	"fmt"
	"os"
)

// These are replaced at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Add a version command
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("goblog %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	} else {
		println("Hello, World!")
	}
}
