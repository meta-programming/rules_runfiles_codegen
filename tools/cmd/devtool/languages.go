package main

// LanguageConfig defines the paths and markers for a language module's examples.
type LanguageConfig struct {
	ID         string // "go", "kotlin" (used in paths: examples/<ID>/)
	LangMarker string // "GO", "KOTLIN" (used in README markers: <!-- <MARKER>_INSTALL_START -->)
	Extension  string // "go", "kotlin" (used for markdown code block syntax highlighting)
	UsageFile  string // "main.go", "Main.kt" (relative to examples/<ID>/)
	GenFile    string // "bazel-bin/...go" (relative to examples/<ID>/)
}

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
