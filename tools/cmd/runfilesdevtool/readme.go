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

func runUpdateReadme() error {
	readmePath := filepath.Join(resolvedRoot, "README.md")
	coreModulePath := filepath.Join(resolvedRoot, "core/MODULE.bazel")

	goBuild := filepath.Join(resolvedRoot, "examples/go/BUILD.bazel")
	goUsage := filepath.Join(resolvedRoot, "examples/go/main.go")
	goGen := filepath.Join(resolvedRoot, "examples/go/bazel-bin/resources_codegen_gen.go")

	kotlinBuild := filepath.Join(resolvedRoot, "examples/kotlin/BUILD.bazel")
	kotlinUsage := filepath.Join(resolvedRoot, "examples/kotlin/Main.kt")
	kotlinGen := filepath.Join(resolvedRoot, "examples/kotlin/bazel-bin/resources_codegen_gen.kt")

	version, err := parseModuleVersion(coreModulePath)
	if err != nil {
		return fmt.Errorf("failed to parse core version: %w", err)
	}

	read := func(path string) (string, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	}

	var goBuildRaw, goUsageRaw, goGenRaw string
	var kotlinBuildRaw, kotlinUsageRaw, kotlinGenRaw string

	if goBuildRaw, err = read(goBuild); err != nil {
		return err
	}
	if goUsageRaw, err = read(goUsage); err != nil {
		return err
	}
	if goGenRaw, err = read(goGen); err != nil {
		return fmt.Errorf("%w (make sure you have run 'bazel build //...' in examples/go)", err)
	}

	if kotlinBuildRaw, err = read(kotlinBuild); err != nil {
		return err
	}
	if kotlinUsageRaw, err = read(kotlinUsage); err != nil {
		return err
	}
	if kotlinGenRaw, err = read(kotlinGen); err != nil {
		return fmt.Errorf("%w (make sure you have run 'bazel build //...' in examples/kotlin)", err)
	}

	goBuildSnippet := extractSection(goBuildRaw, "quickstart")
	kotlinBuildSnippet := extractSection(kotlinBuildRaw, "quickstart")

	readmeBytes, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	readmeContent := string(readmeBytes)

	goInstall := fmt.Sprintf("# MODULE.bazel\nbazel_dep(name = \"rules_runfile_codegen_go\", version = \"%s\")", version)
	kotlinInstall := fmt.Sprintf("# MODULE.bazel\nbazel_dep(name = \"rules_runfile_codegen_kotlin\", version = \"%s\")", version)

	type replacement struct {
		startMarker string
		endMarker   string
		language    string
		content     string
	}

	replacements := []replacement{
		{"<!-- GO_INSTALL_START -->", "<!-- GO_INSTALL_END -->", "bazel", goInstall},
		{"<!-- KOTLIN_INSTALL_START -->", "<!-- KOTLIN_INSTALL_END -->", "bazel", kotlinInstall},
		{"<!-- GO_BUILD_START -->", "<!-- GO_BUILD_END -->", "bazel", goBuildSnippet},
		{"<!-- GO_USAGE_START -->", "<!-- GO_USAGE_END -->", "go", goUsageRaw},
		{"<!-- GENERATED_GO_START -->", "<!-- GENERATED_GO_END -->", "go", goGenRaw},
		{"<!-- KOTLIN_BUILD_START -->", "<!-- KOTLIN_BUILD_END -->", "bazel", kotlinBuildSnippet},
		{"<!-- KOTLIN_USAGE_START -->", "<!-- KOTLIN_USAGE_END -->", "kotlin", kotlinUsageRaw},
		{"<!-- GENERATED_KOTLIN_START -->", "<!-- GENERATED_KOTLIN_END -->", "kotlin", kotlinGenRaw},
	}

	for _, r := range replacements {
		readmeContent, err = replaceBlock(readmeContent, r.startMarker, r.endMarker, r.language, r.content)
		if err != nil {
			return fmt.Errorf("failed to replace block %s: %w", r.startMarker, err)
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
