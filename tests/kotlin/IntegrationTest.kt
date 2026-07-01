package com.example.project.tests

import com.example.project.tests.resources.TestResources
import kotlin.io.path.readText

fun main() {
    // Assertions
    val configFile = TestResources.configJson.resolve()
    println("Config path: ${configFile.path}")
    val configContent = configFile.path.readText().trim()
    if (configContent != "dummy content") {
        throw RuntimeException("Config content mismatch: got '$configContent', want 'dummy content'")
    }

    val externalFile = TestResources.externalFile.resolve()
    println("Schema path: ${externalFile.path}")
    val schemaContent = externalFile.path.readText().trim()
    if (schemaContent.isEmpty()) {
        throw RuntimeException("External file is empty")
    }
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

    // FileSet test
    val fileset = TestResources.groupData.resolve()
    val paths = fileset.relPaths.sorted()
    val expectedPaths = listOf("file1.txt", "file2.txt")
    if (paths != expectedPaths) {
        throw RuntimeException("FileSet paths mismatch: got $paths, want $expectedPaths")
    }
    
    val f1 = fileset.resolveFile("file1.txt")
    val content1 = f1.path.readText().trim()
    if (content1 != "content of file 1 kt") {
        throw RuntimeException("file1.txt content mismatch: got '$content1', want 'content of file 1 kt'")
    }

    val f2 = fileset.resolveFile("file2.txt")
    val content2 = f2.path.readText().trim()
    if (content2 != "content of file 2 kt") {
        throw RuntimeException("file2.txt content mismatch: got '$content2', want 'content of file 2 kt'")
    }

    try {
        fileset.resolveFile("non-existent.txt")
        throw RuntimeException("Expected exception when resolving non-existent.txt, but it succeeded")
    } catch (e: Exception) {
        // Expected
    }

    // FileSet default test
    val defaultFileSet = TestResources.groupDataDefault.resolve()
    val defaultPaths = defaultFileSet.relPaths.sorted()
    val expectedDefaultPaths = listOf("data/collection/file1.txt", "data/collection/file2.txt")
    if (defaultPaths != expectedDefaultPaths) {
        throw RuntimeException("FileSet default paths mismatch: got $defaultPaths, want $expectedDefaultPaths")
    }
    
    val f1Default = defaultFileSet.resolveFile("data/collection/file1.txt")
    val content1Default = f1Default.path.readText().trim()
    if (content1Default != "content of file 1 kt") {
        throw RuntimeException("file1Default.txt content mismatch: got '$content1Default', want 'content of file 1 kt'")
    }

    // Directory test
    val dir = TestResources.dirData.resolve()
    val df1 = dir.child("file1.txt")
    val dcontent1 = df1.path.readText().trim()
    if (dcontent1 != "file1 content kt") {
        throw RuntimeException("Directory file1.txt content mismatch: got '$dcontent1', want 'file1 content kt'")
    }

    val df2 = dir.child("file2.txt")
    val dcontent2 = df2.path.readText().trim()
    if (dcontent2 != "file2 content kt") {
        throw RuntimeException("Directory file2.txt content mismatch: got '$dcontent2', want 'file2 content kt'")
    }

    // Strict assertions tests
    val strictFile = TestResources.strictFile.resolve()
    if (strictFile.path.readText().trim() != "dummy content") {
        throw RuntimeException("StrictFile content mismatch")
    }

    val strictDir = TestResources.strictDir.resolve()
    if (!java.io.File(strictDir.path.toString()).isDirectory) {
        throw RuntimeException("StrictDir is not a directory")
    }

    val strictFileSet = TestResources.strictFileSet.resolve()
    if (strictFileSet.relPaths.sorted() != listOf("file1.txt", "file2.txt")) {
        throw RuntimeException("StrictFileSet paths mismatch")
    }

    val forcedFileSet = TestResources.forcedFileSet.resolve()
    if (forcedFileSet.relPaths != listOf("data/dummy.txt")) {
        throw RuntimeException("ForcedFileSet paths mismatch: got ${forcedFileSet.relPaths}")
    }
    val fc1 = forcedFileSet.resolveFile("data/dummy.txt")
    if (fc1.path.readText().trim() != "dummy content") {
        throw RuntimeException("ForcedFileSet file content mismatch")
    }

    val commonDirFileSet = TestResources.commonDirFileSet.resolve()
    if (commonDirFileSet.relPaths.sorted() != listOf("file1.txt", "file2.txt")) {
        throw RuntimeException("commonDirFileSet paths mismatch: got ${commonDirFileSet.relPaths}")
    }
    val lpg1 = commonDirFileSet.resolveFile("file1.txt")
    if (lpg1.path.readText().trim() != "content of file 1 kt") {
        throw RuntimeException("commonDirFileSet file1 content mismatch")
    }

    // Duplicate targets test
    val dt1 = TestResources.duplicateTarget1.resolve()
    if (dt1.relPaths.sorted() != listOf("file1.txt", "file2.txt")) {
        throw RuntimeException("duplicateTarget1 paths mismatch")
    }
    val dt2 = TestResources.duplicateTarget2.resolve()
    val expectedDt2Paths = listOf(
        "_main/data/collection/file1.txt",
        "_main/data/collection/file2.txt"
    )
    if (dt2.relPaths.sorted() != expectedDt2Paths) {
        throw RuntimeException("duplicateTarget2 paths mismatch: got ${dt2.relPaths.sorted()}, want $expectedDt2Paths")
    }

    // Mixed targets fileset test
    val mixedFs = TestResources.mixedFileSet.resolve()
    val expectedMixedPaths = listOf(
        "collection/file1.txt",
        "collection/file2.txt",
        "dummy.txt"
    )
    // Common prefix is "_main"
    if (mixedFs.relPaths.sorted() != expectedMixedPaths) {
        throw RuntimeException("mixedFileSet paths mismatch: got ${mixedFs.relPaths.sorted()}, want $expectedMixedPaths")
    }

    val m1 = mixedFs.resolveFile("collection/file1.txt")
    if (m1.path.readText().trim() != "content of file 1 kt") {
        throw RuntimeException("mixedFileSet file1 content mismatch")
    }
    val m2 = mixedFs.resolveFile("dummy.txt")
    if (m2.path.readText().trim() != "dummy content") {
        throw RuntimeException("mixedFileSet dummy content mismatch")
    }

    println("All tests passed!")
}
