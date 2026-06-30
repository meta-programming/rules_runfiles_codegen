package main

import (
	"encoding/json"
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

func TestValidatePresubmit_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yml")

	// Invalid YAML syntax
	err := os.WriteFile(path, []byte("invalid: yaml: :"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidatePresubmit(path); err == nil {
		t.Error("expected error for invalid YAML syntax, got nil")
	}
}

func TestValidatePresubmit_MissingTasks(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "missing_tasks.yml")

	content := `
matrix:
  platform: ["ubuntu2004"]
`
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidatePresubmit(path); err == nil {
		t.Error("expected error for missing tasks, got nil")
	}
}

func TestValidatePresubmit_TasksNotDict(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "tasks_not_dict.yml")

	content := `
tasks:
  - task1
  - task2
`
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidatePresubmit(path); err == nil {
		t.Error("expected error for tasks not being a dictionary, got nil")
	}
}

func TestValidatePresubmit_MissingPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "missing_platform.yml")

	content := `
tasks:
  verify_targets:
    name: "Verify"
    build_targets:
      - "//..."
`
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidatePresubmit(path); err == nil {
		t.Error("expected error for missing platform, got nil")
	}
}

func TestValidatePresubmit_MissingTargets(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "missing_targets.yml")

	content := `
tasks:
  verify_targets:
    name: "Verify"
    platform: "ubuntu2004"
`
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidatePresubmit(path); err == nil {
		t.Error("expected error for missing targets, got nil")
	}
}

func TestValidateMetadataTemplate_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(path, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := validation.ValidateMetadataTemplate(path); err == nil {
		t.Error("expected error for invalid JSON syntax, got nil")
	}
}

func TestValidateMetadataTemplate_MissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	
	mandatoryFields := []string{"name", "homepage", "maintainers", "versions", "yanked_versions"}
	
	for _, missingField := range mandatoryFields {
		t.Run("missing_"+missingField, func(t *testing.T) {
			path := filepath.Join(tmpDir, "missing_"+missingField+".json")
			
			// Construct JSON with one missing field
			contentMap := map[string]interface{}{
				"name":            "rules_runfile_codegen_core",
				"homepage":        "https://github.com/meta-programming/rules_runfiles_codegen",
				"maintainers":     []interface{}{},
				"versions":        []interface{}{},
				"yanked_versions": map[string]interface{}{},
			}
			delete(contentMap, missingField)
			
			data, err := json.Marshal(contentMap)
			if err != nil {
				t.Fatal(err)
			}
			
			err = os.WriteFile(path, data, 0644)
			if err != nil {
				t.Fatal(err)
			}
			
			if err := validation.ValidateMetadataTemplate(path); err == nil {
				t.Errorf("expected error for missing field %q, got nil", missingField)
			}
		})
	}
}
