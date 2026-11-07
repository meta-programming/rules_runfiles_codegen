load("@rules_runfile_codegen_core//internal:rules.bzl", "runfile_codegen")
load("@rules_rust//rust:defs.bzl", "rust_library")

def rust_runfile(name, target = None, targets = None, doc = "", base = None, type = "auto"):
    """Creates a runfile entry configuration for Rust code generation.

    This helper function constructs a structured dictionary representing a single runfile
    dependency. It is intended to be passed in the `entries` list of `rust_runfile_library`.

    Args:
        name: The name of the generated `pub static`, in upper snake case (e.g.,
            `CONFIG_JSON`, `HELPER_TOOL`). Its type is `runfile::FileSpec`,
            `runfile::ExecutableSpec`, `runfile::DirectorySpec`, or
            `runfile::FileSetSpec`, depending on the target.
        target: The Bazel target label of the runfile. Alias for `targets = [target]`.
        targets: A list of Bazel target labels to include in this entry. If multiple
            targets are specified, this entry is automatically treated as a file set.
        doc: The doc comment for the generated static.
        base: An optional path base to resolve relative paths for file set files.
            - If set to `""` (empty string), nothing is stripped (keeps full canonical paths).
            - If it starts with `//`, it is resolved relative to the library's repository root.
            - If it starts with `@`, it resolves relative to an external repository.
            - If it starts with `.`, it resolves relative to the library's package path.
            - If set to `"common_dir"`, it automatically computes the longest common prefix.
        type: An optional explicit type assertion for the target.
            - `"auto"` (default): Automatically detects the type (file, directory, or fileset).
            - `"file"`: Asserts the target is a single file.
            - `"directory"`: Asserts the target is a TreeArtifact directory.
            - `"fileset"`: Forces the target to be treated as a file set.
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

def rust_runfile_library(name, entries, **kwargs):
    """Generates a Rust crate containing typed accessors for runfiles.

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

    Args:
        name: A unique name for this target. The generated `rust_library` has this
            name, and its crate name defaults to it.
        entries: A list of runfile entries, constructed using the `rust_runfile` helper.
        **kwargs: Additional arguments to propagate to the underlying `rust_library`
            target, such as `crate_name` or `edition` (default `"2021"`).
    """
    if "srcs" in kwargs or "crate_root" in kwargs:
        fail("Cannot specify 'srcs' or 'crate_root' in rust_runfile_library, they are generated automatically.")

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
    edition = kwargs.pop("edition", "2021")

    # Call the core generator rule
    runfile_codegen(
        name = name + "_codegen",
        package = kwargs.get("crate_name", name),
        language = "rust",
        targets = unique_targets,
        config = json.encode(config_data),
        testonly = testonly,
        tags = tags,
    )

    # Public Rust library
    rust_library(
        name = name,
        srcs = [":" + name + "_codegen"],
        crate_root = ":" + name + "_codegen",
        edition = edition,
        deps = [
            "@rules_runfile_codegen_rust//runfile",
        ] + user_deps,
        data = unique_targets + user_data,
        **kwargs
    )
