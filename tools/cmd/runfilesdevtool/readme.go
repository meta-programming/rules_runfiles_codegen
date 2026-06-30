package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateReadmeCmd)
}

var updateReadmeCmd = &cobra.Command{
	Use:   "update-readme",
	Short: "Synchronize README.md with actual generated code and examples",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdateReadme()
	},
}

// LanguageConfig defines the paths and markers for a language module's examples.
type LanguageConfig struct {
	ID         string // "go", "kotlin" (used in paths: examples/<ID>/)
	LangMarker string // "GO", "KOTLIN" (used in README markers: <!-- <MARKER>_INSTALL_START -->)
	Extension  string // "go", "kotlin" (used for markdown code block syntax highlighting)
	UsageFile  string // "main.go", "Main.kt" (relative to examples/<ID>/)
	GenFile    string // "bazel-bin/...go" (relative to examples/<ID>/)
}

var languages = []LanguageConfig{
	{
		ID:         "go",
		LangMarker: "GO",
		Extension:  "go",
		UsageFile:  "main.go",
		GenFile:    "bazel-bin/resources_codegen_gen.go",
	},
	{
		ID:         "kotlin",
		LangMarker: "KOTLIN",
		Extension:  "kotlin",
		UsageFile:  "Main.kt",
		GenFile:    "bazel-bin/resources_codegen_gen.kt",
	},
}

func runUpdateReadme() error {
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
