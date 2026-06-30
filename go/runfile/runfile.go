// Package runfile provides type-safe, lazy-resolving access to Bazel runfiles.
//
// It defines unresolved descriptors (File, Executable) that can be resolved
// to physical paths (ResolvedFile, ResolvedExecutable) at runtime. This
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
		// runfiles.New() returns (*runfiles.Runfiles, error).
		// Since *runfiles.Runfiles implements Resolver, this assignment is valid.
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
// Unresolved Types (Can fail to resolve)
// ---------------------------------------------------------------------------

// File represents an unresolved runfile descriptor.
type File struct {
	rlocation string
}

// New creates a new unresolved File reference.
//
// The rlocation argument must be the runfile's root-relative path (rlocation path),
// which typically follows the format "workspace_name/path/to/file".
//
// For example:
//   - If your main module is named "my_project", a file at "data/config.json"
//     would have the rlocation: "my_project/data/config.json".
//   - If using Bzlmod and the main module name is not explicitly configured,
//     it may default to "_main", e.g., "_main/data/config.json".
func New(rlocation string) File {
	return File{rlocation: rlocation}
}

// Rlocation returns the logical, workspace-relative path of the runfile
// (e.g., "my_project/data/config.json").
func (f File) Rlocation() string {
	return f.rlocation
}

// Resolve attempts to find the runfile on disk.
// It returns a ResolvedFile if successful, or an error if the runfiles
// registry could not be initialized or the file is missing.
//
// You can pass ResolveOptions to customize the resolution behavior (e.g.,
// supplying a custom registry).
func (f File) Resolve(opts ...ResolveOption) (ResolvedFile, error) {
	var o resolveOpts
	for _, opt := range opts {
		opt(&o)
	}

	resolver := o.resolver
	if resolver == nil {
		var err error
		resolver, err = getDefaultResolver()
		if err != nil {
			return ResolvedFile{}, fmt.Errorf("default runfiles resolver initialization failed: %w", err)
		}
	}

	path, err := resolver.Rlocation(f.rlocation)
	if err != nil {
		return ResolvedFile{}, fmt.Errorf("failed to resolve runfile %q: %w", f.rlocation, err)
	}
	return ResolvedFile{rlocation: f.rlocation, absPath: path}, nil
}

// MustResolve is like Resolve but panics if the runfile cannot be found.
// Use this to fail-fast during initialization if the resource is mandatory.
func (f File) MustResolve(opts ...ResolveOption) ResolvedFile {
	rf, err := f.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return rf
}

// Executable represents an unresolved executable runfile descriptor.
type Executable struct {
	File
}

// NewExecutable creates a new unresolved Executable reference.
// See New() for details on the rlocation argument format.
func NewExecutable(rlocation string) Executable {
	return Executable{File: New(rlocation)}
}

// Resolve attempts to find the executable on disk.
func (e Executable) Resolve(opts ...ResolveOption) (ResolvedExecutable, error) {
	rf, err := e.File.Resolve(opts...)
	if err != nil {
		return ResolvedExecutable{}, err
	}
	return ResolvedExecutable{ResolvedFile: rf}, nil
}

// MustResolve is like Resolve but panics if the executable cannot be found.
func (e Executable) MustResolve(opts ...ResolveOption) ResolvedExecutable {
	re, err := e.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return re
}

// ---------------------------------------------------------------------------
// Resolved Types (Cannot fail)
// ---------------------------------------------------------------------------

// ResolvedFile represents a runfile that has been successfully located on disk.
// Its methods are guaranteed not to fail.
type ResolvedFile struct {
	rlocation string
	absPath   string
}

// Rlocation returns the logical, workspace-relative path of the runfile
// (e.g., "my_project/data/config.json").
func (rf ResolvedFile) Rlocation() string {
	return rf.rlocation
}

// Path returns the physical, absolute path to the runfile on the local filesystem
// (e.g., "/path/to/bazel-out/.../inputs/data/config.json").
//
// This is guaranteed to be a valid, non-empty path.
func (rf ResolvedFile) Path() string {
	return rf.absPath
}

// ResolvedExecutable represents an executable runfile successfully located on disk.
type ResolvedExecutable struct {
	ResolvedFile
}

// Cmd returns an *exec.Cmd pre-configured to run this executable,
// with Bazel runfiles environment variables already propagated.
//
// This method is guaranteed to succeed and does not return an error.
func (re ResolvedExecutable) Cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(re.Path(), args...)
	if env, err := runfiles.Env(); err == nil {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd
}

