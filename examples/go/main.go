package main

import (
	"fmt"
	"os"

	"github.com/example/project/examples/go/resources"
)

func main() {
	// 1. Access the resolved runfile path.
	// Resources are resolved at startup (init-time).
	path := resources.DataFile.Path()

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading runfile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Data: %s\n", string(content))

	// 2. Run an executable runfile with env propagation.
	cmd := resources.HelperTool.Cmd()
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running helper: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Helper output: %s", string(output))
}
