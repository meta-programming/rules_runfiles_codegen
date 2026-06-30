# Bazel Runfiles Concepts and Semantics

This document defines the core concepts, physical layout, and resolution semantics of Bazel **runfiles** (data-dependencies). This is a clarifying document for developers in this repository. It attempts to write up relevant Bazel runfiles concepts as they are used in this project.

---

## 1. Core Definitions

### Runfiles (Data-Dependencies)
The set of files (assets, tools, configuration files, or other binaries) that a program needs to access at runtime. These are declared in the `data` attribute of a Bazel rule.
*   *Citation*: [Bazel Rules Guide: Runfiles](https://bazel.build/extending/rules#runfiles)
*   *Example*: A Go binary `//src:main` depending on `//data:config.json` via `data = ["//data:config.json"]`.

### Runfiles-Root-Relative Path (Rlocation Path)
The logical, canonical path of a runfile relative to the root of the runfiles tree. It serves as the unique key used to look up a runfile at runtime.
*   *Citation*: [Bazel Runfiles Library Specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Library interface")
*   *Semantics*: 
    *   Must be a relative path.
    *   Must use forward slashes (`/`) as directory separators, even on Windows.
    *   Is case-sensitive on Linux/macOS, and case-insensitive on Windows.
    *   Format: `[repository_name]/[package_path]/[file_name]`
*   *Example*: `my_project/data/config.json`

### Rlocation (Runfile Location) Resolution
The process of mapping a logical **Rlocation Path** to its **physical absolute path** on the local filesystem so the program can access it.
*   *Citation*: [Bazel Runfiles Library Specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Library interface")
*   *Example*: Mapping `my_project/data/config.json` to `/usr/local/home/user/project/data/config.json`.

---

## 2. Physical Layout at Runtime

When a Bazel target is executed, Bazel prepares the runfiles using one of two physical layouts, depending on the platform and configuration:

### A. The Runfiles Directory (Runfiles Tree)
A physical directory created next to the executable containing a tree of symlinks (on Linux/macOS) or copies (on Windows if symlinks are disabled) that mimic the logical `rlocation` paths.
*   *Citation*: [Bazel Rules Guide: Runfiles symlink structure](https://bazel.build/extending/rules#runfiles)
*   *Layout*: If the executable is at `path/to/binary`, the runfiles directory is at `path/to/binary.runfiles/`.
*   *Example*:
    ```text
    path/to/binary.runfiles/
    └── my_project/
        └── data/
            └── config.json -> /absolute/path/to/source/data/config.json
    ```

### B. The Runfiles Manifest
A text file mapping logical `rlocation` paths to physical absolute paths. This is used on platforms where symlinks are not supported (primarily Windows) or when sandboxing is disabled.
*   *Citation*: [Bazel Runfiles Library Specification](https://docs.google.com/document/d/e/2PACX-1vSDIrFnFvEYhKsCMdGdD40wZRBX3m3aZ5HhVj4CtHPmiXKDCxioTUbYsDydjKtFDAzER5eg7OjJWs3V/pub) (Section: "Motivation")
*   *Layout*: If the executable is at `path/to/binary`, the manifest is at `path/to/binary.runfiles_manifest` (or `MANIFEST` inside the runfiles directory).
*   *Example Content*:
    ```text
    my_project/data/config.json /absolute/path/to/source/data/config.json
    ```

---

## 3. Bzlmod and Repository Mapping

The transition from the legacy `WORKSPACE` system to `Bzlmod` (Bazel's module system) introduced critical changes to how `repository_name` (the first segment of the Rlocation Path) is resolved.

### The Problem: Canonical vs. Apparent Names
*   **Apparent Name**: The local name (nickname) given to a dependency within a module (e.g., `@rules_go` or `@rules_cc`). This is what developers write in `BUILD` and `MODULE.bazel` files.
*   **Canonical Name**: The globally unique name Bazel generates for a dependency to avoid version conflicts (e.g., `rules_go~0.39.0` or `rules_go++go_sdk+go_default_sdk`). 
*   *Physical Layout*: Under Bzlmod, the physical **Runfiles Directory** and **Runfiles Manifest** use **canonical names** for their root directories.
*   *Citation*: [Bazel External Dependencies: Canonical Repository Names](https://bazel.build/external/overview#canonical-repo-names)

### The Solution: Repository Mapping (`_repo_mapping`)
Because canonical names are unstable and contain version numbers, developers must never hardcode them in `rlocation` calls. Instead, they must use the **apparent name**.

To bridge this gap, Bazel generates a **`_repo_mapping`** file at the root of the runfiles tree. This file maps the tuple `(apparent_name, current_repo)` to the `canonical_name` currently resolved by Bazel.
*   *Citation*: [Bazel External Dependencies: Repository Mapping](https://bazel.build/external/overview#repository-mapping)

### Resolution Semantics under Bzlmod
When a Bzlmod-aware **Runfiles Library** resolves an `rlocation` path:
1.  It receives the path using the apparent name: `rules_go/some/file`.
2.  It detects the calling repository context (where the code is running).
3.  It looks up the apparent name `rules_go` in the `_repo_mapping` file for that context.
4.  It translates the path to use the canonical name: `rules_go~0.39.0/some/file`.
5.  It performs the physical lookup (via directory or manifest) using the canonical path.

---

## 4. Best Practices for Developers

1.  **Never Hardcode Physical Paths**: Always use a Bzlmod-aware runfiles library to resolve paths at runtime.
2.  **Use Apparent Names**: In your code, always construct `rlocation` paths using the apparent name of the repository (e.g., your own module name or the name declared in `MODULE.bazel`).
3.  **Use `$(rlocationpath)` in BUILD files**: When passing runfile paths to tools via command-line arguments or environment variables in `BUILD` files, always use the `$(rlocationpath //target)` helper. This ensures Bazel computes the correct logical path at build time.
    *   *Citation*: [Bazel Rules Guide: Attributes - Runfiles](https://bazel.build/extending/rules#runfiles)
