# Bazel Runfile Code Generators (`rules_runfile_codegen`)

In Bazel, [**runfiles**](https://bazel.build/extending/rules#runfiles) are the files (data dependencies, configuration files, or other executables) that a binary needs at runtime. Accessing these files programmatically requires resolving their paths relative to the workspace, which can be complex and error-prone because their physical locations change depending on the execution environment (e.g., running locally, inside a sandbox during `bazel test`, or in a production deployment).

Traditionally, developers must use Bazel's language-specific runfiles libraries to perform string-based lookups at runtime. This project, `rules_runfile_codegen`, simplifies this by generating type-safe code accessors for your runfiles.

By defining your runfile dependencies in your `BUILD.bazel` files, you can generate libraries that expose these runfiles as strongly-typed symbols. This eliminates the need to hardcode runfile paths as strings in your application code, prevents typos, and ensures that runfile resolution errors are caught at startup rather than deep in runtime execution.

For complete, runnable projects demonstrating these quickstarts, see the [examples/](examples) directory.

---

## Key Features

*   **Type-Safety**: Runfiles are exposed as generated constants/properties. No more stringly-typed paths.
*   **Init-time (Eager) Resolution**: Runfiles are resolved at program startup (`init` blocks). If a runfile is missing, the program panics (Go) or throws (Kotlin) immediately, ensuring **fail-at-startup** safety. *(Note: This behavior may be configurable in the future, pending further design.)*
*   **Subprocess Environment Propagation**: Executable runfiles are wrapped in rich objects that facilitate launching them as subprocesses while automatically propagating the Bazel runfiles environment. This ensures that child processes can also resolve their own runfiles.[^1]
*   **Zero Runtime Overhead**: After successful startup-time resolution, accessing the runfile path is a simple member access with zero overhead.

---

## Go Quickstart

### Bzlmod Setup

To use these rules in your Go project, add the following to your `MODULE.bazel` file:

<!-- GO_INSTALL_START -->
```bazel
# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_go", version = "0.1.0")
```
<!-- GO_INSTALL_END -->

> [!NOTE]
> **Pre-release Usage (Local Overrides)**:
> Since this library is not yet published to the BCR, you must use `local_path_override` pointing to a local clone of this repository. Because overrides are not transitive, you must also explicitly override the core module:
>
> ```bazel
> # Core Module (Required for local development)
> bazel_dep(name = "rules_runfile_codegen_core", version = "0.0.0")
> local_path_override(
>     module_name = "rules_runfile_codegen_core",
>     path = "/path/to/runfile-codegen/repo/core",
> )
>
> # Go Module
> bazel_dep(name = "rules_runfile_codegen_go", version = "0.0.0")
> local_path_override(
>     module_name = "rules_runfile_codegen_go",
>     path = "/path/to/runfile-codegen/repo/go",
> )
> ```

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

	"github.com/example/project/examples/go/resources"
)

func main() {
	// 1. Access the resolved runfile path.
	// Resources are resolved at startup (init-time).
	path := resources.DataFile.Path()

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading runfile: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Data: %s\n", string(content))

	// 2. Run an executable runfile with env propagation.
	cmd := resources.HelperTool.Cmd()
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running helper: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Helper output: %s", string(output))
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
	"fmt"
	"os"
	"os/exec"
	"github.com/bazelbuild/rules_go/go/runfiles"
)

// Runfile represents a resolved runfile.
type Runfile struct {
	absPath string
}

// Path returns the resolved absolute path of the runfile.
func (r Runfile) Path() string {
	return r.absPath
}

// ExecutableRunfile represents a resolved executable runfile.
type ExecutableRunfile struct {
	Runfile
}

// Cmd returns an *exec.Cmd pre-configured to run this executable,
// with the Bazel runfiles environment variables already propagated.
func (e ExecutableRunfile) Cmd(args ...string) *exec.Cmd {
	cmd := exec.Command(e.Path(), args...)
	if env, err := runfiles.Env(); err == nil {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd
}

var (
	// DataFile is A dummy data file.
	// Source: @@//:data/dummy.txt
	DataFile Runfile

	// HelperTool is A helper tool executable.
	// Source: @@//:helper
	HelperTool ExecutableRunfile

)

func init() {
	reg, err := runfiles.New()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize runfiles: %v", err))
	}

	DataFile = mustResolve(reg, "_main/data/dummy.txt")
	HelperTool = mustResolveExecutable(reg, "_main/helper_/helper")
}

func mustResolve(reg *runfiles.Runfiles, rlocationPath string) Runfile {
	absPath, err := reg.Rlocation(rlocationPath)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve runfile %q: %v", rlocationPath, err))
	}
	return Runfile{absPath: absPath}
}

