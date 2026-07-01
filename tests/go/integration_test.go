package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/example/project/repo/tests/go/test_resources"
)

func TestSingleFile(t *testing.T) {
	resolved, err := test_resources.SingleFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve SingleFile: %v", err)
	}

	path := resolved.Path()
	if path == "" {
		t.Fatal("SingleFile path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read SingleFile: %v", err)
	}

	expected := "dummy content"
	got := strings.TrimSpace(string(content))
	if got != expected {
		t.Errorf("SingleFile content = %q, want %q", got, expected)
	}

	if resolved.RlocationPath() != "_main/data/dummy.txt" {
		t.Errorf("Unexpected rlocation path: %q", resolved.RlocationPath())
	}
}

func TestExecutableFile(t *testing.T) {
	resolved, err := test_resources.ExecutableFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve ExecutableFile: %v", err)
	}

	cmd := resolved.Cmd()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run ExecutableFile: %v\nStderr: %s", err, stderr.String())
	}

	expected := "helper data content"
	got := strings.TrimSpace(stdout.String())
	if got != expected {
		t.Errorf("HelperTool output = %q, want %q\nStderr: %s", got, expected, stderr.String())
	}
}

func TestExternalFile(t *testing.T) {
	resolved, err := test_resources.ExternalFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve ExternalFile: %v", err)
	}

	path := resolved.Path()
	if path == "" {
		t.Fatal("ExternalFile path is empty")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read ExternalFile: %v", err)
	}

	if len(content) == 0 {
		t.Error("ExternalFile is empty")
	}

	if !strings.Contains(string(content), "Apache License") {
		t.Errorf("ExternalFile content does not contain 'Apache License'. Got: %s", string(content[:100]))
	}
}

func TestGroupData(t *testing.T) {
	fileset, err := test_resources.GroupData.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve GroupData: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	// Resolve individual files
	f1, err := fileset.File("file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "content of file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "content of file 1")
	}

	f2, err := fileset.File("file2.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file2.txt: %v", err)
	}
	content2, err := os.ReadFile(f2.Path())
	if err != nil {
		t.Fatalf("Failed to read file2.txt: %v", err)
	}
	if strings.TrimSpace(string(content2)) != "content of file 2" {
		t.Errorf("file2.txt content = %q, want %q", string(content2), "content of file 2")
	}
    
	// Test resolve non-existent file
	_, err = fileset.File("non-existent.txt")
	if err == nil {
		t.Error("Expected error resolving non-existent.txt, got nil")
	}
}

func TestGroupDataDefault(t *testing.T) {
	fileset, err := test_resources.GroupDataDefault.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve GroupDataDefault: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"data/collection/file1.txt", "data/collection/file2.txt"}
	
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("data/collection/file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "content of file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "content of file 1")
	}
}

func TestDirectoryData(t *testing.T) {
	dir, err := test_resources.DirData.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DirData: %v", err)
	}

	path := dir.Path()
	if path == "" {
		t.Fatal("DirData path is empty")
	}

	// Resolve files inside directory using Child method
	f1 := dir.Child("file1.txt")
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "file1 content" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "file1 content")
	}

	f2 := dir.Child("file2.txt")
	content2, err := os.ReadFile(f2.Path())
	if err != nil {
		t.Fatalf("Failed to read file2.txt: %v", err)
	}
	if strings.TrimSpace(string(content2)) != "file2 content" {
		t.Errorf("file2.txt content = %q, want %q", string(content2), "file2 content")
	}
}

func TestWrappedDir(t *testing.T) {
	dir, err := test_resources.WrappedDir.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve WrappedDir: %v", err)
	}

	path := dir.Path()
	if path == "" {
		t.Fatal("WrappedDir path is empty")
	}

	f1 := dir.Child("file1.txt")
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "file1 content" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "file1 content")
	}
}

func TestMixedGroup(t *testing.T) {
	fileset, err := test_resources.MixedGroup.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve MixedGroup: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"data/dummy.txt", "dir_data"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	dirFile, err := fileset.File("dir_data")
	if err != nil {
		t.Fatalf("Failed to resolve dir_data: %v", err)
	}
	
	fi, err := os.Stat(dirFile.Path())
	if err != nil {
		t.Fatalf("Failed to stat dir_data: %v", err)
	}
	if !fi.IsDir() {
		t.Errorf("Expected dir_data to be a directory, got file")
	}

	f1Path := filepath.Join(dirFile.Path(), "file1.txt")
	content1, err := os.ReadFile(f1Path)
	if err != nil {
		t.Fatalf("Failed to read file1.txt from dir_data: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "file1 content" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "file1 content")
	}
}

func TestGenData(t *testing.T) {
	resolved, err := test_resources.GenData.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve GenData: %v", err)
	}
	content, err := os.ReadFile(resolved.Path())
	if err != nil {
		t.Fatalf("Failed to read GenData: %v", err)
	}
	if strings.TrimSpace(string(content)) != "generated" {
		t.Errorf("GenData content = %q, want %q", string(content), "generated")
	}
}

