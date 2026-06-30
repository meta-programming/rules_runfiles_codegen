# Target analysis and validation logic for rules_runfile_codegen.

def analyze_entries(ctx, names, targets, docs):
    """Analyzes and validates runfile targets, returning structured entries.

    This function performs build-time validation on the target names (uniqueness),
    checks the single-file constraint, detects if targets are executable (propagating
    the `is_source` fix), and resolves the correct runfile paths.

    Language-specific identifier validation is deferred to the respective emitters
    to keep the core analyzer completely language-agnostic.

    Args:
        ctx: The Bazel rule context (from the calling rule).
        names: A list of strings representing the symbolic names chosen by the user.
        targets: A list of Targets (labels) corresponding to the runfiles.
        docs: A list of strings representing the documentation for each entry.

    Returns:
        A list of dicts, where each dict represents a validated runfile entry:
            {
                "name": str,
                "runfile_path": str (resolved rlocation path),
                "is_executable": bool,
                "doc": str,
                "label": str (target label),
            }
    """
    if len(names) != len(targets):
        fail("names and targets must have the same length")
    if len(names) != len(docs):
        fail("names and docs must have the same length")

    seen_names = {}
    entries = []
    
    for i in range(len(names)):
        name = names[i]
        target = targets[i]
        doc = docs[i]

        # Validate uniqueness
        if name in seen_names:
            fail("Duplicate runfile name '%s' detected in '%s'." % (name, ctx.label))
        seen_names[name] = True

        # Robust executable detection (with is_source fix)
        is_executable = False
        file = None
        if DefaultInfo in target and target[DefaultInfo].files_to_run and target[DefaultInfo].files_to_run.executable and not target[DefaultInfo].files_to_run.executable.is_source:
            is_executable = True
            file = target[DefaultInfo].files_to_run.executable

        if not file:
            files = target.files.to_list()
            if len(files) != 1:
                fail("Target %s produces %d files, but rules_runfile_codegen only supports targets producing exactly one file. Files: %s" % (target.label, len(files), files))
            file = files[0]

        # Determine runfile path
        short_path = file.short_path
        if short_path.startswith("../"):
            runfile_path = short_path[3:]
        else:
            workspace_name = ctx.workspace_name
            if not workspace_name:
                workspace_name = "_main"
            runfile_path = workspace_name + "/" + short_path

        entries.append({
            "name": name,
            "runfile_path": runfile_path,
            "is_executable": is_executable,
            "doc": doc,
            "label": str(target.label),
        })

    return entries
