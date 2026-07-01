<!-- Generated with Stardoc: http://skydoc.bazel.build -->



<a id="go_runfile"></a>

## go_runfile

<pre>
load("@rules_runfile_codegen_go//:defs.bzl", "go_runfile")

go_runfile(<a href="#go_runfile-name">name</a>, <a href="#go_runfile-target">target</a>, <a href="#go_runfile-targets">targets</a>, <a href="#go_runfile-doc">doc</a>, <a href="#go_runfile-base">base</a>, <a href="#go_runfile-type">type</a>)
</pre>

Creates a runfile entry configuration for Go code generation.

This helper function constructs a structured dictionary representing a single runfile
dependency. It is intended to be passed in the `entries` list of `go_runfile_library`.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="go_runfile-name"></a>name |  The Go variable name that will be generated to access this runfile. This should follow Go-idiomatic naming conventions (e.g., `ConfigJSON`, `TestData`). The generated code will expose this as a public `runfile.FileSpec` (or `runfile.ExecutableSpec`) variable.   |  none |
| <a id="go_runfile-target"></a>target |  The Bazel target label of the runfile. Alias for `targets = [target]`.   |  `None` |
| <a id="go_runfile-targets"></a>targets |  A list of Bazel target labels to include in this entry. If multiple targets are specified, this entry is automatically treated as a `FileSet`.   |  `None` |
| <a id="go_runfile-doc"></a>doc |  A descriptive comment for the generated Go variable.   |  `""` |
| <a id="go_runfile-base"></a>base |  An optional path base to resolve relative paths for FileSet files. - If set to `""` (empty string), nothing is stripped (keeps full canonical paths). - If it starts with `//`, it is resolved relative to the library's repository root. - If it starts with `@`, it resolves relative to an external repository. - If it starts with `.`, it resolves relative to the library's package path. - If set to `"common_dir"`, it automatically computes the longest common prefix.   |  `None` |
| <a id="go_runfile-type"></a>type |  An optional explicit type assertion for the target. - `"auto"` (default): Automatically detects the type (file, directory, or fileset). - `"file"`: Asserts the target is a single file. - `"directory"`: Asserts the target is a TreeArtifact directory. - `"fileset"`: Forces the target to be treated as a FileSet.   |  `"auto"` |

**RETURNS**

A dictionary containing the configured runfile entry.


<a id="go_runfile_library"></a>

## go_runfile_library

<pre>
load("@rules_runfile_codegen_go//:defs.bzl", "go_runfile_library")

go_runfile_library(<a href="#go_runfile_library-name">name</a>, <a href="#go_runfile_library-importpath">importpath</a>, <a href="#go_runfile_library-entries">entries</a>, <a href="#go_runfile_library-kwargs">**kwargs</a>)
</pre>

Generates a Go library containing compile-time safe accessors for runfiles.

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="go_runfile_library-name"></a>name |  A unique name for this target. The generated `go_library` will have this name.   |  none |
| <a id="go_runfile_library-importpath"></a>importpath |  The import path for the generated Go library.   |  none |
| <a id="go_runfile_library-entries"></a>entries |  A list of runfile entries, constructed using the `go_runfile` helper.   |  none |
| <a id="go_runfile_library-kwargs"></a>kwargs |  Additional arguments to propagate to the underlying `go_library` target.   |  none |


