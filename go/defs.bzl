load("@rules_go//go:def.bzl", "go_library")
load("@rules_runfile_codegen_core//internal:rules.bzl", "runfile_codegen")

def go_runfile(name, target = None, targets = None, doc = "", base = None, type = "auto"):
    """Creates a runfile entry configuration for Go code generation.

    This helper function constructs a structured dictionary representing a single runfile
    dependency. It is intended to be passed in the `entries` list of `go_runfile_library`.

    Args:
        name: The Go variable name that will be generated to access this runfile.
            This should follow Go-idiomatic naming conventions (e.g., `ConfigJSON`,
            `TestData`). The generated code will expose this as a public `runfile.FileSpec`
            (or `runfile.ExecutableSpec`) variable.
        target: The Bazel target label of the runfile. Alias for `targets = [target]`.
        targets: A list of Bazel target labels to include in this entry. If multiple
            targets are specified, this entry is automatically treated as a `FileSet`.
        doc: A descriptive comment for the generated Go variable.
        base: An optional path base to resolve relative paths for FileSet files.
            - If set to `""` (empty string), nothing is stripped (keeps full canonical paths).
            - If it starts with `//`, it is resolved relative to the library's repository root.
            - If it starts with `@`, it resolves relative to an external repository.
            - If it starts with `.`, it resolves relative to the library's package path.
            - If set to `"common_dir"`, it automatically computes the longest common prefix.
        type: An optional explicit type assertion for the target.
            - `"auto"` (default): Automatically detects the type (file, directory, or fileset).
            - `"file"`: Asserts the target is a single file.
            - `"directory"`: Asserts the target is a TreeArtifact directory.
            - `"fileset"`: Forces the target to be treated as a FileSet.
            - `"executable"`: Asserts the target is an executable binary.

    Returns:
        A dictionary containing the configured runfile entry.
    """
    if target != None and targets != None:
        fail("Cannot specify both 'target' and 'targets' for runfile '%s'" % name)
    if target == None and targets == None:
        fail("Must specify either 'target' or 'targets' for runfile '%s'" % name)
        
    resolved_targets = targets
    if resolved_targets == None:
        resolved_targets = [target]

    resolved_base = base
    if base != None and (base.startswith("@") or base.startswith("//")):
        resolved_base = str(native.package_relative_label(base))

    return {
        "name": name,
        "targets": resolved_targets,
        "doc": doc,
        "base": resolved_base,
        "type": type,
    }

def go_runfile_library(name, importpath, entries, **kwargs):
    """Generates a Go library containing compile-time safe accessors for runfiles.

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
        fmt.Printf("Config: %s\n", content)

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
    """
    if "srcs" in kwargs:
        fail("Cannot specify 'srcs' in go_runfile_library, they are generated automatically.")

    # 1. Collect and deduplicate targets
    unique_targets = []
    target_to_idx = {}
    for entry in entries:
        for t in entry["targets"]:
            if t not in target_to_idx:
                target_to_idx[t] = len(unique_targets)
                unique_targets.append(t)

    # 2. Build structured config
    config_data = {}
    for entry in entries:
        config_data[entry["name"]] = {
            "target_indexes": [target_to_idx[t] for t in entry["targets"]],
            "doc": entry["doc"],
            "base": entry["base"] if entry.get("base") != None else "__default__",
            "type": entry.get("type", "auto"),
        }

    # Propagate testonly and tags
    testonly = kwargs.get("testonly", None)
    tags = kwargs.get("tags", None)

    # Merge data and deps
    user_data = kwargs.pop("data", [])
    user_deps = kwargs.pop("deps", [])

    # Call the core generator rule
    runfile_codegen(
        name = name + "_codegen",
        package = importpath,
        language = "go",
        targets = unique_targets,
        config = json.encode(config_data),
        testonly = testonly,
        tags = tags,
    )

    # Public Go library
    go_library(
        name = name,
        srcs = [":" + name + "_codegen"],
        importpath = importpath,
        deps = [
            "@rules_runfile_codegen_go//runfile",
        ] + user_deps,
        data = unique_targets + user_data,
        **kwargs
    )
