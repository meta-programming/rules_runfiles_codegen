// Package runfile provides type-safe access to Bazel runfiles.
//
// It defines unresolved specifications (FileSpec, ExecutableSpec) that
// can be resolved to physical files (File, Executable) at runtime. This
// separates the fallible resolution step from the infallible path usage,
// avoiding unexpected panics during program initialization.
package runfile

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// Resolver defines the interface for looking up runfiles.
// It is satisfied by the concrete *runfiles.Runfiles struct from rules_go.
type Resolver interface {
	Rlocation(path string) (string, error)
}

var (
	defaultResolver     Resolver
	defaultResolverOnce sync.Once
	defaultResolverErr  error
)

func getDefaultResolver() (Resolver, error) {
	defaultResolverOnce.Do(func() {
		var reg *runfiles.Runfiles
		reg, defaultResolverErr = runfiles.New()
		defaultResolver = reg
	})
	return defaultResolver, defaultResolverErr
}

// SetDefaultResolver overrides the default runfiles resolver (useful for unit tests).
func SetDefaultResolver(resolver Resolver) {
	defaultResolverOnce.Do(func() {
		defaultResolver = resolver
	})
}

// ---------------------------------------------------------------------------
// Resolve Options
// ---------------------------------------------------------------------------

type resolveOpts struct {
	resolver Resolver
}

// ResolveOption configures how a runfile is resolved.
type ResolveOption func(*resolveOpts)

// WithResolver overrides the runfiles resolver used for resolution.
// Use this to inject a mock resolver for testing or to use a custom-configured
// resolver instead of the auto-detected default one.
func WithResolver(resolver Resolver) ResolveOption {
	return func(o *resolveOpts) {
		o.resolver = resolver
	}
}

// ---------------------------------------------------------------------------
// Unresolved Specifications (Resolution can fail)
// ---------------------------------------------------------------------------

// FileSpec represents an unresolved runfile specification.
// It holds the logical rlocation path but has not yet been located on disk.
type FileSpec struct {
	rlocation string
}

// NewSpec creates a new unresolved FileSpec reference.
//
// The rlocationpath argument must be a runfiles-root-relative path (rlocation path).
//
// Depending on how this path is constructed, it can be:
//   - **Apparent**: Starts with an apparent repository name (e.g., "rules_go/path/to/file").
//     This is common when hand-writing paths in code. It is context-dependent and
//     resolved at runtime using the caller's repository mapping.
//   - **Canonical**: Starts with a canonical repository name (e.g., "rules_go~~go~0.39.0/path/to/file"
//     or "_main/path/to/file"). This is typically returned by the `$(rlocationpath ...)`
//     helper in BUILD files. It is globally unique within the runfiles tree and
//     does not require runtime mapping.
//
// For a detailed explanation of these concepts, see RUNFILES_CONCEPTS.md in this
// repository, or refer to the official Bazel documentation:
//   - https://bazel.build/extending/rules#runfiles_location
//   - https://bazel.build/external/overview#apparent-repo-name
func NewSpec(rlocationpath string) FileSpec {
	return FileSpec{rlocation: rlocationpath}
}

// RlocationPath returns the logical, workspace-relative path of the runfile
// (e.g., "my_project/data/config.json").
//
// This is the key used to look up the runfile in the runfiles manifest or directory.
// For details on how Bazel structures these paths, see the Bazel Runfiles guide:
// https://bazel.build/extending/rules#runfiles
func (fs FileSpec) RlocationPath() string {
	return fs.rlocation
}

// Resolve attempts to find the runfile on disk.
// It returns a File if successful, or an error if the runfiles
// resolver could not be initialized or the file is missing.
//
// You can pass ResolveOptions to customize the resolution behavior (e.g.,
// supplying a custom resolver).
func (fs FileSpec) Resolve(opts ...ResolveOption) (File, error) {
	var o resolveOpts
	for _, opt := range opts {
		opt(&o)
	}

	resolver := o.resolver
	if resolver == nil {
		var err error
		resolver, err = getDefaultResolver()
		if err != nil {
			return File{}, fmt.Errorf("default runfiles resolver initialization failed: %w", err)
		}
	}

	path, err := resolver.Rlocation(fs.rlocation)
	if err != nil {
		return File{}, fmt.Errorf("failed to resolve runfile %q: %w", fs.rlocation, err)
	}
	return File{rlocation: fs.rlocation, absPath: path}, nil
}

// MustResolve is like Resolve but panics if the runfile cannot be found.
// Use this to fail-fast during initialization if the resource is mandatory.
func (fs FileSpec) MustResolve(opts ...ResolveOption) File {
	f, err := fs.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return f
}

// ExecutableSpec represents an unresolved executable runfile specification.
type ExecutableSpec struct {
	FileSpec
}

// NewExecutableSpec creates a new unresolved ExecutableSpec reference.
// See NewSpec() for details on the rlocation argument format.
func NewExecutableSpec(rlocation string) ExecutableSpec {
	return ExecutableSpec{FileSpec: NewSpec(rlocation)}
}

// Resolve attempts to find the executable on disk.
func (es ExecutableSpec) Resolve(opts ...ResolveOption) (Executable, error) {
	f, err := es.FileSpec.Resolve(opts...)
	if err != nil {
		return Executable{}, err
	}
	return Executable{File: f}, nil
}

// MustResolve is like Resolve but panics if the executable cannot be found.
func (es ExecutableSpec) MustResolve(opts ...ResolveOption) Executable {
	e, err := es.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return e
}

// ---------------------------------------------------------------------------
// Resolved Types (Guaranteed to exist, cannot fail)
// ---------------------------------------------------------------------------

// File represents a runfile that has been successfully located on disk.
// Its methods are guaranteed not to fail.
type File struct {
	rlocation string
	absPath   string
}

// RlocationPath returns the logical, workspace-relative path of the runfile
// (e.g., "my_project/data/config.json").
//
// This is the key used to look up the runfile in the runfiles manifest or directory.
// For details on how Bazel structures these paths, see the Bazel Runfiles guide:
// https://bazel.build/extending/rules#runfiles
func (f File) RlocationPath() string {
	return f.rlocation
}

// Path returns the physical, absolute path to the runfile on the local filesystem
// (e.g., "/path/to/bazel-out/.../inputs/data/config.json").
//
// This is guaranteed to be a valid, non-empty path.
func (f File) Path() string {
	return f.absPath
}

// Executable represents an executable runfile successfully located on disk.
type Executable struct {
	File
}

// Cmd returns an *exec.Cmd pre-configured to run this executable,
// with Bazel runfiles environment variables already propagated.
//
// This method is guaranteed to succeed and does not return an error.
func (e Executable) Cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(e.Path(), args...)
	if env, err := runfiles.Env(); err == nil {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd
}