func TestDistantGroup(t *testing.T) {
	fileset, err := test_resources.DistantGroup.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DistantGroup: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	// Defaults to target's package ("_main").
	// - _main/data/dummy.txt matches -> "data/dummy.txt"
	// - rules_go/LICENSE.txt does not match -> filtered out
	expectedPaths := []string{"data/dummy.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("data/dummy.txt")
	if err != nil {
		t.Fatalf("Failed to resolve data/dummy.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read dummy.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "dummy content" {
		t.Errorf("dummy.txt content = %q, want %q", string(content1), "dummy content")
	}
}

func TestSubGroupPackageRelative(t *testing.T) {
	fileset, err := test_resources.SubGroupPackageRelative.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve SubGroupPackageRelative: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "sub file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "sub file 1")
	}
}

func TestSubGroupRepoRelative(t *testing.T) {
	fileset, err := test_resources.SubGroupRepoRelative.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve SubGroupRepoRelative: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "sub file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "sub file 1")
	}
}

func TestDistantGroupFiltered(t *testing.T) {
	fileset, err := test_resources.DistantGroupFiltered.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DistantGroupFiltered: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"dummy.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("dummy.txt")
	if err != nil {
		t.Fatalf("Failed to resolve dummy.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read dummy.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "dummy content" {
		t.Errorf("dummy.txt content = %q, want %q", string(content1), "dummy content")
	}

	_, err = fileset.File("rules_go+/LICENSE.txt")
	if err == nil {
		t.Errorf("Expected resolution of filtered out file to fail, but it succeeded")
	}
}

func TestDistantGroupUnfiltered(t *testing.T) {
	fileset, err := test_resources.DistantGroupUnfiltered.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DistantGroupUnfiltered: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"_main/data/dummy.txt", "rules_go+/LICENSE.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("_main/data/dummy.txt")
	if err != nil {
		t.Fatalf("Failed to resolve _main/data/dummy.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read dummy.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "dummy content" {
		t.Errorf("dummy.txt content = %q, want %q", string(content1), "dummy content")
	}

	f2, err := fileset.File("rules_go+/LICENSE.txt")
	if err != nil {
		t.Fatalf("Failed to resolve rules_go+/LICENSE.txt: %v", err)
	}
	content2, err := os.ReadFile(f2.Path())
	if err != nil {
		t.Fatalf("Failed to read LICENSE.txt: %v", err)
	}
	if !strings.Contains(string(content2), "Apache License") {
		t.Errorf("LICENSE.txt content does not contain 'Apache License'")
	}
}

func TestMockRepoGroup(t *testing.T) {
	fileset, err := test_resources.MockRepoGroup.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve MockRepoGroup: %v", err)
	}

	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("FileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	for i, p := range paths {
		if p != expectedPaths[i] {
			t.Errorf("FileSet path[%d] = %q, want %q", i, p, expectedPaths[i])
		}
	}

	f1, err := fileset.File("file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "mock file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "mock file 1")
	}

	f2, err := fileset.File("file2.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file2.txt: %v", err)
	}
	content2, err := os.ReadFile(f2.Path())
	if err != nil {
		t.Fatalf("Failed to read file2.txt: %v", err)
	}
	if strings.TrimSpace(string(content2)) != "mock file 2" {
		t.Errorf("file2.txt content = %q, want %q", string(content2), "mock file 2")
	}
}

func TestStrictFile(t *testing.T) {
	file, err := test_resources.StrictFile.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve StrictFile: %v", err)
	}
	content, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatalf("Failed to read StrictFile: %v", err)
	}
	if strings.TrimSpace(string(content)) != "dummy content" {
		t.Errorf("StrictFile content = %q, want %q", string(content), "dummy content")
	}
}

func TestStrictDir(t *testing.T) {
	dir, err := test_resources.StrictDir.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve StrictDir: %v", err)
	}
	fi, err := os.Stat(dir.Path())
	if err != nil {
		t.Fatalf("Failed to stat StrictDir: %v", err)
	}
	if !fi.IsDir() {
		t.Errorf("Expected StrictDir to be a directory")
	}
}

