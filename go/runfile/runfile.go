// Package runfile provides type-safe access to Bazel runfiles.
//
// It defines unresolved specifications ([FileSpec], [ExecutableSpec]) that
// can be resolved to physical files ([File], [Executable]) at runtime. This
// separates the fallible resolution step from the infallible path usage,
// avoiding unexpected panics during program initialization.
package runfile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// RlocationPath represents a logical, runfiles-root-relative path used to
// locate a data dependency (runfile) at runtime.
//
// This path serves as the key for lookups in the runfiles manifest or symlink tree.
// Depending on how it is constructed, an RlocationPath can be in one of two formats:
//
//   - **Apparent Path**: Starts with an apparent repository name (e.g., "rules_go/path/to/file").
//     This is the "nickname" of the repository as seen by the caller. It is common
//     in hand-written code, is context-dependent, and is resolved at runtime using
//     the calling repository's [apparent repository mapping](https://bazel.build/external/overview#apparent-repo-name).
//
//   - **Canonical Path**: Starts with a canonical repository name (e.g., "rules_go~~go~0.39.0/path/to/file"
//     or "_main/path/to/file"). This is globally unique within the runfiles tree and
//     does not require runtime mapping. This format is typically returned by the
//     [$(rlocationpath ...)](https://bazel.build/extending/rules#runfiles) helper in BUILD files.
//
// ### The Special "_main" Repository
// Under Bzlmod, the main repository (your project) is always assigned the fixed
// canonical name "_main" in the runfiles directory, decoupling it from any legacy
// workspace names defined in WORKSPACE files. Consequently, canonical paths to
// your own project's runfiles will start with "_main/" (e.g., "_main/path/to/file").
// See the [Bazel Bzlmod Migration Guide](https://bazel.build/external/migration) for details.
//
// For a detailed explanation of these concepts, see RUNFILES_CONCEPTS.md in this
// repository, or refer to the official Bazel documentation:
//   - [Bazel Runfiles Location](https://bazel.build/extending/rules#runfiles_location)
//   - [Bazel Apparent Repository Name](https://bazel.build/external/overview#apparent-repo-name)
//
// (Note: The spelling "RlocationPath" uses a lowercase "l" to treat "rlocation" as a single
// coined word—similar to "email"—and to maintain consistency with the underlying
// rules_go package's Rlocation method.)
type RlocationPath string

// String returns the path as a plain string.
func (p RlocationPath) String() string {
	return string(p)
}

// Resolver defines the interface for looking up runfiles.
// It is satisfied by the concrete [*runfiles.Runfiles] struct from rules_go.
//
// Note: This interface uses plain string for the path to remain compatible
// with [*runfiles.Runfiles.Rlocation] without requiring a wrapper. However,
// the argument is conceptually an [RlocationPath].
//   - When calling this method, convert an [RlocationPath] to string using string(path).
//   - When implementing this interface (e.g., in mocks), you can convert the incoming
//     string to [RlocationPath] using RlocationPath(path) for stronger type safety.
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
	rlocation RlocationPath
}

// NewSpec creates a new unresolved [FileSpec] reference.
//
// See [RlocationPath] for details on the expected path formats.
func NewSpec(rlocationpath RlocationPath) FileSpec {
	return FileSpec{rlocation: rlocationpath}
}

// RlocationPath returns the logical, runfiles-root-relative path of the runfile.
//
// See [RlocationPath] for details on the path formats.
func (fs FileSpec) RlocationPath() RlocationPath {
	return fs.rlocation
}

// Resolve attempts to find the runfile on disk.
// It returns a [File] if successful, or an error if the runfiles
// resolver could not be initialized or the file is missing.
//
// You can pass [ResolveOption] to customize the resolution behavior (e.g.,
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

	path, err := resolver.Rlocation(string(fs.rlocation))
	if err != nil {
		return File{}, fmt.Errorf("failed to resolve runfile %q: %w", fs.rlocation, err)
	}
	return File{rlocation: fs.rlocation, absPath: path}, nil
}

// MustResolve is like [FileSpec.Resolve] but panics if the runfile cannot be found.
// Use this to fail-fast during initialization if the resource is mandatory.
func (fs FileSpec) MustResolve(opts ...ResolveOption) File {
	f, err := fs.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return f
}

// ExecutableSpec represents an unresolved executable runfile specification.
//
// This typically points to an executable target (like a go_binary, sh_binary,
// or cc_binary).
type ExecutableSpec struct {
	FileSpec
}

// NewExecutableSpec creates a new unresolved [ExecutableSpec] reference.
//
// See [RlocationPath] for details on the expected path formats.
func NewExecutableSpec(rlocation RlocationPath) ExecutableSpec {
	return ExecutableSpec{FileSpec: NewSpec(rlocation)}
}

// Resolve attempts to find the executable on disk.
// It returns an [Executable] if successful.
func (es ExecutableSpec) Resolve(opts ...ResolveOption) (Executable, error) {
	f, err := es.FileSpec.Resolve(opts...)
	if err != nil {
		return Executable{}, err
	}
	return Executable{File: f}, nil
}

// MustResolve is like [ExecutableSpec.Resolve] but panics if the executable cannot be found.
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

// File represents a runfile that has been successfully located on disk (via [FileSpec.Resolve]).
// Its methods are guaranteed not to fail.
type File struct {
	rlocation RlocationPath
	absPath   string
}

