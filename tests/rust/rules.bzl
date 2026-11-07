def _dir_producer_impl(ctx):
    out_dir = ctx.actions.declare_directory(ctx.label.name)
    ctx.actions.run_shell(
        outputs = [out_dir],
        command = "mkdir -p {dir} && echo 'file1 content' > {dir}/file1.txt && echo 'file2 content' > {dir}/file2.txt".format(dir = out_dir.path),
    )
    return [DefaultInfo(files = depset([out_dir]))]

dir_producer = rule(
    implementation = _dir_producer_impl,
)
