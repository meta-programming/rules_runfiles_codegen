package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run all tests across all modules and workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTests()
	},
}

type testTarget struct {
	workspace string // relative to repo root
	target    string
}

var targets = []testTarget{
	{"core", "//internal/..."},
	{"go", "//..."},
	{"kotlin", "//..."},
	{"tests/go", "//..."},
	{"tests/kotlin", "//..."},
	{"examples/go", "//..."},
	{"examples/kotlin", "//..."},
}

func runTests() error {
	var failedWorkspaces []string
	for _, t := range targets {
		dir := filepath.Join(resolvedRoot, t.workspace)
		fmt.Printf("\n=== Running tests in %s (%s) ===\n", t.workspace, t.target)

		cmd := exec.Command("bazel", "test", "--test_output=errors", t.target)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: tests in %s failed: %v\n", t.workspace, err)
			failedWorkspaces = append(failedWorkspaces, t.workspace)
		}
	}

	if len(failedWorkspaces) > 0 {
		return fmt.Errorf("test suites failed in workspaces: %s", strings.Join(failedWorkspaces, ", "))
	}
	fmt.Println("\n=== All test suites passed! ===")
	return nil
}
