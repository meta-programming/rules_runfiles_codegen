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

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		rlocation string
		mock      *mockResolver
		opts      []runfile.ResolveOption
		wantPath  string
		wantErr   string
	}{
		{
			name:      "success",
			rlocation: "my_workspace/data.txt",
			mock: &mockResolver{
				paths: map[string]string{
					"my_workspace/data.txt": "/absolute/path/to/data.txt",
				},
			},
			wantPath: "/absolute/path/to/data.txt",
		},
		{
			name:      "resolver error",
			rlocation: "my_workspace/missing.txt",
			mock: &mockResolver{
				errs: map[string]error{
					"my_workspace/missing.txt": fmt.Errorf("permission denied"),
				},
			},
			wantErr: "permission denied",
		},
		{
			name:      "file not found",
			rlocation: "my_workspace/notfound.txt",
			mock:      &mockResolver{},
			wantErr:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := runfile.NewSpec(tt.rlocation)
			opts := append([]runfile.ResolveOption{runfile.WithResolver(tt.mock)}, tt.opts...)
			file, err := spec.Resolve(opts...)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Resolve() succeeded, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Resolve() failed: %v", err)
			}

			if file.Rlocation() != tt.rlocation {
				t.Errorf("Rlocation() = %q, want %q", file.Rlocation(), tt.rlocation)
			}
			if file.Path() != tt.wantPath {
				t.Errorf("Path() = %q, want %q", file.Path(), tt.wantPath)
			}
		})
	}
}

func TestMustResolve(t *testing.T) {
	tests := []struct {
		name      string
		rlocation string
		mock      *mockResolver
		wantPath  string
		wantPanic string
	}{
		{
			name:      "success",
			rlocation: "my_workspace/data.txt",
			mock: &mockResolver{
				paths: map[string]string{
					"my_workspace/data.txt": "/absolute/path/to/data.txt",
				},
			},
			wantPath: "/absolute/path/to/data.txt",
		},
		{
			name:      "panic on error",
			rlocation: "my_workspace/missing.txt",
			mock: &mockResolver{
				errs: map[string]error{
					"my_workspace/missing.txt": fmt.Errorf("file missing"),
				},
			},
			wantPanic: "file missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := runfile.NewSpec(tt.rlocation)

			if tt.wantPanic != "" {
				defer func() {
					r := recover()
					if r == nil {
						t.Fatal("MustResolve() did not panic")
					}
					err, ok := r.(error)
					if !ok {
						t.Fatalf("panic value is not an error: %v", r)
					}
					if !strings.Contains(err.Error(), tt.wantPanic) {
						t.Errorf("panic = %v, want it to contain %q", err, tt.wantPanic)
					}
				}()
			}

			file := spec.MustResolve(runfile.WithResolver(tt.mock))
			if tt.wantPanic == "" {
				if file.Path() != tt.wantPath {
					t.Errorf("Path() = %q, want %q", file.Path(), tt.wantPath)
				}
			}
		})
	}
}

func TestDefaultResolver(t *testing.T) {
	mock := &mockResolver{
		paths: map[string]string{
			"my_workspace/data.txt": "/default/path/to/data.txt",
		},
	}

	// Set the default resolver
	runfile.SetDefaultResolver(mock)

	spec := runfile.NewSpec("my_workspace/data.txt")

	file, err := spec.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	if file.Path() != "/default/path/to/data.txt" {
		t.Errorf("Path() = %q, want %q", file.Path(), "/default/path/to/data.txt")
	}
}