// RlocationPath returns the logical, runfiles-root-relative path of the runfile.
//
// See [RlocationPath] for details on the path formats.
func (f File) RlocationPath() RlocationPath {
	return f.rlocation
}

// Path returns the physical, absolute path to the runfile on the local filesystem
// (e.g., "/path/to/bazel-out/.../inputs/data/config.json").
//
// This is guaranteed to be a valid, non-empty path.
func (f File) Path() string {
	return f.absPath
}

// Executable represents an executable runfile successfully located on disk (via [ExecutableSpec.Resolve]).
type Executable struct {
	File
}

// Cmd returns an [exec.Cmd] pre-configured to run this executable.
//
// It automatically propagates the environment variables returned by
// [runfiles.Env] (such as RUNFILES_DIR and RUNFILES_MANIFEST_FILE) to the
// subprocess to ensure it can also locate its runfiles.
//
// This method is guaranteed to succeed and does not return an error.
func (e Executable) Cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(e.Path(), args...)
	if env, err := runfiles.Env(); err == nil {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd
}

// ---------------------------------------------------------------------------
// Directory (TreeArtifact) Support
// ---------------------------------------------------------------------------

// DirectorySpec represents an unresolved directory runfile (TreeArtifact).
//
// See https://bazel.build/extending/rules#tree_artifacts for details.
type DirectorySpec struct {
	FileSpec
}

// NewDirectorySpec creates a new unresolved [DirectorySpec] reference.
func NewDirectorySpec(rlocation RlocationPath) DirectorySpec {
	return DirectorySpec{FileSpec: NewSpec(rlocation)}
}

// Resolve attempts to find the directory on disk.
func (ds DirectorySpec) Resolve(opts ...ResolveOption) (Directory, error) {
	f, err := ds.FileSpec.Resolve(opts...)
	if err != nil {
		return Directory{}, err
	}
	return Directory{File: f}, nil
}

// MustResolve is like [DirectorySpec.Resolve] but panics if the directory cannot be found.
func (ds DirectorySpec) MustResolve(opts ...ResolveOption) Directory {
	d, err := ds.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return d
}

// Directory represents a resolved directory runfile (TreeArtifact).
//
// See https://bazel.build/extending/rules#tree_artifacts for details.
type Directory struct {
	File
}

// Child returns a File reference to a file inside this directory.
// Note: This does NOT resolve the file via the runfiles resolver,
// but simply joins the directory path with the relative path.
func (d Directory) Child(relPath string) File {
	return File{
		rlocation: d.rlocation + "/" + RlocationPath(relPath),
		absPath:   filepath.Join(d.Path(), relPath),
	}
}

// ---------------------------------------------------------------------------
// FileSet Support
// ---------------------------------------------------------------------------

// FileSetSpec represents an unresolved fileset of runfiles (typically representing
// multiple targets or a multi-file target like filegroup).
//
// See https://bazel.build/reference/be/general#filegroup for details.
type FileSetSpec struct {
	files map[string]string // maps user-facing relPath to canonical rlocation path
}

// NewFileSetSpec creates a new unresolved [FileSetSpec] reference.
func NewFileSetSpec(files map[string]string) FileSetSpec {
	return FileSetSpec{files: files}
}

// Resolve attempts to prepare the fileset for resolution.
func (fss FileSetSpec) Resolve(opts ...ResolveOption) (FileSet, error) {
	var o resolveOpts
	for _, opt := range opts {
		opt(&o)
	}
	resolver := o.resolver
	if resolver == nil {
		var err error
		resolver, err = getDefaultResolver()
		if err != nil {
			return FileSet{}, fmt.Errorf("default runfiles resolver initialization failed: %w", err)
		}
	}

	resolvedFiles := make(map[string]File, len(fss.files))
	for rel, rloc := range fss.files {
		absPath, err := resolver.Rlocation(rloc)
		if err != nil {
			return FileSet{}, fmt.Errorf("failed to resolve file %q in fileset: %w", rloc, err)
		}
		resolvedFiles[rel] = File{rlocation: RlocationPath(rloc), absPath: absPath}
	}

	return FileSet{files: resolvedFiles}, nil
}

// MustResolve is like [FileSetSpec.Resolve] but panics if the resolver cannot be initialized.
func (fss FileSetSpec) MustResolve(opts ...ResolveOption) FileSet {
	fs, err := fss.Resolve(opts...)
	if err != nil {
		panic(err)
	}
	return fs
}

// FileSet represents a resolved fileset of runfiles (typically representing
// multiple targets or a multi-file target like filegroup).
//
// See https://bazel.build/reference/be/general#filegroup for details.
type FileSet struct {
	files map[string]File // maps user-facing relPath to resolved File
}

// RelPaths returns the list of relative paths in this fileset.
func (fs FileSet) RelPaths() []string {
	paths := make([]string, 0, len(fs.files))
	for p := range fs.files {
		paths = append(paths, p)
	}
	return paths
}

// File returns a specific file in the fileset by its relative path.
func (fs FileSet) File(relPath string) (File, error) {
	file, ok := fs.files[relPath]
	if !ok {
		return File{}, fmt.Errorf("file %q is not in this fileset", relPath)
	}
	return file, nil
}

// MustFile is like [FileSet.File] but panics if the file is not found.
func (fs FileSet) MustFile(relPath string) File {
	file, err := fs.File(relPath)
	if err != nil {
		panic(err)
	}
	return file
}
