package integration_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/example/project/repo/tests/go/test_resources"
)

func TestDataFile(t *testing.T) {
	resolved, err := test_resources.DataFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DataFile: %v", err)
	}

	path := resolved.Path()
	if path == "" {
		t.Fatal("DataFile path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read DataFile: %v", err)
	}

	expected := "dummy content"
	got := strings.TrimSpace(string(content))
	if got != expected {
		t.Errorf("DataFile content = %q, want %q", got, expected)
	}
}

func TestHelperTool(t *testing.T) {
	// Test the MustResolve API
	helper := test_resources.HelperTool.MustResolve()
	cmd := helper.Cmd()
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("HelperTool failed to run: %v\nStderr: %s", err, stderr.String())
	}

	expected := "helper data content"
	got := strings.TrimSpace(stdout.String())
	if got != expected {
		t.Errorf("HelperTool output = %q, want %q\nStderr: %s", got, expected, stderr.String())
	}
}

func TestExternalFile(t *testing.T) {
	resolved, err := test_resources.ExternalFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve ExternalFile: %v", err)
	}

	path := resolved.Path()
	if path == "" {
		t.Fatal("ExternalFile path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read ExternalFile: %v", err)
	}

	if len(content) == 0 {
		t.Error("ExternalFile is empty")
	}

	if !strings.Contains(string(content), "Apache License") {
		t.Errorf("ExternalFile content does not contain 'Apache License'. Got: %s", string(content[:100]))
	}
}
