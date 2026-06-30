package runfile_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/meta-programming/rules_runfiles_codegen/go/runfile"
)

// mockResolver implements runfile.Resolver.
type mockResolver struct {
	paths map[string]string
	errs  map[string]error
}

func (m *mockResolver) Rlocation(path string) (string, error) {
	if err, ok := m.errs[path]; ok {
		return "", err
	}
	if p, ok := m.paths[path]; ok {
		return p, nil
	}
	return "", fmt.Errorf("runfile %q not found", path)
}

func TestResolve_Success(t *testing.T) {
	t.Fatal("injected failure")
	mock := &mockResolver{
		paths: map[string]string{
			"my_workspace/data.txt": "/absolute/path/to/data.txt",
		},
	}

	file := runfile.New("my_workspace/data.txt")

	// Test Resolve with explicit resolver option
	resolved, err := file.Resolve(runfile.WithResolver(mock))
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if resolved.Rlocation() != "my_workspace/data.txt" {
		t.Errorf("Rlocation() = %q, want %q", resolved.Rlocation(), "my_workspace/data.txt")
	}
	if resolved.Path() != "/absolute/path/to/data.txt" {
		t.Errorf("Path() = %q, want %q", resolved.Path(), "/absolute/path/to/data.txt")
	}
}

func TestResolve_Error(t *testing.T) {
	mock := &mockResolver{
		errs: map[string]error{
			"my_workspace/missing.txt": fmt.Errorf("permission denied"),
		},
	}

	file := runfile.New("my_workspace/missing.txt")

	_, err := file.Resolve(runfile.WithResolver(mock))
	if err == nil {
		t.Fatal("Resolve() succeeded, want error")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error = %v, want it to contain 'permission denied'", err)
	}
}

func TestMustResolve_Panic(t *testing.T) {
	mock := &mockResolver{
		errs: map[string]error{
			"my_workspace/missing.txt": fmt.Errorf("file missing"),
		},
	}

	file := runfile.New("my_workspace/missing.txt")

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustResolve() did not panic")
		}
	}()

	file.MustResolve(runfile.WithResolver(mock))
}

func TestDefaultResolver(t *testing.T) {
	mock := &mockResolver{
		paths: map[string]string{
			"my_workspace/data.txt": "/default/path/to/data.txt",
		},
	}

	// Set the default resolver
	runfile.SetDefaultResolver(mock)

	file := runfile.New("my_workspace/data.txt")

	// Resolve without passing WithResolver option
	resolved, err := file.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if resolved.Path() != "/default/path/to/data.txt" {
		t.Errorf("Path() = %q, want %q", resolved.Path(), "/default/path/to/data.txt")
	}
}
