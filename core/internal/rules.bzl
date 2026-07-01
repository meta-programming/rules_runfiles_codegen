# Core rule definition for rules_runfile_codegen.

load("//internal:analyzer.bzl", "analyze_entries")
load("//internal/emitters:go.bzl", "emit_go")
load("//internal/emitters:kotlin.bzl", "emit_kotlin")

def _runfile_codegen_impl(ctx):
    language = ctx.attr.language
    config_data = json.decode(ctx.attr.config)
    resolved_targets = ctx.attr.targets

    # Reconstruct the list of entry configs with resolved targets
    entries_configs = []
    
    # We want to maintain order. Dict keys preserve insertion order in Starlark.
    for name, entry in config_data.items():
        # Map target indexes back to resolved Target objects
        entry_targets = [resolved_targets[idx] for idx in entry["target_indexes"]]
        entries_configs.append({
            "name": name,
            "targets": entry_targets,
            "doc": entry["doc"],
            "base": entry["base"],
            "type": entry["type"],
        })

    entries = analyze_entries(ctx, entries_configs)

    if language == "go":
        ext = "go"
        code = emit_go(ctx.attr.package, entries)
    elif language == "kotlin":
        ext = "kt"
        code = emit_kotlin(ctx.attr.package, entries, ctx.attr.object_name, str(ctx.label))
    else:
        fail("Unsupported language: %s" % language)

    out_file = ctx.actions.declare_file(ctx.label.name + "_gen." + ext)
    ctx.actions.write(
        output = out_file,
        content = code,
    )

    return [DefaultInfo(files = depset([out_file]))]

runfile_codegen = rule(
    implementation = _runfile_codegen_impl,
    attrs = {
        "package": attr.string(mandatory = True),
        "language": attr.string(mandatory = True, values = ["go", "kotlin"]),
        "targets": attr.label_list(mandatory = True, allow_files = True, providers = [DefaultInfo]),
        "config": attr.string(mandatory = True), # Serialized JSON config
        "object_name": attr.string(), # Optional, only used/required for Kotlin
    },
)
