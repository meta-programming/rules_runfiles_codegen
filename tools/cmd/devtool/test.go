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
		fmt.Printf("\n=== Running Bazel tests in %s (%s) ===\n", t.workspace, t.target)

		cmd := exec.Command("bazel", "test", "--test_output=errors", t.target)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Bazel tests in %s failed: %v\n", t.workspace, err)
			failedWorkspaces = append(failedWorkspaces, t.workspace)
		}
	}

	// Run Go tests in tools directory
	fmt.Println("\n=== Running Go tests in tools ===")
	goTestCmd := exec.Command("go", "test", "./...")
	goTestCmd.Dir = filepath.Join(resolvedRoot, "tools")
	goTestCmd.Stdout = os.Stdout
	goTestCmd.Stderr = os.Stderr
	if err := goTestCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Go tests in tools failed: %v\n", err)
		failedWorkspaces = append(failedWorkspaces, "tools (go)")
	}

	if len(failedWorkspaces) > 0 {
		return fmt.Errorf("test suites failed: %s", strings.Join(failedWorkspaces, ", "))
	}
	fmt.Println("\n=== All test suites passed! ===")
	return nil
}
