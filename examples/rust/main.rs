use std::process::ExitCode;

fn main() -> ExitCode {
    // 1. Resolve a runfile, handling the error.
    let data_file = match resources::DATA_FILE.resolve() {
        Ok(file) => file,
        Err(err) => {
            eprintln!("Error resolving runfile: {err}");
            return ExitCode::FAILURE;
        }
    };
    let content = std::fs::read_to_string(data_file.path()).expect("read data file");
    println!("Data: {content}");

    // 2. Run an executable runfile with the runfiles environment (fail-fast).
    let helper = resources::HELPER_TOOL.must_resolve();
    let output = helper.command().output().expect("run helper");
    print!("Helper output: {}", String::from_utf8_lossy(&output.stdout));

    // 3. Access a set of runfiles by relative path.
    let example_set = resources::EXAMPLE_SET.must_resolve();
    let paths: Vec<_> = example_set.rel_paths().collect();
    println!("FileSet paths: {paths:?}");
    let dummy = example_set.must_file("dummy.txt");
    let dummy_content = std::fs::read_to_string(dummy.path()).expect("read dummy.txt");
    println!("FileSet dummy content: {}", dummy_content.trim());
    ExitCode::SUCCESS
}
