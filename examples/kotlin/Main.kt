package com.example.project.examples

import com.example.project.examples.resources.Resources
import kotlin.io.path.readText

fun main() {
    // 1. Access the resolved runfile path.
    // Resolve the spec and read its content directly using Path.readText().
    val content = Resources.configJson.path.readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    val process = Resources.helperTool.processBuilder().start()
    val output = process.inputStream.reader().use { it.readText() }.trim()
    val exitCode = process.waitFor()
    if (exitCode != 0) {
        error("Helper tool failed with exit code $exitCode")
    }
    println("Helper output: $output")

    // 3. Access a fileset of runfiles (FileSet).
    val exampleSet = Resources.exampleSet.resolve()
    println("FileSet paths: ${exampleSet.relPaths.sorted()}")
    
    // Access a path inside the fileset using the shortcut:
    val dummyContent = Resources.exampleSet["dummy.txt"].path.readText().trim()
    println("FileSet dummy content: $dummyContent")
}
