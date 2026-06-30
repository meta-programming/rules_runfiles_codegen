// Program runfilesdevtool is a Swiss Army knife for developing rules_runfile_codegen.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoRootFlag string
	resolvedRoot string
)

var rootCmd = &cobra.Command{
	Use:   "runfilesdevtool",
	Short: "Developer tool for rules_runfile_codegen",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		resolvedRoot, err = restoreRepoRoot(repoRootFlag)
		return err
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&repoRootFlag, "repo-root", "r", "", "path to the repository root (auto-detected if omitted)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
