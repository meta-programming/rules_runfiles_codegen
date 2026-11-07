//! Unit tests for the `runfile` crate, using a map-backed resolver.

use std::collections::HashMap;
use std::ffi::OsString;
use std::path::PathBuf;

use runfile::{DirectorySpec, Error, ExecutableSpec, FileSetSpec, FileSpec, Resolver};

/// Resolves rlocation paths from a fixed map.
struct MapResolver {
    paths: HashMap<&'static str, PathBuf>,
    env: Vec<(OsString, OsString)>,
}

impl Resolver for MapResolver {
    fn rlocation(&self, rlocation_path: &str) -> Option<PathBuf> {
        self.paths.get(rlocation_path).cloned()
    }

    fn env_vars(&self) -> Vec<(OsString, OsString)> {
        self.env.clone()
    }
}

/// Creates a scratch directory with `a.txt`, `b.txt` and `dir/c.txt`, and a
/// resolver that maps `ws/...` rlocation paths to them.
fn fixture(name: &str) -> (PathBuf, MapResolver) {
    let root = PathBuf::from(std::env::var_os("TEST_TMPDIR").expect("TEST_TMPDIR")).join(name);
    std::fs::create_dir_all(root.join("dir")).unwrap();
    std::fs::write(root.join("a.txt"), "a").unwrap();
    std::fs::write(root.join("b.txt"), "b").unwrap();
    std::fs::write(root.join("dir/c.txt"), "c").unwrap();
    let paths = HashMap::from([
        ("ws/a.txt", root.join("a.txt")),
        ("ws/b.txt", root.join("b.txt")),
        ("ws/dir", root.join("dir")),
        // Known to the resolver but absent on disk.
        ("ws/gone.txt", root.join("gone.txt")),
    ]);
    let env = vec![(OsString::from("RUNFILES_DIR"), OsString::from("/runfiles"))];
    (root, MapResolver { paths, env })
}

#[test]
fn file_resolves() {
    let (root, resolver) = fixture("file_resolves");
    let file = FileSpec::new("ws/a.txt").resolve_with(&resolver).unwrap();
    assert_eq!(file.path(), root.join("a.txt"));
    assert_eq!(file.rlocation_path(), "ws/a.txt");
}

#[test]
fn file_unknown_to_resolver_is_not_found() {
    let (_, resolver) = fixture("file_unknown");
    let err = FileSpec::new("ws/nope.txt")
        .resolve_with(&resolver)
        .unwrap_err();
    assert!(
        matches!(&err, Error::NotFound { rlocation_path } if rlocation_path == "ws/nope.txt"),
        "{err:?}"
    );
    assert_eq!(err.to_string(), r#"runfile "ws/nope.txt" not found"#);
}

#[test]
fn file_missing_on_disk_is_not_found() {
    let (_, resolver) = fixture("file_missing");
    let err = FileSpec::new("ws/gone.txt")
        .resolve_with(&resolver)
        .unwrap_err();
    assert!(matches!(err, Error::NotFound { .. }), "{err:?}");
}

#[test]
fn executable_command_carries_resolver_env() {
    let (root, resolver) = fixture("executable_env");
    let exe = ExecutableSpec::new("ws/a.txt")
        .resolve_with(&resolver)
        .unwrap();
    let command = exe.command();
    assert_eq!(command.get_program(), root.join("a.txt").as_os_str());
    let envs: Vec<_> = command.get_envs().collect();
    assert_eq!(
        envs,
        vec![(
            OsString::from("RUNFILES_DIR").as_os_str(),
            Some(OsString::from("/runfiles").as_os_str())
        )]
    );
}

#[test]
fn directory_child_joins_paths() {
    let (root, resolver) = fixture("directory_child");
    let dir = DirectorySpec::new("ws/dir")
        .resolve_with(&resolver)
        .unwrap();
    let child = dir.child("c.txt");
    assert_eq!(child.path(), root.join("dir/c.txt"));
    assert_eq!(child.rlocation_path(), "ws/dir/c.txt");
    assert_eq!(std::fs::read_to_string(child.path()).unwrap(), "c");
}

static SET: FileSetSpec = runfile::file_set! {
    "b.txt" => "ws/b.txt",
    "a.txt" => "ws/a.txt",
};

#[test]
fn file_set_resolves_every_file() {
    let (root, resolver) = fixture("file_set");
    let set = SET.resolve_with(&resolver).unwrap();
    assert_eq!(set.rel_paths().collect::<Vec<_>>(), ["a.txt", "b.txt"]);
    assert_eq!(set.file("b.txt").unwrap().path(), root.join("b.txt"));
    let err = set.file("c.txt").unwrap_err();
    assert!(
        matches!(&err, Error::NotInFileSet { rel_path } if rel_path == "c.txt"),
        "{err:?}"
    );
}

#[test]
fn file_set_fails_if_any_file_is_missing() {
    let (_, resolver) = fixture("file_set_missing");
    static PARTIAL: FileSetSpec = runfile::file_set! {
        "a.txt" => "ws/a.txt",
        "gone.txt" => "ws/gone.txt",
    };
    let err = PARTIAL.resolve_with(&resolver).unwrap_err();
    assert!(
        matches!(&err, Error::NotFound { rlocation_path } if rlocation_path == "ws/gone.txt"),
        "{err:?}"
    );
}

#[test]
#[should_panic(expected = "runfile \"ws/nope.txt\" not found")]
fn must_resolve_panics_with_the_error() {
    // The test binary has runfiles, so the default resolver initializes and
    // the lookup itself fails.
    FileSpec::new("ws/nope.txt").must_resolve();
}
