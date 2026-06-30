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
	Short: "Verify that core, go, and kotlin modules have the same version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionCheck()
	},
}

var versionSetCmd = &cobra.Command{
	Use:   "set [version]",
	Short: "Set the version in core, go, and kotlin MODULE.bazel files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersionSet(args[0])
	},
}

func runVersionCheck() error {
	// Construct modules list dynamically: core + all languages
	modules := []string{"core"}
	for _, lang := range languages {
		modules = append(modules, lang.ID)
	}

	versions := make(map[string]string)
	for _, mod := range modules {
		path := filepath.Join(resolvedRoot, mod, "MODULE.bazel")
		ver, err := parseModuleVersion(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s version: %w", mod, err)
		}
		versions[mod] = ver
		fmt.Printf("%-15s %s\n", mod+":", ver)
	}

	// Check consistency against the first module (core)
	firstMod := modules[0]
	firstVer := versions[firstMod]
	mismatch := false

	for i := 1; i < len(modules); i++ {
		mod := modules[i]
		ver := versions[mod]
		if ver != firstVer {
			fmt.Fprintf(os.Stderr, "mismatch: %s version (%s) != %s version (%s)\n", firstMod, firstVer, mod, ver)
			mismatch = true
		}
	}

	if mismatch {
		return fmt.Errorf("version mismatch detected")
	}

	fmt.Println("Core, go, and kotlin module versions are consistent.")
	return nil
}

func runVersionSet(newVersion string) error {
	// Construct modules list dynamically: core + all languages
	modules := []string{"core"}
	for _, lang := range languages {
		modules = append(modules, lang.ID)
	}

	fmt.Printf("Setting version to %s in core, go, and kotlin modules...\n", newVersion)

	for _, mod := range modules {
		path := filepath.Join(resolvedRoot, mod, "MODULE.bazel")
		if err := updateModuleVersion(path, newVersion); err != nil {
			return fmt.Errorf("failed to update %s version: %w", mod, err)
		}
	}

	fmt.Println("Successfully updated core, go, and kotlin module versions.")
	return nil
}
