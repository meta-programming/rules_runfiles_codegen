package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meta-programming/rules_runfiles_codegen/tools/internal/validation"
)

func TestValidateBCRFiles(t *testing.T) {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		t.Fatalf("failed to detect repo root: %v", err)
	}

	modules := []string{"core", "go", "kotlin"}

	for _, mod := range modules {
		t.Run(mod+"/presubmit", func(t *testing.T) {
			path := filepath.Join(repoRoot, ".bcr", mod, "presubmit.yml")
			if err := validation.ValidatePresubmit(path); err != nil {
				t.Errorf("validation failed for %s: %v", path, err)
			}
		})

		t.Run(mod+"/metadata", func(t *testing.T) {
			path := filepath.Join(repoRoot, ".bcr", mod, "metadata.template.json")
			if err := validation.ValidateMetadataTemplate(path); err != nil {
				t.Errorf("validation failed for %s: %v", path, err)
			}
		})
	}
}

func TestValidatePresubmit(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		expectedError string
	}{
		{
			name:          "Invalid YAML syntax",
			yamlContent:   "invalid: yaml: :",
			expectedError: "failed to parse YAML",
		},
		{
			name: "Missing tasks key",
			yamlContent: `
matrix:
  platform: ["ubuntu2004"]
`,
			expectedError: "missing 'tasks' key",
		},
		{
			name: "Tasks is not a dictionary",
			yamlContent: `
tasks:
  - task1
  - task2
`,
			expectedError: "'tasks' must be a dictionary/map",
		},
		{
			name: "Task missing platform",
			yamlContent: `
tasks:
  verify_targets:
    name: "Verify"
    bazel: "7.x"
    build_targets:
      - "//..."
`,
			expectedError: `task "verify_targets": missing 'platform'`,
		},
		{
			name: "Task missing bazel version",
			yamlContent: `
tasks:
  verify_targets:
    name: "Verify"
    platform: "ubuntu2004"
    build_targets:
      - "//..."
`,
			expectedError: `task "verify_targets": missing 'bazel' version`,
		},
		{
			name: "Task missing build/test targets",
			yamlContent: `
tasks:
  verify_targets:
    name: "Verify"
    platform: "ubuntu2004"
    bazel: "7.x"
`,
			expectedError: `task "verify_targets": must have at least one of 'build_targets' or 'test_targets'`,
		},
	}

	tmpDir := t.TempDir()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test.yml")
			if err := os.WriteFile(path, []byte(tc.yamlContent), 0644); err != nil {
				t.Fatal(err)
			}

			err := validation.ValidatePresubmit(path)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tc.expectedError != "" && !contains(err.Error(), tc.expectedError) {
				t.Errorf("expected error containing %q, got %q", tc.expectedError, err.Error())
			}
		})
	}
}

func TestValidateMetadataTemplate(t *testing.T) {
	tests := []struct {
		name          string
		jsonContent   string
		expectedError string
	}{
		{
			name:          "Invalid JSON syntax",
			jsonContent:   "{invalid json}",
			expectedError: "failed to parse JSON",
		},
		{
			name: "Missing name",
			jsonContent: `
{
	"homepage": "https://github.com/meta-programming/rules_runfiles_codegen",
	"maintainers": [],
	"versions": [],
	"yanked_versions": {}
}
`,
			expectedError: "field 'name' must be a non-empty string",
		},
		{
			name: "Missing homepage",
			jsonContent: `
{
	"name": "rules_runfile_codegen_core",
	"maintainers": [],
	"versions": [],
	"yanked_versions": {}
}
`,
			expectedError: "field 'homepage' must be a non-empty string",
		},
		{
			name: "Missing maintainers",
			jsonContent: `
{
	"name": "rules_runfile_codegen_core",
	"homepage": "https://github.com/meta-programming/rules_runfiles_codegen",
	"versions": [],
	"yanked_versions": {}
}
`,
			expectedError: "field 'maintainers' must not be null/missing",
		},
		{
			name: "Missing versions",
			jsonContent: `
{
	"name": "rules_runfile_codegen_core",
	"homepage": "https://github.com/meta-programming/rules_runfiles_codegen",
	"maintainers": [],
	"yanked_versions": {}
}
`,
			expectedError: "field 'versions' must not be null/missing",
		},
		{
			name: "Missing yanked_versions",
			jsonContent: `
{
	"name": "rules_runfile_codegen_core",
	"homepage": "https://github.com/meta-programming/rules_runfiles_codegen",
	"maintainers": [],
	"versions": []
}
`,
			expectedError: "field 'yanked_versions' must not be null/missing",
		},
	}

	tmpDir := t.TempDir()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test.json")
			if err := os.WriteFile(path, []byte(tc.jsonContent), 0644); err != nil {
				t.Fatal(err)
			}

			err := validation.ValidateMetadataTemplate(path)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tc.expectedError != "" && !contains(err.Error(), tc.expectedError) {
				t.Errorf("expected error containing %q, got %q", tc.expectedError, err.Error())
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || stringsContains(s, substr))
}

// Simple fallback since strings package is not imported in this test file yet
func stringsContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
