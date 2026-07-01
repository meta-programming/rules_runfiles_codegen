load("@rules_kotlin//kotlin:jvm.bzl", "kt_jvm_library")
load("@rules_runfile_codegen_core//internal:rules.bzl", "runfile_codegen")

def kt_runfile(name, target = None, targets = None, doc = "", base = None, type = "auto"):
    """Creates a runfile entry configuration for Kotlin code generation.

    This helper function constructs a structured dictionary representing a single runfile
    dependency. It is intended to be passed in the `entries` list of `kt_jvm_runfile_library`.

    Args:
        name: The Kotlin property name that will be generated to access this runfile.
            This should follow Kotlin-idiomatic naming conventions (e.g., `configJson`,
            `testData`). The generated code will expose this as a property on a Kotlin object.
        target: The Bazel target label of the runfile. Alias for `targets = [target]`.
        targets: A list of Bazel target labels to include in this entry. If multiple
            targets are specified, this entry is automatically treated as a `FileSet`.
        doc: A descriptive comment for the generated Kotlin property.
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

    return {
        "name": name,
        "targets": resolved_targets,
        "doc": doc,
        "base": base,
        "type": type,
    }

def kt_jvm_runfile_library(name, package, entries, object_name = None, **kwargs):
    """Generates a Kotlin JVM library containing compile-time safe accessors for runfiles.

    Args:
        name: A unique name for this target. The generated `kt_jvm_library` will have this name.
        package: The Kotlin package name for the generated file.
        entries: A list of runfile entries, constructed using the `kt_runfile` helper.
        object_name: The name of the Kotlin `object` (singleton).
        **kwargs: Additional arguments to propagate to the underlying `kt_jvm_library` target.
    """
    if "srcs" in kwargs:
        fail("Cannot specify 'srcs' in kt_jvm_runfile_library, they are generated automatically.")

    # Derive object_name if not provided
    if not object_name:
        clean_name = ""
        for char in name.elems() if hasattr(name, "elems") else name:
            if char.isalnum() or char == "_":
                clean_name += char
            else:
                clean_name += "_"

        parts = [part.capitalize() for part in clean_name.split("_") if part]
        object_name = "".join(parts)

        if not object_name:
            fail("Target name '%s' contains no valid characters for a Kotlin object name." % name)
        if object_name[0].isdigit():
            object_name = "_" + object_name

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
        package = package,
        language = "kotlin",
        object_name = object_name,
        targets = unique_targets,
        config = json.encode(config_data),
        testonly = testonly,
        tags = tags,
    )

    # Public Kotlin library
    kt_jvm_library(
        name = name,
        srcs = [":" + name + "_codegen"],
        deps = [
            "@rules_runfile_codegen_kotlin//runfiles",
        ] + user_deps,
        data = unique_targets + user_data,
        **kwargs
    )
