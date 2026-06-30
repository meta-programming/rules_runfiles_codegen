package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meta-programming/rules_runfiles_codegen/tools/internal/version"
)



// restoreRepoRoot returns the absolute path to the repo root.
// If flagVal is provided, it is used. Otherwise, it detects it.
func restoreRepoRoot(flagVal string) (string, error) {
	if flagVal != "" {
		return filepath.Abs(flagVal)
	}
	return detectRepoRoot()
}

// detectRepoRoot climbs up from the CWD to find the repo root.
func detectRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		// Check for markers
		if isRepoRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return "", fmt.Errorf("could not detect repository root (looked for .git or core/MODULE.bazel starting from %s)", cwd)
}

func isRepoRoot(dir string) bool {
	// Check for .git
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return true
	}
	// Check for core/MODULE.bazel
	if _, err := os.Stat(filepath.Join(dir, "core/MODULE.bazel")); err == nil {
		return true
	}
	return false
}

// parseModuleVersion extracts the version string from a MODULE.bazel file.
func parseModuleVersion(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return version.Parse(content)
}

// updateModuleVersion updates the version string in a MODULE.bazel file.
func updateModuleVersion(path string, newVersion string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	newContent, err := version.Update(content, newVersion)
	if err != nil {
		return err
	}
	return os.WriteFile(path, newContent, 0644)
}
