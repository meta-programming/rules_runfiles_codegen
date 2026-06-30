package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var parallel bool

func init() {
	defaultParallel := os.Getenv("CI") == "true"
	testCmd.Flags().BoolVar(&parallel, "parallel", defaultParallel, "Run tests in parallel")
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

type testResult struct {
	target testTarget
	output string
	err    error
}

func runTests() error {
	if parallel {
		return runTestsParallel()
	}
	return runTestsSequential()
}

func runTestsSequential() error {
	failed := false
	for _, t := range targets {
		dir := filepath.Join(resolvedRoot, t.workspace)
		fmt.Printf("\n=== Running tests in %s (%s) ===\n", t.workspace, t.target)

		cmd := exec.Command("bazel", "test", t.target)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: tests in %s failed: %v\n", t.workspace, err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("some test suites failed")
	}
	fmt.Println("\n=== All test suites passed! ===")
	return nil
}

func runTestsParallel() error {
	fmt.Printf("Running %d test suites in parallel...\n", len(targets))

	results := make([]testResult, len(targets))
	var wg sync.WaitGroup

	for i, t := range targets {
		wg.Add(1)
		go func(idx int, target testTarget) {
			defer wg.Done()

			dir := filepath.Join(resolvedRoot, target.workspace)
			cmd := exec.Command("bazel", "test", target.target)
			cmd.Dir = dir

			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf

			err := cmd.Run()
			results[idx] = testResult{
				target: target,
				output: buf.String(),
				err:    err,
			}
		}(i, t)
	}

	wg.Wait()

	failed := false
	for _, r := range results {
		fmt.Printf("\n=== Results for %s (%s) ===\n", r.target.workspace, r.target.target)
		fmt.Print(r.output)
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "Error: tests in %s failed: %v\n", r.target.workspace, r.err)
			failed = true
		} else {
			fmt.Printf("tests in %s passed\n", r.target.workspace)
		}
	}

	if failed {
		return fmt.Errorf("some test suites failed")
	}
	fmt.Println("\n=== All test suites passed! ===")
	return nil
}
