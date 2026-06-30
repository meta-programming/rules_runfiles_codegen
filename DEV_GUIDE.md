# Developer Guide: rules_runfile_codegen

This guide explains how to set up a local development environment for `rules_runfile_codegen`, how the project is structured for testing, and how to run the integration tests.

## Local Development Setup (Bzlmod)

During development, you will want to test your changes locally. Because this project uses Bzlmod and is split into multiple modules, you must configure your test workspaces to point to your local clone.

### The Non-Transitive Override Gotcha

> [!IMPORTANT]
> In Bzlmod, `local_path_override` (and `git_override`) is **non-transitive**. It only applies to the **root module** being built. 
> 
> If you are testing a binder module (like `rules_runfile_codegen_go`) locally, you **must also** explicitly override the core module (`rules_runfile_codegen_core`) in your test project's `MODULE.bazel`. If you do not, Bazel will attempt to resolve the core module from the registry (which will fail or use an outdated version).

### Example `MODULE.bazel` Configuration for Local Testing

To test Go and Kotlin support locally, add the following to your test project's `MODULE.bazel`:

```bazel
# 1. Core Module (Required by all binders)
bazel_dep(name = "rules_runfile_codegen_core", version = "0.0.0")
local_path_override(
    module_name = "rules_runfile_codegen_core",
    path = "/path/to/rules_runfile_codegen/repo/core",
)

# 2. Go Binder Module
bazel_dep(name = "rules_runfile_codegen_go", version = "0.0.0")
local_path_override(
    module_name = "rules_runfile_codegen_go",
    path = "/path/to/rules_runfile_codegen/repo/go",
)

# 3. Kotlin Binder Module
bazel_dep(name = "rules_runfile_codegen_kotlin", version = "0.0.0")
local_path_override(
    module_name = "rules_runfile_codegen_kotlin",
    path = "/path/to/rules_runfile_codegen/repo/kotlin",
)
```
*(Replace `/path/to/rules_runfile_codegen` with the absolute path to your local clone).*

---

## Integration Tests Structure

To prevent test-only dependencies (like `rules_shell` or test runners) from leaking to users, the integration tests are structured as **completely separate Bazel workspaces** located in `repo/tests/`.

*   **`repo/tests/go/`**: A standalone Bazel workspace testing the Go binder.
*   **`repo/tests/kotlin/`**: A standalone Bazel workspace testing the Kotlin binder.

These workspaces import the local runtime modules using the `local_path_override` strategy described above.

### Running the Tests

To run the integration tests, navigate to the respective test directory and run `bazel test //...`:

#### Run Go Tests:
```bash
cd repo/tests/go
bazel test //...
```

#### Run Kotlin Tests:
```bash
cd repo/tests/kotlin
bazel test //...
```

---

## Updating the Documentation (README.md)

The **[README.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/README.md)** contains sections showing the "Actual Generated Code" for both Go and Kotlin. To ensure these snippets are 100% accurate and compiler-verified, they are automatically synchronized from the integration test outputs.

If you make changes to the code generators, you should update the README before committing:

1.  **Generate the latest outputs** by running the integration tests (this ensures the generated files exist in the `bazel-bin` directories):
    ```bash
    # From the repo root:
    (cd tests/go && bazel build //...)
    (cd tests/kotlin && bazel build //...)
    ```
2.  **Run the synchronization script** from the repository root:
    ```bash
    go run tools/update_readme.go
    ```
    This script will read the generated files from the `bazel-bin` directories and inject them into the README between the `<!-- GENERATED_..._START -->` and `<!-- GENERATED_..._END -->` markers.
3.  **Commit the updated README.md**.

---

## Project Structure Reference

For details on the architecture, code generation design, and path resolution logic, see **[DESIGN.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/DESIGN.md)**.
