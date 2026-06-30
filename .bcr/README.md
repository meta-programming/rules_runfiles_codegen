# Bazel Central Registry Integration

This directory contains the metadata and configuration required to publish the modules in this repository to the [Bazel Central Registry (BCR)](https://bazel.build/external/registry). 

We use the [bazel-contrib/publish-to-bcr](https://github.com/bazel-contrib/publish-to-bcr) GitHub Action to automate the release process. When a new release is published, our [release workflow](../.github/workflows/release.yml) packages our three modules—[core](../core), [go](../go), and [kotlin](../kotlin)—and submits them to the registry. This automation relies on the configuration in [config.yml](config.yml) and the templates in the [core](core/), [go](go/), and [kotlin](kotlin/) directories, which conform to the [BCR contribution guidelines](https://github.com/bazelbuild/bazel-central-registry/blob/main/CONTRIBUTING.md).

Releases follow the project [versioning policy](../VERSIONING.md), which adopts a unified independent semantic versioning scheme. All modules are released together under a shared version number, ensuring compatibility across the suite.

---

## One-time Maintainer Setup

To enable the automatic publishing workflow, the following setup is required on GitHub:

1.  **Fork the registry**: Fork the [bazelbuild/bazel-central-registry](https://github.com/bazelbuild/bazel-central-registry) repository to your personal `gonzojive` account. The workflow is configured to use `gonzojive/bazel-central-registry` as the staging area.
2.  **Create a personal access token**: Generate a fine-grained GitHub personal access token (PAT) with write access restricted only to your `bazel-central-registry` fork repository.
3.  **Configure repository secret**: Add the generated token as a repository secret named `BCR_PUBLISH_TOKEN` in the `rules_runfiles_codegen` repository settings (`Settings -> Secrets and variables -> Actions`).
4.  **Submit pull requests**: Because fine-grained PATs cannot automatically create cross-repository pull requests, the workflow has `open_pull_request` set to `false`. It will push to your fork and output a URL in the workflow logs. You must click this URL to manually open the pull request against the upstream registry.

---

## Source Templates (`source.template.json`)

To support publishing multiple Bazel modules from this single repository, each module (`core`, `go`, `kotlin`) has a `source.template.json` file in its corresponding `.bcr` subdirectory (e.g., [go/source.template.json](go/source.template.json)).

### Template Structure:
*   **`url`**: The URL of the GitHub release archive to download for a given release tag.
*   **`strip_prefix`**: Specifies the subdirectory within the release archive that contains the module's code. This allows Bazel to isolate `core`, `go`, and `kotlin` as independent modules even though they share the same release archive.

During a release, the GHA workflow automatically replaces `{TAG}` and `{VERSION}` with the actual release details, computes the SHA-256 integrity hash, and generates the final `source.json` for the BCR.
