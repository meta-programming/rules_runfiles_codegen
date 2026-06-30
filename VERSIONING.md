# Versioning Policy

This document defines the versioning policy for the `rules_runfile_codegen` project. 

Since this project is a **multi-module repository** (containing `core`, `go`, and `kotlin` modules) and acts as an extension to existing Bazel rulesets ([rules_go](https://github.com/bazelbuild/rules_go), [rules_kotlin](https://github.com/bazelbuild/rules_kotlin)), we must choose a versioning strategy that balances user clarity, compatibility tracking, and maintenance overhead.

---

## Proposed Strategy: Unified Independent SemVer

We use a **Unified Independent SemVer** scheme, supplemented by a clear **Compatibility Matrix** in our documentation.

### How it works:
1.  **Shared Versioning**: All three modules (`core`, [`go`](go/docs/defs.md), [`kotlin`](kotlin/docs/defs.md)) share the **same version number** (e.g., `0.1.0`, `1.0.0`) and are released together from a single Git tag (e.g., `v0.1.0`).
2.  **Strict SemVer**: We version our releases strictly according to our own API changes:
    *   **Major** (x.0.0): Breaking changes to our macros (e.g., removing/renaming attributes, changing generated code APIs).
    *   **Minor** (0.y.0): New features (e.g., adding support for a new language, adding new helper methods to [`ExecutableRunfile`](README.md#rich-object-wrapper)).
    *   **Patch** (0.0.z): Bug fixes and documentation updates.
3.  **Compatibility Matrix**: We maintain a prominent compatibility table in our [README.md](README.md) showing which versions of `rules_go` and `rules_kotlin` are supported by each of our releases.

### Compatibility Matrix

| `rules_runfile_codegen` Version | Supported `rules_go` | Supported `rules_kotlin` | Minimum Bazel Version |
| :--- | :--- | :--- | :--- |
| **`0.1.x`** (Actual) | `>= 0.60.0` (likely compatible back to `v0.38.0`) | `>= 2.4.0` (likely compatible back to `v1.9.0`) | `6.4.0` |
| **`1.0.x`** (Hypothetical) | `>= 0.60.0` (likely compatible back to `v0.38.0`) | `>= 2.4.0` (likely compatible back to `v1.9.0`) | `6.4.0` |
| **`1.1.x`** (Hypothetical) | `>= 0.60.0` (likely compatible back to `v0.38.0`) | `>= 2.4.0` (likely compatible back to `v1.9.0`) | `6.4.0` |

*The minimum versions of `rules_go` (0.60.0) and `rules_kotlin` (2.4.0) are selected to align with the stable, modern releases used during the development and testing of this library, ensuring robust Bzlmod support. Note that there is no direct dependency on `rules_java` for the Kotlin runtime; it relies on the built-in `@bazel_tools//tools/java/runfiles` library which is always available in Bazel.*

### Bzlmod Compatibility Level
Since Bzlmod's `compatibility_level` is deprecated in newer Bazel versions (8.6.0+ / 9.1.0+) (see the [deprecation notice](https://bazel.build/rules/lib/globals/module)), we do not specify it in our `module()` declarations. For older Bazel versions (6.x / 7.x) where it is still active, omitting the attribute effectively defaults the compatibility level to `0` for all our releases.

---

## Compatibility & Incompatibility Factors

A key reason for choosing **Independent SemVer** is that our library's compatibility with the underlying rulesets ([rules_go](https://github.com/bazelbuild/rules_go), [rules_kotlin](https://github.com/bazelbuild/rules_kotlin)) is stable, meaning we do not need to couple our releases to theirs.

### What could cause an incompatibility?
Incompatibilities would only be expected to arise from two sources:
1.  **Breaking changes in the runfiles library APIs**: If `rules_go` removes or renames `runfiles.New()` or `Rlocation()`, or if the Java runfiles library (`@bazel_tools//tools/java/runfiles`) removes or renames `Runfiles.preload()`. This is **unlikely**, as these APIs implement the core [Bazel Runfiles Specification](https://bazel.build/extending/rules#runfiles) which has been stable for many years.
2.  **Breaking changes in macro signatures**: If [rules_go](https://github.com/bazelbuild/rules_go) changes the signature of [`go_library`](https://github.com/bazelbuild/rules_go/blob/master/go/core.rst#go_library) (e.g., removing the `deps` or `data` attributes), or if [rules_kotlin](https://github.com/bazelbuild/rules_kotlin) does the same for [`kt_jvm_library`](https://github.com/bazelbuild/rules_kotlin). This is also rare, as these are core Bazel rules.

### Supporting Multiple Major Versions
Because the underlying APIs are stable, **a single release of `rules_runfile_codegen` can support multiple major versions of `rules_go` or `rules_kotlin` simultaneously.** 

For example, [`rules_runfile_codegen_go`](go/docs/defs.md) v1.0.0 is compatible with both `rules_go` v0.48.0 and a future `rules_go` v1.0.0 or v2.0.0, without requiring any changes or new releases on our part. We only need to specify a *minimum* version in our `MODULE.bazel` to ensure basic features are present; Bzlmod's Minimal Version Selection (MVS) handles this.

### Go Compatibility Factors

Our Go module ([`rules_runfile_codegen_go`](go/docs/defs.md)) relies on the `go_library` rule from `rules_go` and the `@rules_go//go/runfiles` library. Both are stable components of the Go ruleset.
*   **`go_library` Stability**: We only use standard attributes (`srcs`, `deps`, `data`, `visibility`) which have remained unchanged for many versions.
*   **Runfiles Library**: The `@rules_go//go/runfiles` API (specifically `runfiles.New()` and `Rlocation()`) is stable and adheres to the Bazel runfiles specification.
*   **Bzlmod Support**: `rules_go` introduced experimental Bzlmod support in `v0.33.0`, which stabilized significantly by `v0.38.0`.
*   **Conclusion**: While we officially test against and support `rules_go` `>= 0.60.0` (to align with modern development environments), our implementation is likely compatible back to **`v0.38.0`**.

### Kotlin Compatibility Factors

Our Kotlin module ([`rules_runfile_codegen_kotlin`](kotlin/docs/defs.md)) relies on the `kt_jvm_library` rule from `rules_kotlin` and the `@bazel_tools//tools/java/runfiles` library.
*   **`kt_jvm_library` Stability**: Similar to Go, we only use standard attributes that are stable across major versions.
*   **Runfiles Library**: We currently depend on `@bazel_tools//tools/java/runfiles`, which is built into Bazel and is stable. However, there are ongoing efforts to migrate this library:
    *   [rules_java Issue #46](https://github.com/bazelbuild/rules_java/issues/46) discusses migrating the runfiles library to `rules_java`.
    *   [rules_java Issue #360](https://github.com/bazelbuild/rules_java/issues/360) discusses releasing it as a Maven coordinate.
    Future versions of our library may need to transition to these new dependency paths, but the current implementation remains compatible with the legacy `@bazel_tools` path.
*   **Bzlmod Support**: `rules_kotlin` introduced Bzlmod support in **`v1.9.0`**.
*   **Conclusion**: We officially test against `rules_kotlin` `>= 2.4.0`, but we are likely compatible back to **`v1.9.0`** (the earliest version with Bzlmod support).

---

## Advantages of this Policy

*   **Simplicity**: We only manage one release tag (`v0.1.0`) per repository release. The automation (`publish-to-bcr` Action) packages all three modules under this single version.
*   **Guaranteed Core Compatibility**: Since `core`, `go`, and `kotlin` are released together under the same version, they are compatible with each other. [`rules_runfile_codegen_go`](go/docs/defs.md) v0.1.0 will always depend on `rules_runfile_codegen_core` v0.1.0.
*   **Bzlmod Integration**: Bzlmod's Minimal Version Selection (MVS) handles this well. If a user's project uses `rules_runfile_codegen_go` and another dependency uses it too, Bzlmod will resolve to the newest compatible version.
*   **Broad Compatibility**: We can support a wide range of underlying ruleset versions, giving users the freedom to upgrade their toolchains independently of our library.

---

## Alternatives Considered

We also considered letting each module have its own independent version number (e.g., `core` is `v1.2.0`, [`go`](go/docs/defs.md) is `v0.8.0`, [`kotlin`](kotlin/docs/defs.md) is `v2.1.0`) without mirroring the underlyings.

*   **Why we rejected it**: While this offers greater flexibility, it introduces additional overhead in a mono-repo. We would need to manage multiple Git tags for a single commit, and we would have to constantly update the inter-module dependencies (e.g., updating [`go/MODULE.bazel`](go/MODULE.bazel) to point to the new version of `core` every time `core` changes), leading to a "version churn" cycle.
