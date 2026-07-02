<!-- Generated with Stardoc: http://skydoc.bazel.build -->



<a id="kt_jvm_runfile_library"></a>

## kt_jvm_runfile_library

<pre>
load("@rules_runfile_codegen_kotlin//:defs.bzl", "kt_jvm_runfile_library")

kt_jvm_runfile_library(<a href="#kt_jvm_runfile_library-name">name</a>, <a href="#kt_jvm_runfile_library-package">package</a>, <a href="#kt_jvm_runfile_library-entries">entries</a>, <a href="#kt_jvm_runfile_library-object_name">object_name</a>, <a href="#kt_jvm_runfile_library-kwargs">**kwargs</a>)
</pre>

Generates a Kotlin JVM library containing compile-time safe accessors for runfiles.

This macro automates the process of accessing Bazel runfiles in Kotlin. It generates
a Kotlin source file containing a singleton `object` with properties that hold the
resolved runfile paths.

The generated library uses the official Bazel Java runfiles library (`@bazel_tools//tools/java/runfiles`)
to resolve the runfiles at runtime.

### Example Usage

In your `BUILD.bazel`:

```bazel
load("@rules_runfile_codegen_kotlin//:defs.bzl", "kt_runfile", "kt_jvm_runfile_library")

kt_jvm_runfile_library(
    name = "my_runfiles",
    package = "com.example.project.runfiles",
    entries = [
        kt_runfile(
            name = "configJson",
            target = "//config:config.json",
            doc = "The main configuration file.",
        ),
        kt_runfile(
            name = "helperTool",
            target = "//tools:helper_tool", # An executable target
            doc = "A helper CLI tool.",
        ),
        kt_runfile(
            name = "exampleSet",
            targets = ["//data:file1.txt", "//data:file2.txt"],
            base = "common_dir",
            doc = "A set of data files.",
        ),
    ],
)
```

In your `Main.kt`:

```kotlin
package com.example.project

import com.example.project.runfiles.MyRunfiles
import kotlin.io.path.readText

fun main() {
    // 1. Accessing a regular runfile:
    // Resolve the spec and read its content directly using Path.readText().
    val content = MyRunfiles.configJson.resolve().path.readText()
    println("Content: $content")

    // The resolved path is a java.nio.file.Path
    val path = MyRunfiles.configJson.resolve().path

    // 2. Running an executable runfile:
    // Resolve, configure, start, and wait for the process in a fluent chain.
    val exitCode = MyRunfiles.helperTool.resolve()
        .processBuilder("--verbose", "run")
        .inheritIO()
        .start()
        .waitFor()
    if (exitCode != 0) {
        error("Helper tool failed with exit code $exitCode")
    }

    // 3. Accessing a fileset:
    val fileset = MyRunfiles.exampleSet.resolve()
    // Access file inside fileset by its relative path (after base stripping)
    val f1 = fileset.resolveFile("file1.txt")
    // ... read f1.path.readText()
}
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="kt_jvm_runfile_library-name"></a>name |  A unique name for this target. The generated `kt_jvm_library` will have this name.   |  none |
| <a id="kt_jvm_runfile_library-package"></a>package |  The Kotlin package name for the generated file.   |  none |
| <a id="kt_jvm_runfile_library-entries"></a>entries |  A list of runfile entries, constructed using the `kt_runfile` helper.   |  none |
| <a id="kt_jvm_runfile_library-object_name"></a>object_name |  The name of the Kotlin `object` (singleton).   |  `None` |
| <a id="kt_jvm_runfile_library-kwargs"></a>kwargs |  Additional arguments to propagate to the underlying `kt_jvm_library` target.   |  none |


<a id="kt_runfile"></a>

## kt_runfile

<pre>
load("@rules_runfile_codegen_kotlin//:defs.bzl", "kt_runfile")

kt_runfile(<a href="#kt_runfile-name">name</a>, <a href="#kt_runfile-target">target</a>, <a href="#kt_runfile-targets">targets</a>, <a href="#kt_runfile-doc">doc</a>, <a href="#kt_runfile-base">base</a>, <a href="#kt_runfile-type">type</a>)
</pre>

Creates a runfile entry configuration for Kotlin code generation.

This helper function constructs a structured dictionary representing a single runfile
dependency. It is intended to be passed in the `entries` list of `kt_jvm_runfile_library`.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="kt_runfile-name"></a>name |  The Kotlin property name that will be generated to access this runfile. This should follow Kotlin-idiomatic naming conventions (e.g., `configJson`, `testData`). The generated code will expose this as a property on a Kotlin object.   |  none |
| <a id="kt_runfile-target"></a>target |  The Bazel target label of the runfile. Alias for `targets = [target]`.   |  `None` |
| <a id="kt_runfile-targets"></a>targets |  A list of Bazel target labels to include in this entry. If multiple targets are specified, this entry is automatically treated as a `FileSet`.   |  `None` |
| <a id="kt_runfile-doc"></a>doc |  A descriptive comment for the generated Kotlin property.   |  `""` |
| <a id="kt_runfile-base"></a>base |  An optional path base to resolve relative paths for FileSet files. - If set to `""` (empty string), nothing is stripped (keeps full canonical paths). - If it starts with `//`, it is resolved relative to the library's repository root. - If it starts with `@`, it resolves relative to an external repository. - If it starts with `.`, it resolves relative to the library's package path. - If set to `"common_dir"`, it automatically computes the longest common prefix.   |  `None` |
| <a id="kt_runfile-type"></a>type |  An optional explicit type assertion for the target. - `"auto"` (default): Automatically detects the type (file, directory, or fileset). - `"file"`: Asserts the target is a single file. - `"directory"`: Asserts the target is a TreeArtifact directory. - `"fileset"`: Forces the target to be treated as a FileSet. - `"executable"`: Asserts the target is an executable binary.   |  `"auto"` |

**RETURNS**

A dictionary containing the configured runfile entry.


