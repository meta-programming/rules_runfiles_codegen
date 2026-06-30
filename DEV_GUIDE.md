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

## Developer Tools (`runfilesdevtool`)

The project includes a unified developer tool to assist with common tasks like updating documentation and managing module versions. A wrapper script is provided at `tools/devtool` so you can run it easily from anywhere in the repository.

To run the tool, use the wrapper script from the repository root:
```bash
tools/devtool [command]
```

### Updating the Documentation (README.md)

The **[README.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/README.md)** contains sections showing the "Actual Generated Code" for both Go and Kotlin. To ensure these snippets are 100% accurate and compiler-verified, they are automatically synchronized from the integration test outputs.

If you make changes to the code generators, you should update the README before committing.

Run the update-readme command from the repository root:
```bash
tools/devtool update-readme
```
This will automatically build the examples for all languages to generate the latest outputs, read them, and inject them into the README.

If you want to skip the build step (e.g., if you already built them and want it to run faster), you can pass the `--build=false` flag:
```bash
tools/devtool update-readme --build=false
```

### Managing Module Versions

During a release, all three released modules (`core`, `go`, `kotlin`) must share the exact same version.

*   **Check version consistency**:
    ```bash
    tools/devtool version check
    ```
    This verifies that the core, go, and kotlin modules have matching versions.
*   **Set a new version**:
    ```bash
    tools/devtool version set 0.2.0
    ```
    This uses the official Bazel AST parser to safely update the version in the core, go, and kotlin MODULE.bazel files at once, preserving formatting and comments.



---

## Project Structure Reference

For details on the architecture, code generation design, and path resolution logic, see **[DESIGN.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/DESIGN.md)**.
