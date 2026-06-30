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

## Developer Tools (`devtool`)

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

## Adding Support for a New Language

To add support for a new language (e.g., Python, C++), you must follow the multi-module structure established in this project. Each language is published as a separate Bazel module to keep dependencies minimal for users.

Follow these steps to implement and integrate a new language:

### 1. Implement the Rules and Generator

1.  **Create the module directory**: Create a new directory at the root (e.g., `repo/python/`).
2.  **Implement the Starlark Emitter**: In the core module (`repo/core/internal/emitters/<lang>.bzl`), write a Starlark function that takes the analyzed runfiles and generates the source code string for the target language.
3.  **Update the Core Rule**: Modify `repo/core/internal/rules.bzl` to load your new emitter and support the new language in the `runfile_codegen` rule.
4.  **Define the Public API**: In `repo/<lang>/defs.bzl`, define the public macros that developers will use (e.g., `<lang_prefix>_runfile_library`). This macro should wrap the core `runfile_codegen` rule and the target language's native library rule.
5.  **Create the MODULE.bazel**: Define the module in `repo/<lang>/MODULE.bazel`. It must depend on `rules_runfile_codegen_core`.

### 2. Create an Example Workspace

Create a standalone Bazel workspace in `repo/examples/<lang>/` to demonstrate usage and provide source material for the documentation.

1.  **Configure Bzlmod**: Use `local_path_override` in `examples/<lang>/MODULE.bazel` to point to your local `<lang>/` and `core/` modules (see the local development setup section above).
2.  **Create a usage example**: Write a simple program that uses the generated library.
3.  **Annotate the BUILD.bazel**: In `examples/<lang>/BUILD.bazel`, wrap the quickstart target definition in `# [START quickstart]` and `# [END quickstart]` comments. The devtool uses these markers to extract the snippet.
4.  **Verify the build**: Ensure that running `bazel build //...` in the example directory succeeds and generates the expected output in `bazel-bin/`.

### 3. Integrate with the Developer Tool (devtool)

To enable automatic documentation updates and version management, you must register the new language with the devtool:

1.  **Register the language**: Open [tools/cmd/devtool/languages.go](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/tools/cmd/devtool/languages.go) and add a new `LanguageConfig` entry to the `languages` slice.
2.  **Add README placeholders**: Open [README.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/README.md) and add the following HTML comment markers in the appropriate sections:
    *   `<!-- <LANG>_INSTALL_START -->` / `<!-- <LANG>_INSTALL_END -->` (for the `MODULE.bazel` dependency snippet)
    *   `<!-- <LANG>_BUILD_START -->` / `<!-- <LANG>_BUILD_END -->` (for the `BUILD.bazel` quickstart snippet)
    *   `<!-- <LANG>_USAGE_START -->` / `<!-- <LANG>_USAGE_END -->` (for the example usage code)
    *   `<!-- GENERATED_<LANG>_START -->` / `<!-- GENERATED_<LANG>_END -->` (for the actual generated code output)
    *   *(Replace `<LANG>` with the uppercase name of the language, e.g., `PYTHON`).*
3.  **Sync the README**: Run `tools/devtool update-readme` from the repository root to verify that the devtool automatically builds your example and injects the code into the README.

### 4. Create Integration Tests

Create a standalone test workspace in `repo/tests/<lang>/` to run integration tests.

1.  **Configure Bzlmod**: Like the example workspace, use `local_path_override` to point to the local modules.
2.  **Write tests**: Implement tests that verify the generated code compiles, runs, and correctly resolves runfiles at runtime.
3.  **Verify**: Ensure `bazel test //...` passes in the test directory.

### 5. Add BCR Release Metadata

Since each language module is published separately to the Bazel Central Registry (BCR), you must set up the release metadata:

1.  **Create BCR metadata**: In `repo/.bcr/modules/`, create a new directory `rules_runfile_codegen_<lang>/`.
2.  **Add configuration**: Create `metadata.json` and `presubmit.yml` in that directory, following the pattern of the existing modules.
3.  **Update release workflow**: If necessary, update the GitHub Actions release workflow in `.github/workflows/release.yml` to include the new module in the release process.

---

## Project Structure Reference

For details on the architecture, code generation design, and path resolution logic, see **[DESIGN.md](file:///usr/local/google/home/reddaly/tcode/runfile-codegen/repo/DESIGN.md)**.
