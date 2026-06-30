package validation

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Task represents a task in presubmit.yml
// Task represents a task in presubmit.yml.
// Its structure is defined and parsed by the Bazel CI system's bazelci.py script.
// See: https://github.com/bazelbuild/continuous-integration/blob/master/buildkite/bazelci.py
type Task struct {
	Name         string   `yaml:"name"`
	Platform     string   `yaml:"platform"`
	Bazel        string   `yaml:"bazel"`
	BuildTargets []string `yaml:"build_targets"`
	TestTargets  []string `yaml:"test_targets"`
}

// PresubmitConfig represents the structure of presubmit.yml
type PresubmitConfig struct {
	Matrix interface{}     `yaml:"matrix"`
	Tasks  map[string]Task `yaml:"tasks"`
}

// ValidatePresubmit validates the presubmit.yml file.
// It verifies that the config matches the expected Bazel CI structure to prevent
// failures during the Buildkite pipeline generation phase (bcr_presubmit.py).
// See: https://github.com/bazelbuild/continuous-integration/blob/master/buildkite/bazel-central-registry/bcr_presubmit.py
func ValidatePresubmit(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// We first unmarshal into a generic map to check if 'tasks' is indeed a map/dictionary
	// and not something else (like a list).
	var generic map[string]interface{}
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return fmt.Errorf("failed to parse YAML (generic): %w", err)
	}

	tasksRaw, ok := generic["tasks"]
	if !ok {
		return fmt.Errorf("missing 'tasks' key")
	}

	if _, ok := tasksRaw.(map[string]interface{}); !ok {
		return fmt.Errorf("'tasks' must be a dictionary/map")
	}

	// Now parse into structured data
	var config PresubmitConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML (structured): %w", err)
	}

	if len(config.Tasks) == 0 {
		return fmt.Errorf("no tasks defined")
	}

	for taskName, task := range config.Tasks {
		if task.Platform == "" {
			return fmt.Errorf("task %q: missing 'platform'", taskName)
		}
		if task.Bazel == "" {
			return fmt.Errorf("task %q: missing 'bazel' version", taskName)
		}
		if len(task.BuildTargets) == 0 && len(task.TestTargets) == 0 {
			return fmt.Errorf("task %q: must have at least one of 'build_targets' or 'test_targets'", taskName)
		}
	}

	return nil
}

// MetadataTemplate represents the structure of metadata.template.json
type MetadataTemplate struct {
	Name           *string                 `json:"name"`
	Homepage       *string                 `json:"homepage"`
	Maintainers    *[]interface{}          `json:"maintainers"`
	Versions       *[]interface{}          `json:"versions"`
	YankedVersions *map[string]interface{} `json:"yanked_versions"`
}

// ValidateMetadataTemplate validates the metadata.template.json file.
func ValidateMetadataTemplate(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse into struct
	var template MetadataTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if template.Name == nil || *template.Name == "" {
		return fmt.Errorf("field 'name' must be a non-empty string")
	}
	if template.Homepage == nil || *template.Homepage == "" {
		return fmt.Errorf("field 'homepage' must be a non-empty string")
	}
	if template.Maintainers == nil {
		return fmt.Errorf("field 'maintainers' must not be null/missing")
	}
	if template.Versions == nil {
		return fmt.Errorf("field 'versions' must not be null/missing")
	}
	if template.YankedVersions == nil {
		return fmt.Errorf("field 'yanked_versions' must not be null/missing")
	}

	return nil
}
