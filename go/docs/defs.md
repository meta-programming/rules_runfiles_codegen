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
| <a id="go_runfile-type"></a>type |  An optional explicit type assertion for the target. - `"auto"` (default): Automatically detects the type (file, directory, or fileset). - `"file"`: Asserts the target is a single file. - `"directory"`: Asserts the target is a TreeArtifact directory. - `"fileset"`: Forces the target to be treated as a FileSet. - `"executable"`: Asserts the target is an executable binary.   |  `"auto"` |

**RETURNS**

A dictionary containing the configured runfile entry.


<a id="go_runfile_library"></a>

## go_runfile_library

<pre>
load("@rules_runfile_codegen_go//:defs.bzl", "go_runfile_library")

go_runfile_library(<a href="#go_runfile_library-name">name</a>, <a href="#go_runfile_library-importpath">importpath</a>, <a href="#go_runfile_library-entries">entries</a>, <a href="#go_runfile_library-kwargs">**kwargs</a>)
</pre>

Generates a Go library containing compile-time safe accessors for runfiles.

    This macro automates the process of accessing Bazel runfiles in Go. It generates
    a Go source file containing resolved runfile paths, allowing you to access them
    via type-safe variables instead of hardcoded string literals.

    The generated library uses `@rules_go//go/runfiles` to resolve the runfiles at runtime.

    ### Example Usage

    In your `BUILD.bazel`:

    ```bazel
    load("@rules_runfile_codegen_go//:defs.bzl", "go_runfile", "go_runfile_library")

    go_runfile_library(
        name = "my_runfiles",
        importpath = "github.com/example/project/my_runfiles",
        entries = [
            go_runfile(
                name = "ConfigJSON",
                target = "//config:config.json",
                doc = "The main configuration file.",
            ),
            go_runfile(
                name = "HelperTool",
                target = "//tools:helper_tool", # An executable target
                doc = "A helper CLI tool.",
            ),
            go_runfile(
                name = "ExampleSet",
                targets = ["//data:file1.txt", "//data:file2.txt"],
                base = "common_dir",
                doc = "A set of data files.",
            ),
        ],
    )
    ```

    In your `main.go`:

    ```go
    package main

    import (
        "fmt"
        "os"
        "log"

        "github.com/example/project/my_runfiles"
    )

    func main() {
        // 1. Accessing a regular runfile:
        // Resolve the file safely (returns an error if missing).
        configFile, err := my_runfiles.ConfigJSON.Resolve()
        if err != nil {
            log.Fatalf("Failed to resolve config: %v", err)
        }
        content, err := os.ReadFile(configFile.Path())
        if err != nil {
            log.Fatalf("Failed to read config: %v", err)
        }
        fmt.Printf("Config: %s
", content)

        // 2. Running an executable runfile:
        // Use MustResolve() for easy fail-fast access (panics if missing).
        helper := my_runfiles.HelperTool.MustResolve()
        cmd := helper.Cmd("--verbose", "run")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            log.Fatalf("Helper tool failed: %v", err)
        }

        // 3. Accessing a fileset:
        fileset, err := my_runfiles.ExampleSet.Resolve()
        if err != nil {
            log.Fatalf("Failed to resolve fileset: %v", err)
        }
        // Access file inside fileset by its relative path (after base stripping)
        f1, err := fileset.ResolveFile("file1.txt")
        // ... read f1.Path()
    }
    ```

    Args:
        name: A unique name for this target. The generated `go_library` will have this name.
        importpath: The import path for the generated Go library.
        entries: A list of runfile entries, constructed using the `go_runfile` helper.
        **kwargs: Additional arguments to propagate to the underlying `go_library` target.

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="go_runfile_library-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="go_runfile_library-importpath"></a>importpath |  <p align="center"> - </p>   |  none |
| <a id="go_runfile_library-entries"></a>entries |  <p align="center"> - </p>   |  none |
| <a id="go_runfile_library-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


