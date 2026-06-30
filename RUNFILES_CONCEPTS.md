# Bazel Runfiles Concepts and Semantics

This document defines the core concepts, physical layout, and resolution semantics of Bazel **runfiles** (data-dependencies). This is a clarifying document for developers in this repository. It attempts to write up relevant Bazel runfiles concepts as they are officially documented by specifications and documents by the Bazel project (bazel.build).

---

## 1. Anatomy of the Runfiles Tree

To avoid ambiguity, it is critical to distinguish between the **Runfiles Root** and the **Workspace Subdirectory** within it.

### Runfiles Root (Runfiles Directory / `RUNFILES_DIR`)
The top-level directory containing the entire runfiles structure for a target. It is physically located next to the executable and is typically named `<binary_name>.runfiles`.
*   **This is the root of the runfiles tree.**
*   All logical `rlocation` paths are relative to this directory.
*   *Citation*: [Bazel Rules Guide: Runfiles Location](https://bazel.build/extending/rules#runfiles_location)
*   *Example Path*: `bazel-bin/src/main.runfiles/`

### Workspace/Repository Subdirectory
A directory located directly *under* the Runfiles Root. It is named after a Bazel repository (either the main workspace or an external dependency).
*   *Example Path*: `bazel-bin/src/main.runfiles/my_project/`

### Runfiles-Root-Relative Path (Rlocation Path)
The logical, canonical path of a runfile relative to the **Runfiles Root**. It serves as the unique key used to look up a runfile at runtime.
*   **Format**: Since it is relative to the Runfiles Root, the first segment of the path is always the name of the Workspace/Repository Subdirectory:
    `[workspace_or_repository_name]/[package_path]/[file_name]`
*   *Citation*: [Bazel Runfiles Library Specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Library interface")
*   *Example*: `my_project/data/config.json`

### Physical Path vs. Logical Rlocation Path
Using the target `//src:main` in a workspace named `my_project` with a data file `//data:config.json`:
*   **Physical Path on disk**: 
    `/path/to/project/bazel-bin/src/main.runfiles/my_project/data/config.json`
*   **Runfiles Root**: 
    `/path/to/project/bazel-bin/src/main.runfiles/`
*   **Logical Rlocation Path** (passed to `Rlocation()`): 
    `my_project/data/config.json`

---

## 2. Physical Layouts: Directory vs. Manifest

At runtime, Bazel prepares the runfiles using one of two physical layouts, depending on the platform and configuration:

### A. The Runfiles Directory (Symlink Tree)
On Linux and macOS, Bazel populates the Runfiles Root with a physical tree of symlinks pointing to the actual files in the source or output trees.
*   *Citation*: [Bazel Rules Guide: Runfiles Symlinks](https://bazel.build/extending/rules#runfiles_symlinks)
    > "The runfiles directory contains symlinks to the runfiles [...] The symlinks are structured as follows:
    > `runfiles_dir/workspace_name/package_name/file_name`"

### B. The Runfiles Manifest
On platforms where symlinks are not supported (primarily Windows) or when sandboxing is disabled, Bazel does not create a symlink tree. Instead, it writes a text file mapping logical `rlocation` paths to physical absolute paths.
*   *Citation*: [Bazel Runfiles Library Specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Motivation")
*   *Layout*: The manifest is located at `path/to/binary.runfiles_manifest` (or `MANIFEST` inside the runfiles directory).
*   *Example Line*:
    ```text
    my_project/data/config.json /absolute/path/to/source/data/config.json
    ```

---

## 3. Bzlmod and Repository Mapping

The transition from the legacy `WORKSPACE` system to `Bzlmod` (Bazel's module system) changed how the first segment of the Rlocation Path (`workspace_or_repository_name`) is resolved physically.

### Apparent vs. Canonical Names
*   **Apparent Name**: The local name (nickname) given to a dependency within a module (e.g., `@rules_go`). This is what developers write in `BUILD` and `MODULE.bazel` files.
*   **Canonical Name**: The globally unique name Bazel generates for a dependency to avoid version conflicts (e.g., `rules_go~0.39.0`).
*   **Physical Layout**: Under Bzlmod, the subdirectories under the Runfiles Root are named using **canonical names**, not apparent names.
    *   *Citation*: [Bazel External Dependencies: Canonical Repository Name](https://bazel.build/external/overview#canonical-repo-name)
        > "The canonical repository name of a repository is the name of the directory it occupies in the execution root and in the runfiles directory."

### The Repository Mapping File (`_repo_mapping`)
Because canonical names are unstable and contain version numbers, developers must never hardcode them in `rlocation` calls. Instead, they must use the **apparent name**.

To resolve this, Bazel generates a **`_repo_mapping`** file (also referred to as the `repo_mapping` manifest) at the root of the runfiles tree. This file maps the tuple `(apparent_name, current_repo)` to the `canonical_name` currently resolved by Bazel.
*   *Citation*: [Bazel External Dependencies: Apparent Repository Name / Repository Mapping](https://bazel.build/external/overview#apparent-repo-name)
    > "Conversely, this can be understood as a repository mapping: each repo maintains a mapping from "apparent repo name" to a "canonical repo name"."
    
    *(Note: Bazel also generates a `{binary}.repo_mapping` file to help binaries resolve these repository mappings at runtime).*

### Resolution Semantics under Bzlmod
When a Bzlmod-aware **Runfiles Library** resolves an `rlocation` path:
1.  It receives the path using the apparent name: `rules_go/some/file`.
2.  It detects the calling repository context (the repository containing the code executing the lookup).
3.  It looks up the apparent name `rules_go` in the `_repo_mapping` file for that context.
4.  It translates the path to use the canonical name: `rules_go~0.39.0/some/file`.
5.  It performs the physical lookup (via symlink tree or manifest) using the translated canonical path.

---

## 4. Best Practices

1.  **Never Hardcode Physical Paths**: Always use a Bzlmod-aware runfiles library to resolve paths at runtime.
2.  **Use Apparent Names**: In your code, always construct `rlocation` paths using the apparent name of the repository (e.g., your own module name or the dependency name declared in `MODULE.bazel`).
3.  **Use `$(rlocationpath)` in BUILD files**: When passing runfile paths to tools via command-line arguments or environment variables in `BUILD` files, always use the `$(rlocationpath //target)` helper. This ensures Bazel computes the correct logical path at build time.
    *   *Citation*: [Bazel Rules Guide: Attributes - Runfiles](https://bazel.build/extending/rules#runfiles)
