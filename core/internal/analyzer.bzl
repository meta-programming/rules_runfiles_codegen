# Target analysis and validation logic for rules_runfile_codegen.

def _get_runfile_path(ctx, file):
    short_path = file.short_path
    if short_path.startswith("../"):
        return short_path[3:]
    else:
        workspace_name = ctx.workspace_name
        if not workspace_name:
            workspace_name = "_main"
        return workspace_name + "/" + short_path

def _longest_common_prefix(paths):
    if not paths:
        return ""
    if len(paths) == 1:
        parts = paths[0].split("/")
        if len(parts) > 1:
            return "/".join(parts[:-1])
        return ""
        
    split_paths = [p.split("/") for p in paths]
    min_len = min([len(p) for p in split_paths])
    
    common_parts = []
    for i in range(min_len - 1): # Don't include the filename
        part = split_paths[0][i]
        match = True
        for p in split_paths:
            if p[i] != part:
                match = False
                break
        if match:
            common_parts.append(part)
        else:
            break
            
    return "/".join(common_parts)

def _resolve_base(ctx, targets, base_attr, full_paths):
    # Library repo (where go_runfile_library is instantiated)
    lib_repo = ctx.label.workspace_name
    if not lib_repo:
        lib_repo = ctx.workspace_name
        if not lib_repo:
            lib_repo = "_main"
    lib_pkg = ctx.label.package

    resolved_base = ""

    if base_attr == "__default__" or base_attr == ".":
        if lib_pkg:
            resolved_base = lib_repo + "/" + lib_pkg
        else:
            resolved_base = lib_repo
            
    elif base_attr.startswith("./"):
        sub = base_attr[2:]
        if lib_pkg:
            resolved_base = lib_repo + "/" + lib_pkg + "/" + sub
        else:
            resolved_base = lib_repo + "/" + sub
            
    elif base_attr == "common_dir":
        return _longest_common_prefix(full_paths)
        
    elif base_attr == "":
        return ""
        
    elif base_attr.startswith("//") or base_attr.startswith("@"):
        clean_base = base_attr
        if clean_base.endswith("/") and not clean_base.endswith("//"):
            clean_base = clean_base[:-1]
            
        if clean_base.endswith("//"):
            label_str = clean_base + ":dummy"
        elif ":" in clean_base:
            label_str = clean_base
        else:
            label_str = clean_base + ":dummy"
            
        lbl = ctx.label.relative(label_str)
        repo = lbl.workspace_name
        if not repo:
            repo = lib_repo
        pkg = lbl.package
        if pkg:
            resolved_base = repo + "/" + pkg
        else:
            resolved_base = repo
    else:
        fail("Invalid base '%s': must start with '//', '@', '.', be 'common_dir', or empty ''." % base_attr)

    if resolved_base.endswith("/"):
        resolved_base = resolved_base[:-1]
        
    return resolved_base

def analyze_entries(ctx, entries_configs):
    """Analyzes and validates runfile entries.

    Args:
        ctx: The Bazel rule context.
        entries_configs: A list of dicts, each representing an entry:
            {
                "name": str,
                "targets": list of Targets,
                "doc": str,
                "base": str,
                "type": str,
            }
    """
    seen_names = {}
    entries = []
    
    for config in entries_configs:
        name = config["name"]
        targets = config["targets"]
        doc = config["doc"]
        base_attr = config["base"]
        type_attr = config["type"]

        if name in seen_names:
            fail("Duplicate runfile name '%s' detected in '%s'." % (name, ctx.label))
        seen_names[name] = True

        files = []
        is_executable = False
        executable_file = None
        
        is_single_target = len(targets) == 1
        
        for target in targets:
            files.extend(target.files.to_list())
            
            if (is_single_target and 
                DefaultInfo in target and 
                target[DefaultInfo].files_to_run and 
                target[DefaultInfo].files_to_run.executable and 
                not target[DefaultInfo].files_to_run.executable.is_source and 
                not target[DefaultInfo].files_to_run.executable.is_directory):
                is_executable = True
                executable_file = target[DefaultInfo].files_to_run.executable

        if len(files) == 0:
            fail("Targets for entry '%s' produce no files." % name)

        actual_is_directory = False
        actual_is_fileset = False
        
        if len(files) == 1:
            if files[0].is_directory:
                actual_is_directory = True
        else:
            actual_is_fileset = True

        resolved_type = type_attr
        if resolved_type == "auto":
            if is_executable:
                resolved_type = "file"
            elif actual_is_directory:
                resolved_type = "directory"
            elif actual_is_fileset or len(targets) > 1:
                resolved_type = "fileset"
            else:
                resolved_type = "file"
        elif resolved_type == "executable":
            if not is_executable:
                fail("Entry '%s' does not produce an executable, but type 'executable' was expected." % name)
            resolved_type = "file"
                
        if resolved_type == "group":
            resolved_type = "fileset"

        is_directory = False
        is_fileset = False
        fileset_files = {}
        runfile_path = ""

        if resolved_type == "file":
            if not is_executable and len(files) > 1:
                fail("Entry '%s' produces multiple files, but type 'file' was expected." % name)
            if actual_is_directory:
                fail("Entry '%s' produces a directory, but type 'file' was expected." % name)
            
            if is_executable:
                runfile_path = _get_runfile_path(ctx, executable_file)
            else:
                runfile_path = _get_runfile_path(ctx, files[0])
                
        elif resolved_type == "directory":
            if not actual_is_directory:
                fail("Entry '%s' does not produce a directory, but type 'directory' was expected." % name)
            is_directory = True
            runfile_path = _get_runfile_path(ctx, files[0])
            
        elif resolved_type == "fileset":
            if actual_is_directory:
                fail("Entry '%s' produces a directory, but type 'fileset' was expected." % name)
            is_fileset = True
            
            full_paths = []
            for f in files:
                full_paths.append(_get_runfile_path(ctx, f))
            
            common_prefix = _resolve_base(ctx, targets, base_attr, full_paths)
            runfile_path = common_prefix
            
            prefix_len = len(common_prefix)
            matched_any = False
            for f in full_paths:
                if f.startswith(common_prefix):
                    matched_any = True
                    rel = f[prefix_len:]
                    if rel.startswith("/"):
                        rel = rel[1:]
                    
                    if rel in fileset_files:
                        fail("Duplicate relative path '%s' in fileset '%s'. Filename collision between targets." % (rel, name))
                    fileset_files[rel] = f
            
            if not matched_any:
                display_prefix = base_attr
                if display_prefix == "__default__":
                    display_prefix = "default (library's package)"
                
                labels_str = ", ".join([str(t.label) for t in targets])
                fail("Targets [%s] produce no files under base %s (resolved: %s)" % (labels_str, display_prefix, common_prefix))
        else:
            fail("Internal error: unknown resolved type %s" % resolved_type)

        entries.append({
            "name": name,
            "runfile_path": runfile_path,
            "is_executable": is_executable and resolved_type == "file",
            "is_directory": is_directory,
            "is_collection": is_fileset,
            "collection_files": fileset_files,
            "doc": doc,
            "label": ", ".join([str(t.label) for t in targets]),
        })

    return entries
