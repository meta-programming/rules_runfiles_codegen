package com.example.project.tests

import com.google.devtools.build.runfiles.Runfiles
import java.io.File

fun main() {
    val runfiles = Runfiles.create()
    val path = runfiles.rlocation("rules_runfile_codegen_kotlin_tests/data/helper_data.txt")
        ?: throw RuntimeException("Helper: Failed to resolve runfile")
    
    val content = File(path).readText()
    print(content)
}
