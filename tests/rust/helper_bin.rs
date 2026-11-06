//! Prints `data/helper_data.txt`, found through its own runfiles.
//!
//! The integration test runs this binary as a child process to check that
//! `Executable::command()` passes the runfiles environment on.

use std::process::ExitCode;

fn main() -> ExitCode {
    let runfiles = match runfiles::Runfiles::create() {
        Ok(runfiles) => runfiles,
        Err(err) => {
            eprintln!("helper: cannot find runfiles: {err}");
            return ExitCode::FAILURE;
        }
    };
    let Some(path) = runfiles::rlocation!(runfiles, "_main/data/helper_data.txt") else {
        eprintln!("helper: data/helper_data.txt is not in the runfiles");
        return ExitCode::FAILURE;
    };
    match std::fs::read_to_string(&path) {
        Ok(content) => {
            print!("{content}");
            ExitCode::SUCCESS
        }
        Err(err) => {
            eprintln!("helper: cannot read {}: {err}", path.display());
            ExitCode::FAILURE
        }
    }
}
