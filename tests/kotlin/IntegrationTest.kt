package com.example.project.tests

import com.example.project.tests.resources.TestResources
import java.io.File
import java.nio.file.Files

fun main() {
    // Assertions
    val configFile = TestResources.configJson.resolve()
    val configPath = configFile.path
    println("Config path: $configPath")
    val configContent = File(configPath).readText().trim()
    if (configContent != "dummy content") {
        throw RuntimeException("Config content mismatch: got '$configContent', want 'dummy content'")
    }

    val externalFile = TestResources.externalFile.resolve()
    val schemaPath = externalFile.jvmPath
    println("Schema path: $schemaPath")
    val schemaContent = Files.readString(schemaPath).trim()
    if (schemaContent.isEmpty()) {
        throw RuntimeException("External file is empty")
    }
    // rules_kotlin LICENSE should contain Apache or similar, let's just check it's not empty.
    // We can also check for "Apache" if we are sure.
    if (!schemaContent.contains("Apache") && !schemaContent.contains("License") && !schemaContent.contains("Copyright")) {
        throw RuntimeException("External file content doesn't look like a LICENSE: $schemaContent")
    }

    // Executable test
    val helper = TestResources.helperTool.resolve()
    val process = helper.processBuilder().start()
    val stdout = process.inputStream.bufferedReader().readText().trim()
    val stderr = process.errorStream.bufferedReader().readText().trim()
    val exitCode = process.waitFor()

    println("Helper stdout: $stdout")
    println("Helper stderr: $stderr")

    if (exitCode != 0) {
        throw RuntimeException("Helper exited with code $exitCode. Stderr: $stderr")
    }

    if (stdout != "helper data content") {
        throw RuntimeException("Helper output mismatch: got '$stdout', want 'helper data content'")
    }

    println("All tests passed!")
}
