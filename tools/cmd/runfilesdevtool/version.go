package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionCheckCmd)
	versionCmd.AddCommand(versionSetCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage module versions",
}

var versionCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify that all modules have the same version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionCheck()
	},
}

var versionSetCmd = &cobra.Command{
	Use:   "set [version]",
	Short: "Set the version in all MODULE.bazel files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionSet(args[0])
	},
}

func runVersionCheck() error {
	corePath := filepath.Join(resolvedRoot, "core/MODULE.bazel")
	goPath := filepath.Join(resolvedRoot, "go/MODULE.bazel")
	kotlinPath := filepath.Join(resolvedRoot, "kotlin/MODULE.bazel")

	coreVer, err := parseModuleVersion(corePath)
	if err != nil {
		return fmt.Errorf("failed to parse core version: %w", err)
	}
	goVer, err := parseModuleVersion(goPath)
	if err != nil {
		return fmt.Errorf("failed to parse go version: %w", err)
	}
	kotlinVer, err := parseModuleVersion(kotlinPath)
	if err != nil {
		return fmt.Errorf("failed to parse kotlin version: %w", err)
	}

	fmt.Printf("core version:   %s\n", coreVer)
	fmt.Printf("go version:     %s\n", goVer)
	fmt.Printf("kotlin version: %s\n", kotlinVer)

	mismatch := false
	if coreVer != goVer {
		fmt.Fprintf(os.Stderr, "mismatch: core version (%s) != go version (%s)\n", coreVer, goVer)
		mismatch = true
	}
	if coreVer != kotlinVer {
		fmt.Fprintf(os.Stderr, "mismatch: core version (%s) != kotlin version (%s)\n", coreVer, kotlinVer)
		mismatch = true
	}

	if mismatch {
		return fmt.Errorf("version mismatch detected")
	}

	fmt.Println("All module versions are consistent.")
	return nil
}

func runVersionSet(newVersion string) error {
	corePath := filepath.Join(resolvedRoot, "core/MODULE.bazel")
	goPath := filepath.Join(resolvedRoot, "go/MODULE.bazel")
	kotlinPath := filepath.Join(resolvedRoot, "kotlin/MODULE.bazel")

	fmt.Printf("Setting version to %s in all modules...\n", newVersion)

	if err := updateModuleVersion(corePath, newVersion); err != nil {
		return fmt.Errorf("failed to update core version: %w", err)
	}
	if err := updateModuleVersion(goPath, newVersion); err != nil {
		return fmt.Errorf("failed to update go version: %w", err)
	}
	if err := updateModuleVersion(kotlinPath, newVersion); err != nil {
		return fmt.Errorf("failed to update kotlin version: %w", err)
	}

	fmt.Println("Successfully updated all module versions.")
	return nil
}
