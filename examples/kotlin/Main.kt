package com.example.project.examples

import com.example.project.examples.resources.Resources

fun main() {
    // 1. Access the resolved runfile path.
    // Resolve the spec and read its content directly using the helper property.
    val content = Resources.configJson.resolve().file.readText().trim()
    println("Data: $content")

    // 2. Run an executable runfile with env propagation.
    // Resolve, start, and read the output in a fluent chain.
    val output = Resources.helperTool.resolve().processBuilder().start()
        .inputStream.reader().readText().trim()
    println("Helper output: $output")
}
