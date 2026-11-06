//! Checks that the source generated for `:test_resources` is formatted as
//! `rustfmt` would format it.

use format_check_resources::{GENERATED_SOURCE, RUSTFMT};

#[test]
fn generated_source_is_rustfmt_clean() {
    let source = GENERATED_SOURCE.must_resolve();
    let output = RUSTFMT
        .must_resolve()
        .command()
        .args(["--check", "--edition", "2021"])
        .arg(source.path())
        .output()
        .expect("cannot run rustfmt");
    assert!(
        output.status.success(),
        "rustfmt --check failed with {}\nstdout:\n{}\nstderr:\n{}",
        output.status,
        String::from_utf8_lossy(&output.stdout),
        String::from_utf8_lossy(&output.stderr)
    );
}