func mustResolveExecutable(reg *runfiles.Runfiles, rlocationPath string) ExecutableRunfile {
	absPath, err := reg.Rlocation(rlocationPath)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve executable runfile %q: %v", rlocationPath, err))
	}
	return ExecutableRunfile{Runfile{absPath: absPath}}
}
```
<!-- GENERATED_GO_END -->
</details>

---

## Kotlin Quickstart

### Bzlmod Setup

To use these rules in your Kotlin project, add the following to your `MODULE.bazel` file:

<!-- KOTLIN_INSTALL_START -->
```bazel
# MODULE.bazel
bazel_dep(name = "rules_runfile_codegen_kotlin", version = "0.1.0")
```
<!-- KOTLIN_INSTALL_END -->

> [!NOTE]
> **Pre-release Usage (Local Overrides)**:
> Since this library is not yet published to the BCR, you must use `local_path_override` pointing to a local clone of this repository. Because overrides are not transitive, you must also explicitly override the core module:
>
> ```bazel
> # Core Module (Required for local development)
> bazel_dep(name = "rules_runfile_codegen_core", version = "0.0.0")
> local_path_override(
>     module_name = "rules_runfile_codegen_core",
>     path = "/path/to/runfile-codegen/repo/core",
> )
>
> # Kotlin Module
> bazel_dep(name = "rules_runfile_codegen_kotlin", version = "0.0.0")
> local_path_override(
>     module_name = "rules_runfile_codegen_kotlin",
>     path = "/path/to/runfile-codegen/repo/kotlin",
> )
> ```

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
import java.io.File

fun main() {
    // 1. Access the resolved runfile path.
    // Resources are resolved at startup (init-time).
    val path = Resources.configJson.path
    val content = File(path).readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    val process = Resources.helperTool.processBuilder().start()
    val output = process.inputStream.bufferedReader().readText().trim()
    process.waitFor()
    println("Helper output: $output")
}
```
<!-- KOTLIN_USAGE_END -->

### Actual Generated Kotlin Code

<details>
<summary>Click to view the actual generated Kotlin code for the example above</summary>

<!-- GENERATED_KOTLIN_START -->
```kotlin
// Code generated by rules_runfile_codegen. DO NOT EDIT.
// This file provides type-safe access to Bazel runfiles.
package com.example.project.examples.resources

import com.google.devtools.build.runfiles.Runfiles

object Resources {
    private val _runfiles: Runfiles = try {
        Runfiles.create()
    } catch (e: java.io.IOException) {
        throw java.lang.RuntimeException("Failed to create runfiles registry", e)
    }

    open class Runfile(val path: String) {
        val jvmPath: java.nio.file.Path by lazy { java.nio.file.Paths.get(path) }
    }

    class ExecutableRunfile(path: String) : Runfile(path) {
        fun processBuilder(vararg args: String): ProcessBuilder {
            val pb = ProcessBuilder(path, *args)
            pb.environment().putAll(Resources._runfiles.envVars)
            return pb
        }
    }

    private fun _mustResolve(rlocationPath: String): Runfile {
        val absPath = _runfiles.rlocation(rlocationPath)
            ?: throw java.lang.RuntimeException("Failed to resolve runfile: $rlocationPath")
        return Runfile(absPath)
    }

    private fun _mustResolveExecutable(rlocationPath: String): ExecutableRunfile {
        val absPath = _runfiles.rlocation(rlocationPath)
            ?: throw java.lang.RuntimeException("Failed to resolve executable runfile: $rlocationPath")
        return ExecutableRunfile(absPath)
    }

    /**
     * A dummy data file.
     * Source: @@//:data/dummy.txt
     */
    val configJson: Runfile = _mustResolve("_main/data/dummy.txt")

    /**
     * A helper tool executable.
     * Source: @@//:helper
     */
    val helperTool: ExecutableRunfile = _mustResolveExecutable("_main/helper")

}
```
<!-- GENERATED_KOTLIN_END -->
</details>

---

## Design Philosophy

### Eager (Init-time) Resolution
We deliberately chose **eager resolution** over lazy resolution for both Go and Kotlin. 
*   **The Problem with Lazy Resolution**: If a runfile is missing (e.g., due to a misconfigured `data` dependency in a `BUILD.bazel` file), a lazy implementation would only fail when the code actually tries to access the runfile. This could happen hours or days into a production deployment.
*   **The Eager Solution**: By resolving all runfiles during module initialization (`init` blocks), we guarantee that if *any* runfile is missing, the program will fail immediately at startup. This aligns with the "fail-fast" principle and makes deployments much safer.
*   **Future Configurability**: Note that this behavior may be configurable in the future (e.g., allowing lazy resolution via a macro attribute), pending further design.

### Rich Object Wrapper
Rather than just returning raw string paths, the generators wrap runfiles in rich objects (`Runfile` and `ExecutableRunfile`).
*   This allows us to attach helper methods (like `jvmPath` in Kotlin to get a native `java.nio.file.Path` object).
*   Crucially, it allows us to distinguish **executables** and provide safe execution wrappers (`Cmd` in Go, `processBuilder` in Kotlin) that automatically handle the propagation of Bazel runfiles environment variables. This solves the common "nested runfiles" problem where a tool run from a test cannot find its own dependencies.

[^1]: Bazel runfiles discovery relies on environment variables like `RUNFILES_DIR` (path to the runfiles directory) and `RUNFILES_MANIFEST_FILE` (path to the manifest file mapping runfile paths to their physical locations, used when symlinks are not available, e.g., on Windows). If these variables are not propagated to child processes, those processes will fail to resolve their own runfiles. For details, see the [Bazel Runfiles Guide](https://bazel.build/extending/rules#runfiles) and the [Bazel Runfiles Library specification](https://github.com/bazelbuild/bazel/blob/master/tools/cpp/runfiles/runfiles_src.h).
