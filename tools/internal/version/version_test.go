package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "simple version",
			content: `module(
    name = "my_module",
    version = "1.2.3",
)`,
			want: "1.2.3",
		},
		{
			name: "with other attributes",
			content: `module(
    name = "my_module",
    compatibility_level = 1,
    version = "0.1.0-alpha",
)`,
			want: "0.1.0-alpha",
		},
		{
			name: "no version",
			content: `module(
    name = "my_module",
)`,
			wantErr: true,
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: true,
		},
		{
			name: "invalid syntax",
			content: `module(name = `,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		newVersion string
		want       string
		wantErr    bool
	}{
		{
			name: "update version",
			content: `module(
    name = "my_module",
    version = "1.2.3",
)`,
			newVersion: "2.0.0",
			want: `module(
    name = "my_module",
    version = "2.0.0",
)
`,
		},
		{
			name: "preserve formatting and comments",
			content: `# This is a comment
module(
    name = "my_module",
    # Another comment
    version = "1.2.3",
    compatibility_level = 1,
)`,
			newVersion: "0.2.0",
			want: `# This is a comment
module(
    name = "my_module",
    # Another comment
    version = "0.2.0",
    compatibility_level = 1,
)
`,
		},
		{
			name: "no version to update",
			content: `module(
    name = "my_module",
)`,
			newVersion: "1.0.0",
			wantErr:    true,
		},
		{
			name: "update bazel_dep versions",
			content: `module(
    name = "my_module",
    version = "1.2.3",
)

bazel_dep(name = "rules_runfile_codegen_core", version = "1.2.3")
bazel_dep(name = "other_dep", version = "1.2.3")
`,
			newVersion: "2.0.0",
			want: `module(
    name = "my_module",
    version = "2.0.0",
)

bazel_dep(name = "rules_runfile_codegen_core", version = "2.0.0")
bazel_dep(name = "other_dep", version = "1.2.3")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Update([]byte(tt.content), tt.newVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if string(got) != tt.want {
					t.Errorf("Update() = %q, want %q", string(got), tt.want)
				}
			}
		})
	}
}
