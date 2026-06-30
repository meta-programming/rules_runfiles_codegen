package main

// LanguageConfig defines the paths and markers for a language module's examples.
type LanguageConfig struct {
	ID         string // "go", "kotlin" (used in paths: examples/<ID>/)
	LangMarker string // "GO", "KOTLIN" (used in README markers: <!-- <MARKER>_INSTALL_START -->)
	Extension  string // "go", "kotlin" (used for markdown code block syntax highlighting)
	UsageFile  string // "main.go", "Main.kt" (relative to examples/<ID>/)
	GenFile    string // "bazel-bin/...go" (relative to examples/<ID>/)
}

// languages lists all supported language modules in the repository.
//
// To add support for a new language (e.g., "python"):
//
// 1. Create the new language module directory (e.g., "python/") with its Starlark rules and generator.
// 2. Create an example project in "examples/python/" that:
//    - Defines a "quickstart" target in its BUILD.bazel, wrapped in "[START quickstart]" and "[END quickstart]" markers.
//    - Contains a basic usage file (e.g., "main.py").
//    - Generates a file via the new rules (e.g., "bazel-bin/codegen_gen.py").
// 3. Add a new LanguageConfig entry to this slice.
// 4. Add the corresponding HTML comment placeholders to README.md:
//    - <!-- PYTHON_INSTALL_START --> / <!-- PYTHON_INSTALL_END -->
//    - <!-- PYTHON_BUILD_START --> / <!-- PYTHON_BUILD_END -->
//    - <!-- PYTHON_USAGE_START --> / <!-- PYTHON_USAGE_END -->
//    - <!-- GENERATED_PYTHON_START --> / <!-- GENERATED_PYTHON_END -->
// 5. Run "tools/devtool update-readme" to automatically build the example and sync the code into the README.
var languages = []LanguageConfig{
	{
		ID:         "go",
		LangMarker: "GO",
		Extension:  "go",
		UsageFile:  "main.go",
		GenFile:    "bazel-bin/resources_codegen_gen.go",
	},
	{
		ID:         "kotlin",
		LangMarker: "KOTLIN",
		Extension:  "kotlin",
		UsageFile:  "Main.kt",
		GenFile:    "bazel-bin/resources_codegen_gen.kt",
	},
}
