# Bazel Runfile Code Generators (`rules_runfile_codegen`)

In Bazel, [**runfiles**](https://bazel.build/extending/rules#runfiles) are the files (data dependencies, configuration files, or other executables) that a binary needs at runtime. Accessing these files programmatically requires resolving their paths relative to the workspace, which can be complex and error-prone because their physical locations change depending on the execution environment (e.g., running locally, inside a sandbox during `bazel test`, or in a production deployment).

Traditionally, developers must use Bazel's language-specific runfiles libraries to perform string-based lookups at runtime. This project, `rules_runfile_codegen`, simplifies this by generating type-safe code accessors for your runfiles.

By defining your runfile dependencies in your `BUILD.bazel` files, you can generate libraries that expose these runfiles as strongly-typed symbols. This eliminates the need to hardcode runfile paths as strings in your application code, prevents typos, and ensures that runfile resolution errors are caught at startup rather than deep in runtime execution.

For complete, runnable projects demonstrating these quickstarts, see the [examples/](examples) directory.

---

## Key Features

*   **Type-Safety**: Runfiles are exposed as generated constants/properties. No more stringly-typed paths.
*   **Explicit Resolution**: Go, Kotlin, and Rust use an explicit, non-eager (lazy) resolution model to avoid startup side-effects and improve testability. The generated libraries expose unresolved specifications (e.g. `FileSpec`) that must be explicitly resolved at runtime.
*   **Subprocess Environment Propagation**: Executable runfiles are wrapped in rich objects that facilitate launching them as subprocesses while automatically propagating the Bazel runfiles environment. This ensures that child processes can also resolve their own runfiles.[^1]
*   **Zero Runtime Overhead**: After successful startup-time resolution, accessing the runfile path is a simple member access with zero overhead.

---

## Go Quickstart

### Bzlmod Setup

To use these rules in your Go project, add the following to your `MODULE.bazel` file (see [rules_runfile_codegen_go on BCR](https://registry.bazel.build/modules/rules_runfile_codegen_go)):

<!-- GO_INSTALL_START -->
```bazel
# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_go", version = "0.1.2")
```
<!-- GO_INSTALL_END -->

### 1. Define the Runfile Library

In your `BUILD.bazel`, load the Go rules and define a `go_runfile_library`.

<!-- GO_BUILD_START -->
```bazel
load("@rules_go//go:def.bzl", "go_binary")
load("@rules_runfile_codegen_go//:defs.bzl", "go_runfile", "go_runfile_library")

package(default_visibility = ["//visibility:private"])

# A helper binary to demonstrate executable runfiles
go_binary(
    name = "helper",
    srcs = ["helper.go"],
)

# Generate the runfile accessor library
go_runfile_library(
    name = "resources",
    importpath = "github.com/example/project/examples/go/resources",
    entries = [
        go_runfile(
            name = "DataFile",
            target = "data/dummy.txt",
            doc = "A dummy data file.",
        ),
        go_runfile(
            name = "HelperTool",
            target = ":helper",
            doc = "A helper tool executable.",
        ),
        go_runfile(
            name = "ExampleSet",
            targets = ["data/dummy.txt", "data/info.txt"],
            base = "common_dir",
            doc = "A set of example data files.",
        ),
    ],
)

# Use the library in a binary
go_binary(
    name = "main",
    srcs = ["main.go"],
    deps = [
        ":resources",
    ],
)
# #
```
<!-- GO_BUILD_END -->

### 2. Use the Generated Symbols in Go

Import the generated package and access the symbols. Regular files are generated as `Runfile` types, and executables are generated as `ExecutableRunfile` types.

Here is the actual example:

<!-- GO_USAGE_START -->
```go
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/example/project/examples/go/resources"
)

func main() {
	// 1. Access the resolved runfile path safely.
	dataFile, err := resources.DataFile.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving runfile: %v\n", err)
		os.Exit(1)
	}

	content, err := os.ReadFile(dataFile.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading runfile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Data: %s\n", string(content))

	// 2. Run an executable runfile with env propagation (fail-fast).
	helper := resources.HelperTool.MustResolve()
	cmd := helper.Cmd()
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running helper: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Helper output: %s", string(output))

	// 3. Access a fileset of runfiles (FileSet).
	exampleSet, err := resources.ExampleSet.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving fileset: %v\n", err)
		os.Exit(1)
	}
	paths := exampleSet.RelPaths()
	sort.Strings(paths)
	fmt.Printf("FileSet paths: %v\n", paths)
	f1, err := exampleSet.File("dummy.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving dummy.txt: %v\n", err)
		os.Exit(1)
	}
	c1, _ := os.ReadFile(f1.Path())
	fmt.Printf("FileSet dummy content: %s\n", strings.TrimSpace(string(c1)))
}
```
<!-- GO_USAGE_END -->

### Actual Generated Go Code

<details>
<summary>Click to view the actual generated Go code for the example above</summary>

<!-- GENERATED_GO_START -->
```go
// Code generated by rules_runfile_codegen. DO NOT EDIT.

// Package resources provides type-safe access to Bazel runfiles.
package resources

import (
	"github.com/meta-programming/rules_runfiles_codegen/go/runfile"
)

var (
	// DataFile is A dummy data file.
	// Source: @@//:data/dummy.txt
	DataFile = runfile.NewSpec("_main/data/dummy.txt")

	// ExampleSet is A set of example data files.
	// Source: @@//:data/dummy.txt, @@//:data/info.txt
	ExampleSet = runfile.NewFileSetSpec(map[string]string{"dummy.txt": "_main/data/dummy.txt", "info.txt": "_main/data/info.txt"})

	// HelperTool is A helper tool executable.
	// Source: @@//:helper
	HelperTool = runfile.NewExecutableSpec("_main/helper_/helper")

)
```
<!-- GENERATED_GO_END -->
</details>

---

## Kotlin Quickstart

### Bzlmod Setup

To use these rules in your Kotlin project, add the following to your `MODULE.bazel` file (see [rules_runfile_codegen_kotlin on BCR](https://registry.bazel.build/modules/rules_runfile_codegen_kotlin)):

<!-- KOTLIN_INSTALL_START -->
```bazel
# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_kotlin", version = "0.1.2")
```
<!-- KOTLIN_INSTALL_END -->

### 1. Define the Runfile Library

In your `BUILD.bazel`, load the Kotlin rules and define a `kt_jvm_runfile_library`. Note that dashed target names (like `test-resources`) are automatically sanitized to PascalCase Kotlin object names (`TestResources`).

<!-- KOTLIN_BUILD_START -->
```bazel
load("@rules_kotlin//kotlin:jvm.bzl", "kt_jvm_binary")
load("@rules_runfile_codegen_kotlin//:defs.bzl", "kt_runfile", "kt_jvm_runfile_library")

package(default_visibility = ["//visibility:private"])

# A helper binary to demonstrate executable runfiles
kt_jvm_binary(
    name = "helper",
    srcs = ["Helper.kt"],
    main_class = "com.example.project.examples.HelperKt",
)

# Generate the runfile accessor library
kt_jvm_runfile_library(
    name = "resources",
    package = "com.example.project.examples.resources",
    entries = [
        kt_runfile(
            name = "configJson",
            target = "data/dummy.txt",
            doc = "A dummy data file.",
        ),
        kt_runfile(
            name = "helperTool",
            target = ":helper",
            doc = "A helper tool executable.",
        ),
        kt_runfile(
            name = "exampleSet",
            targets = ["data/dummy.txt", "data/info.txt"],
            base = "common_dir",
            doc = "A set of example data files.",
        ),
    ],
)

# Use the library in a binary
kt_jvm_binary(
    name = "main",
    srcs = ["Main.kt"],
    main_class = "com.example.project.examples.MainKt",
    deps = [
        ":resources",
    ],
)
# #
```
<!-- KOTLIN_BUILD_END -->

### 2. Use the Generated Symbols in Kotlin

Import the generated object and access the properties. Regular files are generated as `Runfile` types (exposing `path` and `jvmPath`), and executables are generated as `ExecutableRunfile` types (adding `processBuilder()`).

Here is the actual example:

<!-- KOTLIN_USAGE_START -->
```kotlin
package com.example.project.examples

import com.example.project.examples.resources.Resources
import kotlin.io.path.readText

fun main() {
    // 1. Access the resolved runfile path.
    // Resolve the spec and read its content directly using Path.readText().
    val content = Resources.configJson.path.readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    val process = Resources.helperTool.processBuilder().start()
    val output = process.inputStream.reader().use { it.readText() }.trim()
    val exitCode = process.waitFor()
    if (exitCode != 0) {
        error("Helper tool failed with exit code $exitCode")
    }
    println("Helper output: $output")

    // 3. Access a fileset of runfiles (FileSet).
    val exampleSet = Resources.exampleSet.resolve()
    println("FileSet paths: ${exampleSet.relPaths.sorted()}")
    
    // Access a path inside the fileset using the shortcut:
    val dummyContent = Resources.exampleSet["dummy.txt"].path.readText().trim()
    println("FileSet dummy content: $dummyContent")
}
```
<!-- KOTLIN_USAGE_END -->

### Coroutines & Blocking I/O

The `resolve()` method and subsequent file/process operations (like `readText()` or `process.waitFor()`) are **blocking I/O operations**. If you are using this library inside a coroutine-based application (such as Ktor, Spring WebFlux, or Android), you should offload these calls to `Dispatchers.IO` to avoid blocking event loops or the main thread:

```kotlin
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlin.io.path.readText

suspend fun loadConfig(): String = withContext(Dispatchers.IO) {
    Resources.configJson.resolve().path.readText()
}
```

### Actual Generated Kotlin Code

<details>
<summary>Click to view the actual generated Kotlin code for the example above</summary>

<!-- GENERATED_KOTLIN_START -->
```kotlin
// Code generated by rules_runfile_codegen. DO NOT EDIT.
// This file provides type-safe access to Bazel runfiles.
package com.example.project.examples.resources

import com.github.metaprogramming.runfiles.FileSpec
import com.github.metaprogramming.runfiles.ExecutableSpec
import com.github.metaprogramming.runfiles.DirectorySpec
import com.github.metaprogramming.runfiles.FileSetSpec
import com.github.metaprogramming.runfiles.RlocationPath

object Resources {
    /**
     * A dummy data file.
     * Source: @@//:data/dummy.txt
     */
    val configJson = FileSpec(RlocationPath("_main/data/dummy.txt"))

    /**
     * A set of example data files.
     * Source: @@//:data/dummy.txt, @@//:data/info.txt
     */
    val exampleSet = FileSetSpec(mapOf("dummy.txt" to "_main/data/dummy.txt", "info.txt" to "_main/data/info.txt"))

    /**
     * A helper tool executable.
     * Source: @@//:helper
     */
    val helperTool = ExecutableSpec(RlocationPath("_main/helper"))
}
```
<!-- GENERATED_KOTLIN_END -->
</details>

---

## Rust Quickstart

### Bzlmod Setup

To use these rules in your Rust project, add the following to your `MODULE.bazel` file. The module needs [rules_rust](https://github.com/bazelbuild/rules_rust) with a registered Rust toolchain.

<!-- RUST_INSTALL_START -->
```bazel
# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_rust", version = "0.1.2")
```
<!-- RUST_INSTALL_END -->

### 1. Define the Runfile Library

Define a `rust_runfile_library` in your `BUILD.bazel` file. It generates a crate, named after the target, with one `pub static` per entry. Static names are upper snake case.

<!-- RUST_BUILD_START -->
```bazel
load("@rules_runfile_codegen_rust//:defs.bzl", "rust_runfile", "rust_runfile_library")
load("@rules_rust//rust:defs.bzl", "rust_binary")

package(default_visibility = ["//visibility:private"])

# A helper binary to demonstrate executable runfiles
rust_binary(
    name = "helper",
    srcs = ["helper.rs"],
)

# Generate the runfile accessor crate
rust_runfile_library(
    name = "resources",
    entries = [
        rust_runfile(
            name = "DATA_FILE",
            target = "data/dummy.txt",
            doc = "A dummy data file.",
        ),
        rust_runfile(
            name = "HELPER_TOOL",
            target = ":helper",
            doc = "A helper tool executable.",
        ),
        rust_runfile(
            name = "EXAMPLE_SET",
            targets = ["data/dummy.txt", "data/info.txt"],
            base = "common_dir",
            doc = "A set of example data files.",
        ),
    ],
)

# Use the crate in a binary
rust_binary(
    name = "main",
    srcs = ["main.rs"],
    deps = [
        ":resources",
    ],
)
# #
```
<!-- RUST_BUILD_END -->

### 2. Use the Generated Symbols in Rust

Each static is an unresolved spec (`FileSpec`, `ExecutableSpec`, `DirectorySpec`, or `FileSetSpec`) from the `runfile` crate. Call `resolve()` to get a `Result`, or `must_resolve()` to panic if the runfile is missing. `command()` on a resolved executable returns a `std::process::Command` that passes `RUNFILES_DIR` or `RUNFILES_MANIFEST_FILE` to the child. Lookups use the [`runfiles` crate from rules_rust](https://github.com/bazelbuild/rules_rust/tree/main/rust/runfiles), so they work with a runfiles directory or with a manifest only.

<!-- RUST_USAGE_START -->
```rust
use std::process::ExitCode;

fn main() -> ExitCode {
    // 1. Resolve a runfile, handling the error.
    let data_file = match resources::DATA_FILE.resolve() {
        Ok(file) => file,
        Err(err) => {
            eprintln!("Error resolving runfile: {err}");
            return ExitCode::FAILURE;
        }
    };
    let content = std::fs::read_to_string(data_file.path()).expect("read data file");
    println!("Data: {content}");

    // 2. Run an executable runfile with the runfiles environment (fail-fast).
    let helper = resources::HELPER_TOOL.must_resolve();
    let output = helper.command().output().expect("run helper");
    print!("Helper output: {}", String::from_utf8_lossy(&output.stdout));

    // 3. Access a set of runfiles by relative path.
    let example_set = resources::EXAMPLE_SET.must_resolve();
    let paths: Vec<_> = example_set.rel_paths().collect();
    println!("FileSet paths: {paths:?}");
    let dummy = example_set.must_file("dummy.txt");
    let dummy_content = std::fs::read_to_string(dummy.path()).expect("read dummy.txt");
    println!("FileSet dummy content: {}", dummy_content.trim());
    ExitCode::SUCCESS
}
```
<!-- RUST_USAGE_END -->

In unit tests, pass your own `runfile::Resolver` to `resolve_with()` instead of `resolve()`.

The `runfile` crate reads the file system, so it does not build for `wasm32-unknown-unknown`. Use it in binaries and tests, not in browser code.

### Actual Generated Rust Code

<details>
<summary>Click to view the actual generated Rust code for the example above</summary>
<!-- GENERATED_RUST_START -->
```rust
// Code generated by rules_runfile_codegen. DO NOT EDIT.

//! Typed accessors for the Bazel runfiles of `//:resources`.
//!
//! Each static is an unresolved spec from the `runfile` crate. Call
//! `resolve()` to find the runfile on disk, or `must_resolve()` to panic if
//! it is missing.

/// A dummy data file.
///
/// Source: `//:data/dummy.txt`
pub static DATA_FILE: runfile::FileSpec = runfile::FileSpec::new("_main/data/dummy.txt");

/// A set of example data files.
///
/// Source: `//:data/dummy.txt, //:data/info.txt`
pub static EXAMPLE_SET: runfile::FileSetSpec = runfile::file_set! {
    "dummy.txt" => "_main/data/dummy.txt",
    "info.txt" => "_main/data/info.txt",
};

/// A helper tool executable.
///
/// Source: `//:helper`
pub static HELPER_TOOL: runfile::ExecutableSpec = runfile::ExecutableSpec::new("_main/helper");
```
<!-- GENERATED_RUST_END -->
</details>

---

## Design Philosophy

### Explicit (Non-Eager) Resolution
Go, Kotlin, and Rust use an explicit, non-eager (lazy) resolution model.
*   **Why?** Eager resolution (resolving everything during module initialization or `init` blocks) can cause dangerous side-effects, makes unit testing and mocking difficult, and violates best practices.
*   **How it works**: The generated code defines unresolved specifications (`FileSpec`, `ExecutableSpec`, `DirectorySpec`, `FileSetSpec`). Because runfile resolution can fail at runtime (e.g., if a data dependency is missing from the runfiles manifest), resolution is explicit and fallible. The developer must call `.Resolve()` (Go, which returns an error), `.resolve()` (Kotlin, which throws a `RunfileResolutionException`), or `.resolve()` (Rust, which returns a `Result`) at runtime to obtain the resolved runfile reference. Go and Rust also provide `.MustResolve()` and `.must_resolve()`, which panic on failure for cases where missing runfiles should be immediately fatal. This avoids initialization-time side-effects, handles failures gracefully, and allows injecting mock resolvers for testing.

### Rich Object Wrapper
Rather than just returning raw string paths, the generators wrap runfiles in rich objects (`Runfile` and `ExecutableRunfile`).
*   This allows us to attach helper methods (like `jvmPath` in Kotlin to get a native `java.nio.file.Path` object).
*   Crucially, it allows us to distinguish **executables** and provide safe execution wrappers (`Cmd` in Go, `processBuilder` in Kotlin, `command` in Rust) that automatically handle the propagation of Bazel runfiles environment variables. This solves the common "nested runfiles" problem where a tool run from a test cannot find its own dependencies.

[^1]: Bazel runfiles discovery relies on environment variables like `RUNFILES_DIR` (path to the runfiles directory) and `RUNFILES_MANIFEST_FILE` (path to the manifest file mapping runfile paths to their physical locations, used when symlinks are not available, e.g., on Windows). If these variables are not propagated to child processes, those processes will fail to resolve their own runfiles. For details, see the [Bazel Runfiles Guide](https://bazel.build/extending/rules#runfiles) and the [Bazel Runfiles Library specification](https://github.com/bazelbuild/bazel/blob/master/tools/cpp/runfiles/runfiles_src.h).
