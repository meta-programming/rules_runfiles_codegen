# Bazel runfiles concepts and semantics

This document defines the core concepts, physical layout, and resolution semantics of Bazel **runfiles** (data dependencies). It clarifies relevant Bazel runfiles concepts as officially documented by the Bazel project ([bazel.build](https://bazel.build)).

## References

The authoritative sources of truth for runfile semantics and specifications are:

*   **Bazel rules guide**: The [official documentation](https://bazel.build/extending/rules#runfiles_location) on runfiles location and symlink structures.
*   **Bazel external dependencies**: The [official documentation](https://bazel.build/external/overview) defining Bzlmod concepts such as canonical names and repository mapping.
*   **Bazel runfiles library specification**: The [official design document](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (written in 2018) detailing the interfaces, expected behavior, and motivations behind the modern runfiles libraries. *Note: This document predates Bzlmod and does not mention repository mapping, which was layered on top of this specification later.*

---

## 1. Anatomy of the runfiles tree

To avoid ambiguity, distinguish between the **runfiles root** and the **workspace subdirectory** within it.

### Runfiles root (runfiles directory / `RUNFILES_DIR`)

The top-level directory containing the entire runfiles structure for a target. It is physically located next to the executable and is typically named `<binary_name>.runfiles`.

*   This is the root of the runfiles tree.
*   All logical `rlocation` paths are relative to this directory.
*   *Citation*: [Bazel rules guide: runfiles location](https://bazel.build/extending/rules#runfiles_location)
*   *Example path*: `bazel-bin/src/main.runfiles/`

### Workspace or repository subdirectory

A directory located directly *under* the runfiles root. It is named after a Bazel repository (either the main workspace or an external dependency).

*   *Example path*: `bazel-bin/src/main.runfiles/my_project/`

### Runfiles-root-relative path (rlocation path)

The logical, canonical path of a runfile relative to the **runfiles root**. It serves as the unique key used to look up a runfile at runtime.

*   **Format**: Since it is relative to the runfiles root, the first segment of the path is always the name of the workspace or repository subdirectory:
    `[workspace_or_repository_name]/[package_path]/[file_name]`
*   *Citation*: [Bazel runfiles library specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Library interface")
*   *Example*: `my_project/data/config.json`

### Physical path vs. logical rlocation path

Using the target `//src:main` in a workspace named `my_project` with a data file `//data:config.json`:

*   **Physical path on disk**: `/path/to/project/bazel-bin/src/main.runfiles/my_project/data/config.json`
*   **Runfiles root**: `/path/to/project/bazel-bin/src/main.runfiles/`
*   **Logical rlocation path** (passed to `Rlocation()`): `my_project/data/config.json`

---

## 2. Physical layouts: directory vs. manifest

At runtime, Bazel prepares the runfiles using one of two physical layouts, depending on the platform and configuration:

### A. The runfiles directory (symlink tree)

On Linux and macOS, Bazel populates the runfiles root with a physical tree of symlinks pointing to the actual files in the source or output trees.

*   *Citation*: [Bazel rules guide: runfiles symlinks](https://bazel.build/extending/rules#runfiles_symlinks)
    > "The runfiles directory contains symlinks to the runfiles [...] The symlinks are structured as follows:
    > `runfiles_dir/workspace_name/package_name/file_name`"

### B. The runfiles manifest

On platforms where symlinks are not supported (primarily Windows) or when sandboxing is disabled, Bazel does not create a symlink tree. Instead, it writes a text file mapping logical `rlocation` paths to physical absolute paths.

*   *Citation*: [Bazel runfiles library specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Motivation")
*   *Layout*: The manifest is located at `path/to/binary.runfiles_manifest` (or `MANIFEST` inside the runfiles directory).
*   *Example line*:
    ```text
    my_project/data/config.json /absolute/path/to/source/data/config.json
    ```

---

## 3. Bzlmod and repository mapping

The transition from the legacy `WORKSPACE` system to `Bzlmod` (Bazel's module system) changed how the first segment of the rlocation path (`workspace_or_repository_name`) is resolved physically.

### Apparent vs. canonical names

*   **Apparent name**: The local name (nickname) given to a dependency within a module (e.g., `@rules_go`). This is what developers write in `BUILD` and `MODULE.bazel` files.
*   **Canonical name**: The globally unique name Bazel generates for a dependency to avoid version conflicts (e.g., `rules_go~0.39.0`).
*   **Physical layout**: Under Bzlmod, the subdirectories under the runfiles root are named using **canonical names**, not apparent names.
    *   *Citation*: [Bazel external dependencies: canonical repository name](https://bazel.build/external/overview#canonical-repo-name)
        > "The canonical repository name of a repository is the name of the directory it occupies in the execution root and in the runfiles directory."

### The special `_main` repository name

Under Bzlmod, the main repository (your project) is assigned the fixed canonical name **`_main`** in the runfiles directory, decoupling it from any arbitrary names defined in legacy `WORKSPACE` files.

*   *Citation*: [Bazel Bzlmod migration guide](https://bazel.build/external/migration) & [Starlark API: `ctx.workspace_name`](https://bazel.build/rules/lib/builtins/ctx#workspace_name)
    > "When `--enable_bzlmod` is active, `ctx.workspace_name` (which acts as the execution root name and runfiles prefix for the main repo) is fixed to the string `_main`. Previously, this was determined by the name defined in the `WORKSPACE` file."
*   *Impact*: If Bzlmod is active, paths to your own project's runfiles will always begin with `_main/` (e.g., `_main/data/config.json`) unless a legacy workspace name is explicitly configured.

### The repository mapping file (`_repo_mapping`)

Because canonical names are unstable and contain version numbers, do not hardcode them in `rlocation` calls. Instead, use the **apparent name**.

To resolve this, Bazel generates a **`_repo_mapping`** file (also referred to as the `repo_mapping` manifest) at the root of the runfiles tree. This file maps the tuple `(apparent_name, current_repo)` to the `canonical_name` currently resolved by Bazel.

*   *Citation*: [Bazel external dependencies: apparent repository name / repository mapping](https://bazel.build/external/overview#apparent-repo-name)
    > "Conversely, this can be understood as a repository mapping: each repo maintains a mapping from "apparent repo name" to a "canonical repo name"."
    
    *(Note: Bazel also generates a `{binary}.repo_mapping` file to help binaries resolve these repository mappings at runtime).*

### Resolution semantics under Bzlmod

When a Bzlmod-aware **runfiles library** resolves an `rlocation` path:

1.  It receives the path using the apparent name: `rules_go/some/file`.
2.  It detects the calling repository context (the repository containing the code executing the lookup).
3.  It looks up the apparent name `rules_go` in the `_repo_mapping` file for that context.
4.  It translates the path to use the canonical name: `rules_go~0.39.0/some/file`.
5.  It performs the physical lookup (via symlink tree or manifest) using the translated canonical path.

---

## 4. Best practices

1.  **Do not hardcode physical paths**: Always use a Bzlmod-aware runfiles library to resolve paths at runtime.
2.  **Use apparent names**: Construct `rlocation` paths using the apparent name of the repository (for example, your own module name or the dependency name declared in `MODULE.bazel`).
3.  **Use `$(rlocationpath)` in BUILD files**: When passing runfile paths to tools via command-line arguments or environment variables in `BUILD` files, use the `$(rlocationpath //target)` helper. This ensures Bazel computes the correct logical path at build time.
    *   *Citation*: [Bazel rules guide: attributes - runfiles](https://bazel.build/extending/rules#runfiles)
