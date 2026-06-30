// Package main implements a tool to synchronize the README.md with actual
// generated code and example code from the quickstart examples.
//
// This tool ensures that the documentation always reflects the exact,
// compiler-verified output of the generators and the actual working
// configuration and example files. This prevents stale, incorrect, or
// non-working code examples in the README.
//
// For a complete overview of the development workflow, integration tests,
// and how to use this tool, please refer to the Developer Guide:
//   DEV_GUIDE.md
//
// Usage:
//   Run from the repository root:
//     go run tools/update_readme.go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Define paths relative to the repository root.
	readmePath := "README.md"
	
	// Source files from Go quickstart example.
	goBuildPath := "examples/go/BUILD.bazel"
	goUsagePath := "examples/go/main.go"
	goGenPath := "examples/go/bazel-bin/resources_codegen_gen.go"
	
	// Source files from Kotlin quickstart example.
	kotlinBuildPath := "examples/kotlin/BUILD.bazel"
	kotlinUsagePath := "examples/kotlin/Main.kt"
	kotlinGenPath := "examples/kotlin/bazel-bin/resources_codegen_gen.kt"

	// Check if we are running from the repo root. If README.md isn't here,
	// try the parent directory (in case the user ran it from inside tools/).
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readmePath = "../README.md"
		
		goBuildPath = "../examples/go/BUILD.bazel"
		goUsagePath = "../examples/go/main.go"
		goGenPath = "../examples/go/bazel-bin/resources_codegen_gen.go"
		
		kotlinBuildPath = "../examples/kotlin/BUILD.bazel"
		kotlinUsagePath = "../examples/kotlin/Main.kt"
		kotlinGenPath = "../examples/kotlin/bazel-bin/resources_codegen_gen.kt"
	}

	// If we still can't find README.md, fail with a clear usage message.
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error: Could not find README.md.")
		fmt.Fprintln(os.Stderr, "Please run this tool from the repository root:")
		fmt.Fprintln(os.Stderr, "  go run tools/update_readme.go")
		os.Exit(1)
	}

	// Helper to read a file and fail if it doesn't exist/can't be read.
	readOrFail := func(path string, helpMsg string) string {
		content, err := readAndTrimFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			if helpMsg != "" {
				fmt.Fprintln(os.Stderr, helpMsg)
			}
			os.Exit(1)
		}
		return content
	}

	// Read Go files.
	goBuildRaw := readOrFail(goBuildPath, "")
	goUsage := readOrFail(goUsagePath, "")
	goGen := readOrFail(goGenPath, "Make sure you have run 'bazel build //...' in examples/go first.")

	// Read Kotlin files.
	kotlinBuildRaw := readOrFail(kotlinBuildPath, "")
	kotlinUsage := readOrFail(kotlinUsagePath, "")
	kotlinGen := readOrFail(kotlinGenPath, "Make sure you have run 'bazel build //...' in examples/kotlin first.")

	// Extract only the quickstart section from the BUILD files to hide test targets.
	goBuild := extractSection(goBuildRaw, "quickstart")
	kotlinBuild := extractSection(kotlinBuildRaw, "quickstart")

	// Read the current README.md.
	readmeBytes, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading README.md: %v\n", err)
		os.Exit(1)
	}
	readmeContent := string(readmeBytes)

	// Define all replacements.
	replacements := []struct {
		startMarker string
		endMarker   string
		language    string
		content     string
	}{
		// Go Section
		{"<!-- GO_BUILD_START -->", "<!-- GO_BUILD_END -->", "bazel", goBuild},
		{"<!-- GO_USAGE_START -->", "<!-- GO_USAGE_END -->", "go", goUsage},
		{"<!-- GENERATED_GO_START -->", "<!-- GENERATED_GO_END -->", "go", goGen},

		// Kotlin Section
		{"<!-- KOTLIN_BUILD_START -->", "<!-- KOTLIN_BUILD_END -->", "bazel", kotlinBuild},
		{"<!-- KOTLIN_USAGE_START -->", "<!-- KOTLIN_USAGE_END -->", "kotlin", kotlinUsage},
		{"<!-- GENERATED_KOTLIN_START -->", "<!-- GENERATED_KOTLIN_END -->", "kotlin", kotlinGen},
	}

	// Apply all replacements.
	for _, r := range replacements {
		readmeContent, err = replaceBlock(readmeContent, r.startMarker, r.endMarker, r.language, r.content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error replacing block %s: %v\n", r.startMarker, err)
			os.Exit(1)
		}
	}

	// Write the updated content back to README.md.
	err = os.WriteFile(readmePath, []byte(readmeContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated README.md with actual example and generated code!")
}

// readAndTrimFile reads the file at the given path and returns its content
// as a string, with leading and trailing whitespace trimmed.
func readAndTrimFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

// replaceBlock finds the section in 'input' delimited by 'startMarker' and 'endMarker',
// and replaces it with 'newContent' wrapped in a markdown code block of the given 'language'.
//
// This function uses pure string manipulation (slicing) instead of regular expressions
// to avoid escaping issues with characters like '$' in the replacement content.
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

// extractSection extracts the content between "[START blockName]" and "[END blockName]"
// in the input string. If the markers are not found, it returns the whole input.
// This allows us to hide boilerplate/test targets from the README examples.
func extractSection(content, blockName string) string {
	startMarker := fmt.Sprintf("[START %s]", blockName)
	endMarker := fmt.Sprintf("[END %s]", blockName)

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}

	// Find the end of the line containing the start marker to avoid including it.
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
