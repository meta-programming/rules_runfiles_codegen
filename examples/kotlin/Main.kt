package com.example.project.examples

import com.example.project.examples.resources.Resources
import kotlin.io.path.readText

fun main() {
    // 1. Access the resolved runfile path.
    // Resolve the spec and read its content directly using Path.readText().
    val content = Resources.configJson.resolve().path.readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    val process = Resources.helperTool.resolve().processBuilder().start()
    val output = process.inputStream.reader().use { it.readText() }.trim()
    val exitCode = process.waitFor()
    if (exitCode != 0) {
        error("Helper tool failed with exit code $exitCode")
    }
    println("Helper output: $output")

    // 3. Access a fileset of runfiles (FileSet).
    val exampleSet = Resources.exampleSet.resolve()
    println("FileSet paths: ${exampleSet.relPaths.sorted()}")
    val f1 = exampleSet["dummy.txt"]
    println("FileSet dummy content: ${f1.path.readText().trim()}")
}
