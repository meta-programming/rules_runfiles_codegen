package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	buildExamples bool
	updateReadmeFlag bool
	updateStardocFlag bool
)

func init() {
	updateDocsCmd.Flags().BoolVar(&buildExamples, "build", true, "Build the examples and docs before updating")
	updateDocsCmd.Flags().BoolVar(&updateReadmeFlag, "readme", true, "Update the README")
	updateDocsCmd.Flags().BoolVar(&updateStardocFlag, "stardoc", true, "Update the Stardoc markdown files")
	rootCmd.AddCommand(updateDocsCmd)
}

var updateDocsCmd = &cobra.Command{
	Use:   "update-docs",
	Short: "Synchronize README.md and Stardoc with generated outputs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdateDocs()
	},
}

func runUpdateDocs() error {
	if updateReadmeFlag {
		if err := updateReadme(); err != nil {
			return fmt.Errorf("failed to update README: %w", err)
		}
	}
	if updateStardocFlag {
		if err := updateStardoc(); err != nil {
			return fmt.Errorf("failed to update Stardoc: %w", err)
		}
	}
	return nil
}

func updateReadme() error {
	readmePath := filepath.Join(resolvedRoot, "README.md")
	coreModulePath := filepath.Join(resolvedRoot, "core/MODULE.bazel")

	version, err := parseModuleVersion(coreModulePath)
	if err != nil {
		return fmt.Errorf("failed to parse core version: %w", err)
	}

	readmeBytes, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	readmeContent := string(readmeBytes)

	read := func(path string) (string, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	}

	for _, lang := range languages {
		if buildExamples {
			fmt.Printf("Building examples for %s...\n", lang.ID)
			cmd := exec.Command("bazel", "build", "//...")
			cmd.Dir = filepath.Join(resolvedRoot, "examples", lang.ID)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build examples for %s: %w", lang.ID, err)
			}
		}

		buildPath := filepath.Join(resolvedRoot, "examples", lang.ID, "BUILD.bazel")
		usagePath := filepath.Join(resolvedRoot, "examples", lang.ID, lang.UsageFile)
		genPath := filepath.Join(resolvedRoot, "examples", lang.ID, lang.GenFile)

		buildRaw, err := read(buildPath)
		if err != nil {
			return err
		}
		usageRaw, err := read(usagePath)
		if err != nil {
			return err
		}
		genRaw, err := read(genPath)
		if err != nil {
			return fmt.Errorf("%w (make sure you have run 'bazel build //...' in examples/%s)", err, lang.ID)
		}

		buildSnippet := extractSection(buildRaw, "quickstart")
		moduleFileSnippet := fmt.Sprintf(`# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_%s", version = "%s")`, lang.ID, version)

		type replacement struct {
			startMarker string
			endMarker   string
			language    string
			content     string
		}

		replacements := []replacement{
			{fmt.Sprintf("<!-- %s_INSTALL_START -->", lang.LangMarker), fmt.Sprintf("<!-- %s_INSTALL_END -->", lang.LangMarker), "bazel", moduleFileSnippet},
			{fmt.Sprintf("<!-- %s_BUILD_START -->", lang.LangMarker), fmt.Sprintf("<!-- %s_BUILD_END -->", lang.LangMarker), "bazel", buildSnippet},
			{fmt.Sprintf("<!-- %s_USAGE_START -->", lang.LangMarker), fmt.Sprintf("<!-- %s_USAGE_END -->", lang.LangMarker), lang.Extension, usageRaw},
			{fmt.Sprintf("<!-- GENERATED_%s_START -->", lang.LangMarker), fmt.Sprintf("<!-- GENERATED_%s_END -->", lang.LangMarker), lang.Extension, genRaw},
		}

		for _, r := range replacements {
			readmeContent, err = replaceBlock(readmeContent, r.startMarker, r.endMarker, r.language, r.content)
			if err != nil {
				return fmt.Errorf("failed to replace block %s: %w", r.startMarker, err)
			}
		}
	}

	err = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	if err != nil {
		return err
	}

	fmt.Println("Successfully updated README.md with actual example and generated code.")
	return nil
}

func updateStardoc() error {
	for _, lang := range languages {
		if buildExamples {
			fmt.Printf("Building Stardoc for %s...\n", lang.ID)
			cmd := exec.Command("bazel", "build", "//docs:defs_doc")
			cmd.Dir = filepath.Join(resolvedRoot, lang.ID)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build stardoc for %s: %w", lang.ID, err)
			}
		}

		genPath := filepath.Join(resolvedRoot, lang.ID, "bazel-bin", "docs", "defs.md.gen")
		destPath := filepath.Join(resolvedRoot, lang.ID, "docs", "defs.md")

		content, err := os.ReadFile(genPath)
		if err != nil {
			return fmt.Errorf("failed to read generated stardoc for %s: %w", lang.ID, err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write stardoc for %s: %w", lang.ID, err)
		}
		fmt.Printf("Successfully updated Stardoc for %s.\n", lang.ID)
	}
	return nil
}

func replaceBlock(input, startMarker, endMarker, language, newContent string) (string, error) {
	startIdx := strings.Index(input, startMarker)
	if startIdx == -1 {
		return "", fmt.Errorf("could not find start marker %s", startMarker)
	}

	endIdx := strings.Index(input, endMarker)
	if endIdx == -1 {
		return "", fmt.Errorf("could not find end marker %s", endMarker)
	}

	if endIdx < startIdx {
		return "", fmt.Errorf("end marker %s appears before start marker %s", endMarker, startMarker)
	}

	prefix := input[:startIdx]
	suffix := input[endIdx+len(endMarker):]

	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString(startMarker)
	sb.WriteString("\n```")
	sb.WriteString(language)
	sb.WriteString("\n")
	sb.WriteString(newContent)
	sb.WriteString("\n```\n")
	sb.WriteString(endMarker)
	sb.WriteString(suffix)

	return sb.String(), nil
}

func extractSection(content, blockName string) string {
	startMarker := fmt.Sprintf("[START %s]", blockName)
	endMarker := fmt.Sprintf("[END %s]", blockName)

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}

	lineEndIdx := strings.Index(content[startIdx:], "\n")
	if lineEndIdx == -1 {
		return content
	}
	realStartIdx := startIdx + lineEndIdx + 1

	endIdx := strings.Index(content, endMarker)
	if endIdx == -1 {
		return content
	}

	return strings.TrimSpace(content[realStartIdx:endIdx])
}
