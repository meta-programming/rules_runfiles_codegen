package com.example.project.examples

import com.example.project.examples.resources.Resources
import java.io.File

fun main() {
    // 1. Access the resolved runfile path.
    // Resources are resolved at startup (init-time).
    val path = Resources.configJson.path
    val content = File(path).readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    val process = Resources.helperTool.processBuilder().start()
    val output = process.inputStream.bufferedReader().readText().trim()
    process.waitFor()
    println("Helper output: $output")
}
