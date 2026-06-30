# Core rule definition for rules_runfile_codegen.

load("//internal:analyzer.bzl", "analyze_entries")
load("//internal/emitters:go.bzl", "emit_go")
load("//internal/emitters:kotlin.bzl", "emit_kotlin")

def _runfile_codegen_impl(ctx):
    language = ctx.attr.language
    entries = analyze_entries(ctx, ctx.attr.names, ctx.attr.targets, ctx.attr.docs)

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
        "names": attr.string_list(mandatory = True),
        "targets": attr.label_list(mandatory = True, allow_files = True, providers = [DefaultInfo]),
        "docs": attr.string_list(mandatory = True),
        "object_name": attr.string(), # Optional, only used/required for Kotlin
    },
)
