<!-- Generated with Stardoc: http://skydoc.bazel.build -->



<a id="rust_runfile"></a>

## rust_runfile

<pre>
load("@rules_runfile_codegen_rust//:defs.bzl", "rust_runfile")

rust_runfile(<a href="#rust_runfile-name">name</a>, <a href="#rust_runfile-target">target</a>, <a href="#rust_runfile-targets">targets</a>, <a href="#rust_runfile-doc">doc</a>, <a href="#rust_runfile-base">base</a>, <a href="#rust_runfile-type">type</a>)
</pre>

Creates a runfile entry configuration for Rust code generation.

This helper function constructs a structured dictionary representing a single runfile
dependency. It is intended to be passed in the `entries` list of `rust_runfile_library`.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="rust_runfile-name"></a>name |  The name of the generated `pub static`, in upper snake case (e.g., `CONFIG_JSON`, `HELPER_TOOL`). Its type is `runfile::FileSpec`, `runfile::ExecutableSpec`, `runfile::DirectorySpec`, or `runfile::FileSetSpec`, depending on the target.   |  none |
| <a id="rust_runfile-target"></a>target |  The Bazel target label of the runfile. Alias for `targets = [target]`.   |  `None` |
| <a id="rust_runfile-targets"></a>targets |  A list of Bazel target labels to include in this entry. If multiple targets are specified, this entry is automatically treated as a file set.   |  `None` |
| <a id="rust_runfile-doc"></a>doc |  The doc comment for the generated static.   |  `""` |
| <a id="rust_runfile-base"></a>base |  An optional path base to resolve relative paths for file set files. - If set to `""` (empty string), nothing is stripped (keeps full canonical paths). - If it starts with `//`, it is resolved relative to the library's repository root. - If it starts with `@`, it resolves relative to an external repository. - If it starts with `.`, it resolves relative to the library's package path. - If set to `"common_dir"`, it automatically computes the longest common prefix.   |  `None` |
| <a id="rust_runfile-type"></a>type |  An optional explicit type assertion for the target. - `"auto"` (default): Automatically detects the type (file, directory, or fileset). - `"file"`: Asserts the target is a single file. - `"directory"`: Asserts the target is a TreeArtifact directory. - `"fileset"`: Forces the target to be treated as a file set. - `"executable"`: Asserts the target is an executable binary.   |  `"auto"` |

**RETURNS**

A dictionary containing the configured runfile entry.


<a id="rust_runfile_library"></a>

## rust_runfile_library

<pre>
load("@rules_runfile_codegen_rust//:defs.bzl", "rust_runfile_library")

rust_runfile_library(<a href="#rust_runfile_library-name">name</a>, <a href="#rust_runfile_library-entries">entries</a>, <a href="#rust_runfile_library-kwargs">**kwargs</a>)
</pre>

Generates a Rust crate containing typed accessors for runfiles.

The generated crate has one `pub static` per entry. Each static is an
unresolved spec from the `runfile` crate
(`@rules_runfile_codegen_rust//runfile`), which looks runfiles up with
`@rules_rust//rust/runfiles`. Declaring a spec does no I/O. Call
`resolve()` to get a `Result`, or `must_resolve()` to panic on a missing
runfile. A resolved executable has `command()`, which returns a
`std::process::Command` that passes the runfiles environment to the child.

The generated source is formatted as `rustfmt` formats it, and every public
item has a doc comment.

### Example Usage

In your `BUILD.bazel`:

```bazel
load("@rules_runfile_codegen_rust//:defs.bzl", "rust_runfile", "rust_runfile_library")

rust_runfile_library(
    name = "my_runfiles",
    entries = [
        rust_runfile(
            name = "CONFIG_JSON",
            target = "//config:config.json",
            doc = "The main configuration file.",
        ),
        rust_runfile(
            name = "HELPER_TOOL",
            target = "//tools:helper_tool",  # An executable target
            doc = "A helper CLI tool.",
        ),
    ],
)
```

In your `main.rs`:

```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = my_runfiles::CONFIG_JSON.resolve()?;
    let text = std::fs::read_to_string(config.path())?;

    let status = my_runfiles::HELPER_TOOL.must_resolve().command().arg("--verbose").status()?;
    Ok(())
}
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="rust_runfile_library-name"></a>name |  A unique name for this target. The generated `rust_library` has this name, and its crate name defaults to it.   |  none |
| <a id="rust_runfile_library-entries"></a>entries |  A list of runfile entries, constructed using the `rust_runfile` helper.   |  none |
| <a id="rust_runfile_library-kwargs"></a>kwargs |  Additional arguments to propagate to the underlying `rust_library` target, such as `crate_name` or `edition` (default `"2021"`).   |  none |


