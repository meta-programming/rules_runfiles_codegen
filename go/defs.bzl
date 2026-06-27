load("@rules_runfiles_codegen//:codegen.bzl", "generate_go_source")
load("@rules_go//go:def.bzl", "go_library")

def go_runfile(name, target, doc = ""):
    """Creates a runfile entry definition."""
    return struct(
        name = name,
        target = target,
        doc = doc,
    )

def _go_runfile_codegen_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".go")
    
    entries = []
    runfiles_deps = []
    
    for i in range(len(ctx.attr.targets)):
        target = ctx.attr.targets[i]
        meta_str = ctx.attr.metadata[i]
        
        # Since json.decode might not be available in all Starlark environments,
        # we can parse a simple CSV or just use structural attributes.
        # But json.decode is available in Bazel.
        meta = json.decode(meta_str)
        
        # Get the primary file
        files = target[DefaultInfo].files.to_list()
        if not files:
            fail("Target {} does not provide any files".format(target.label))
        
        file = files[0]
        
        # Construct runfiles path
        # In Bazel, the runfiles path for a file is generally its short_path.
        # But if it's from an external repository, short_path starts with `../repo_name/`.
        # The runfiles tree layout: `repo_name/...`.
        if file.short_path.startswith("../"):
            runfiles_path = file.short_path[3:]
        else:
            runfiles_path = ctx.workspace_name + "/" + file.short_path
            
        entries.append(struct(
            name = meta["name"],
            doc = meta["doc"],
            runfiles_path = runfiles_path,
        ))
        
        if target[DefaultInfo].default_runfiles:
            runfiles_deps.append(target[DefaultInfo].default_runfiles)
            
    src_content = generate_go_source(ctx.attr.importpath, entries)
    ctx.actions.write(out, src_content)
    
    merged_runfiles = ctx.runfiles().merge_all(runfiles_deps)
    
    return [DefaultInfo(
        files = depset([out]),
        runfiles = merged_runfiles,
    )]

_go_runfile_codegen = rule(
    implementation = _go_runfile_codegen_impl,
    attrs = {
        "importpath": attr.string(mandatory = True),
        "targets": attr.label_list(allow_files = True, mandatory = True),
        "metadata": attr.string_list(mandatory = True),
    },
)

def go_runfile_library(name, importpath, entries, **kwargs):
    """Generates a go_library that provides access to the specified runfiles.
    
    Args:
        name: Name of the target.
        importpath: The Go importpath of the generated library.
        entries: A list of `go_runfile` structs.
        **kwargs: Additional arguments to pass to go_library.
    """
    targets = [e.target for e in entries]
    metadata = [json.encode({"name": e.name, "doc": e.doc}) for e in entries]
    
    codegen_name = name + "_codegen"
    
    _go_runfile_codegen(
        name = codegen_name,
        importpath = importpath,
        targets = targets,
        metadata = metadata,
    )
    
    go_library(
        name = name,
        srcs = [codegen_name],
        importpath = importpath,
        data = targets, # Needed to ensure they are built and available as runfiles
        deps = ["@rules_go//go/runfiles:go_default_library"],
        **kwargs
    )
