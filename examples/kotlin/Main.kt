package com.example.project.examples

import com.example.project.examples.resources.Resources
import java.io.File

fun main() {
    // 1. Access the resolved runfile path.
    // Resolve the spec to a File.
    val configFile = Resources.configJson.resolve()
    val path = configFile.path
    val content = File(path).readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    // Resolve the spec to an Executable.
    val helper = Resources.helperTool.resolve()
    val process = helper.processBuilder().start()
    val output = process.inputStream.bufferedReader().readText().trim()
    process.waitFor()
    println("Helper output: $output")
}