func TestStrictFileSet(t *testing.T) {
	fileset, err := test_resources.StrictFileSet.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve StrictFileSet: %v", err)
	}
	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("StrictFileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
}

func TestForcedFileSet(t *testing.T) {
	fileset, err := test_resources.ForcedFileSet.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve ForcedFileSet: %v", err)
	}
	paths := fileset.RelPaths()
	expectedPaths := []string{"data/dummy.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("ForcedFileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	
	f1, err := fileset.File("data/dummy.txt")
	if err != nil {
		t.Fatalf("Failed to resolve dummy.txt from ForcedFileSet: %v", err)
	}
	content, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read dummy.txt: %v", err)
	}
	if strings.TrimSpace(string(content)) != "dummy content" {
		t.Errorf("dummy.txt content = %q, want %q", string(content), "dummy content")
	}
}

func TestCommonDirFileSet(t *testing.T) {
	fileset, err := test_resources.CommonDirFileSet.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve CommonDirFileSet: %v", err)
	}
	paths := fileset.RelPaths()
	sort.Strings(paths)
	expectedPaths := []string{"file1.txt", "file2.txt"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("CommonDirFileSet has %d paths, want %d. Got: %v", len(paths), len(expectedPaths), paths)
	}
	
	f1, err := fileset.File("file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve file1.txt: %v", err)
	}
	content1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read file1.txt: %v", err)
	}
	if strings.TrimSpace(string(content1)) != "content of file 1" {
		t.Errorf("file1.txt content = %q, want %q", string(content1), "content of file 1")
	}
}

func TestDuplicateTargets(t *testing.T) {
	// Verify DuplicateTarget1 (stripped)
	fs1, err := test_resources.DuplicateTarget1.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DuplicateTarget1: %v", err)
	}
	paths1 := fs1.RelPaths()
	sort.Strings(paths1)
	if !reflectEqual(paths1, []string{"file1.txt", "file2.txt"}) {
		t.Errorf("DuplicateTarget1 paths = %v, want [file1.txt, file2.txt]", paths1)
	}

	// Verify DuplicateTarget2 (unstripped)
	fs2, err := test_resources.DuplicateTarget2.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve DuplicateTarget2: %v", err)
	}
	paths2 := fs2.RelPaths()
	sort.Strings(paths2)
	expected := []string{"_main/data/collection/file1.txt", "_main/data/collection/file2.txt"}
	if !reflectEqual(paths2, expected) {
		t.Errorf("DuplicateTarget2 paths = %v, want %v", paths2, expected)
	}
}

func TestMixedFileSet(t *testing.T) {
	fs, err := test_resources.MixedFileSet.Resolve()
	if err != nil {
		t.Fatalf("Failed to resolve MixedFileSet: %v", err)
	}
	paths := fs.RelPaths()
	sort.Strings(paths)
	expected := []string{
		"data/collection/file1.txt",
		"data/collection/file2.txt",
		"subdir/sub_data/file1.txt",
		"subdir/sub_data/file2.txt",
	}
	// Common prefix for these in repo is "_main" (resolved from Bzlmod workspace name).
	// Under Bzlmod, paths start with "_main/..."
	// Wait, the files produced are:
	// - _main/data/collection/file1.txt
	// - _main/subdir/sub_data/file1.txt
	// Longest common prefix is "_main/".
	// So stripped paths should be:
	// - data/collection/file1.txt
	// - subdir/sub_data/file1.txt
	// Let's verify.
	if !reflectEqual(paths, expected) {
		t.Errorf("MixedFileSet paths = %v, want %v", paths, expected)
	}

	// Resolve a file from local target
	f1, err := fs.File("data/collection/file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve local file in MixedFileSet: %v", err)
	}
	c1, err := os.ReadFile(f1.Path())
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if strings.TrimSpace(string(c1)) != "content of file 1" {
		t.Errorf("content mismatch: got %q", string(c1))
	}

	// Resolve a file from subdir target
	sf1, err := fs.File("subdir/sub_data/file1.txt")
	if err != nil {
		t.Fatalf("Failed to resolve subdir file in MixedFileSet: %v", err)
	}
	sc1, err := os.ReadFile(sf1.Path())
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if strings.TrimSpace(string(sc1)) != "sub file 1" {
		t.Errorf("content mismatch: got %q", string(sc1))
	}
}

// Helper to compare slices since reflect.DeepEqual can be strict on nil vs empty
func reflectEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
