load("@bazel_skylib//lib:unittest.bzl", "asserts", "unittest")
load("//internal/emitters:rust.bzl", "emit_rust")

def _entry(name, runfile_path, doc = "", is_executable = False, is_directory = False, collection_files = None):
    return {
        "name": name,
        "runfile_path": runfile_path,
        "is_executable": is_executable,
        "is_directory": is_directory,
        "is_collection": collection_files != None,
        "collection_files": collection_files or {},
        "doc": doc,
        "label": "//pkg:target",
    }

def _body(entries):
    """Returns the emitted source after the crate docs."""
    code = emit_rust(entries, "//pkg:lib")
    return code[code.index("\n\n///") + 2:]

def _spec_layout_test_impl(ctx):
    env = unittest.begin(ctx)

    # Each expected layout below was checked with `rustfmt --check`.
    asserts.equals(
        env,
        """/// A file.
///
/// Source: `//pkg:target`
pub static A: runfile::FileSpec = runfile::FileSpec::new("_main/data/dummy.txt");
""",
        _body([_entry("A", "_main/data/dummy.txt", doc = "A file.")]),
    )

    # Too long for one line: break after `=`.
    asserts.equals(
        env,
        """/// Runfile for `//pkg:target`.
///
/// Source: `//pkg:target`
pub static HELPER_TOOL_BINARY: runfile::ExecutableSpec =
    runfile::ExecutableSpec::new("_main/tools/helper/helper_tool");
""",
        _body([_entry("HELPER_TOOL_BINARY", "_main/tools/helper/helper_tool", is_executable = True)]),
    )

    # Too long after `=`: the argument gets its own line.
    long_path = "_main/" + "a" * 80 + ".txt"
    asserts.equals(
        env,
        """/// Runfile for `//pkg:target`.
///
/// Source: `//pkg:target`
pub static DIR: runfile::DirectorySpec = runfile::DirectorySpec::new(
    "%s",
);
""" % long_path,
        _body([_entry("DIR", long_path, is_directory = True)]),
    )

    # Too long anywhere: rustfmt leaves the line as written.
    longer_path = "_main/" + "a" * 100 + ".txt"
    asserts.equals(
        env,
        """/// Runfile for `//pkg:target`.
///
/// Source: `//pkg:target`
pub static F: runfile::FileSpec = runfile::FileSpec::new("%s");
""" % longer_path,
        _body([_entry("F", longer_path)]),
    )
    return unittest.end(env)

def _file_set_test_impl(ctx):
    env = unittest.begin(ctx)
    asserts.equals(
        env,
        """/// Two files.
///
/// Source: `//pkg:target`
pub static SET: runfile::FileSetSpec = runfile::file_set! {
    "a.txt" => "_main/data/a.txt",
    "b.txt" => "_main/data/b.txt",
};
""",
        _body([_entry("SET", "_main/data", doc = "Two files.", collection_files = {
            "b.txt": "_main/data/b.txt",
            "a.txt": "_main/data/a.txt",
        })]),
    )
    return unittest.end(env)

def _doc_test_impl(ctx):
    env = unittest.begin(ctx)
    asserts.equals(
        env,
        """/// First line.
///
/// Third line with "quotes".
///
/// Source: `//pkg:target`
pub static A: runfile::FileSpec = runfile::FileSpec::new("_main/a\\"b");
""",
        _body([_entry("A", "_main/a\"b", doc = "First line.\n\nThird line with \"quotes\".  \n")]),
    )
    return unittest.end(env)

spec_layout_test = unittest.make(_spec_layout_test_impl)
file_set_test = unittest.make(_file_set_test_impl)
doc_test = unittest.make(_doc_test_impl)

def rust_test_suite(name):
    unittest.suite(
        name,
        spec_layout_test,
        file_set_test,
        doc_test,
    )
