package com.github.metaprogramming.runfiles

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertThrows
import org.junit.Assert.assertTrue
import org.junit.Test
import java.nio.file.Paths

class RunfileTest {

    private class MockResolver(
        private val paths: Map<String, String>,
        override val envVars: Map<String, String> = emptyMap()
    ) : Resolver {
        override fun rlocation(path: RlocationPath): String? = paths[path.value]
    }

    @Test
    fun testRlocationPath() {
        val path = RlocationPath("foo/bar")
        assertEquals("foo/bar", path.value)
        assertEquals("foo/bar", path.toString())
    }

    @Test
    fun testRlocationPathEmptyValidation() {
        val exception = assertThrows(IllegalArgumentException::class.java) {
            RlocationPath("")
        }
        assertTrue(exception.message!!.contains("rlocation path cannot be empty"))
    }

    @Test
    fun testFileSpecResolutionSuccess() {
        val rpath = RlocationPath("foo/bar.txt")
        val absPath = "/absolute/path/to/foo/bar.txt"
        val mockResolver = MockResolver(mapOf(rpath.value to absPath))

        val spec = FileSpec(rpath)
        val file = spec.resolve(mockResolver)

        assertEquals(rpath, file.rlocationPath)
        assertEquals(Paths.get(absPath), file.path)
    }

    @Test
    fun testFileSpecResolutionFailure() {
        val rpath = RlocationPath("foo/bar.txt")
        val mockResolver = MockResolver(emptyMap())

        val spec = FileSpec(rpath)
        val exception = assertThrows(RunfileResolutionException::class.java) {
            spec.resolve(mockResolver)
        }
        assertTrue(exception.message!!.contains("Failed to resolve runfile: foo/bar.txt"))
    }

    @Test
    fun testExecutableSpecResolutionSuccess() {
        val rpath = RlocationPath("bin/tool")
        val absPath = "/absolute/path/to/bin/tool"
        val mockEnv = mapOf("VAR1" to "value1", "VAR2" to "value2")
        val mockResolver = MockResolver(mapOf(rpath.value to absPath), mockEnv)

        val spec = ExecutableSpec(rpath)
        val executable = spec.resolve(mockResolver)

        assertEquals(rpath, executable.rlocationPath)
        assertEquals(Paths.get(absPath), executable.path)
        assertEquals(mockEnv, executable.envVars)

        val pb = executable.processBuilder("arg1", "arg2")
        assertEquals(listOf(absPath, "arg1", "arg2"), pb.command())
        assertEquals("value1", pb.environment()["VAR1"])
        assertEquals("value2", pb.environment()["VAR2"])

        // Test the shortcut directly on the ExecutableSpec
        val pbDirect = spec.processBuilder("arg1", "arg2", resolver = mockResolver)
        assertEquals(listOf(absPath, "arg1", "arg2"), pbDirect.command())
        assertEquals("value1", pbDirect.environment()["VAR1"])
        assertEquals("value2", pbDirect.environment()["VAR2"])
    }

    @Test
    fun testExecutableSpecResolutionFailure() {
        val rpath = RlocationPath("bin/tool")
        val mockResolver = MockResolver(emptyMap())

        val spec = ExecutableSpec(rpath)
        val exception = assertThrows(RunfileResolutionException::class.java) {
            spec.resolve(mockResolver)
        }
        assertTrue(exception.message!!.contains("Failed to resolve runfile: bin/tool"))
    }

    @Test
    fun testDefaultResolverSuccess() {
        // Since we are running under Bazel, the default resolver should be able to resolve
        // files declared in the `data` attribute of this test.
        // We declared "Runfile.kt" as data in the BUILD file.
        val spec = FileSpec(RlocationPath("rules_runfile_codegen_kotlin/runfiles/Runfile.kt"))
        val file = spec.resolve()
        assertNotNull(file.path)
        assertTrue(file.path.toFile().exists())
    }

    @Test
    fun testPathShortcut() {
        val spec = FileSpec(RlocationPath("rules_runfile_codegen_kotlin/runfiles/Runfile.kt"))
        val resolvedPath = spec.path
        assertNotNull(resolvedPath)
        assertTrue(resolvedPath.toFile().exists())
        assertEquals(spec.resolve().path, resolvedPath)
    }

    @Test
    fun testFileSetSpecShortcut() {
        val filesMap = mapOf(
            "file1.txt" to "foo/bar1.txt",
            "file2.txt" to "foo/bar2.txt"
        )
        val mockResolver = MockResolver(mapOf(
            "foo/bar1.txt" to "/abs/bar1.txt",
            "foo/bar2.txt" to "/abs/bar2.txt"
        ))

        val spec = FileSetSpec(filesMap)
        
        // Use bracket operator on FileSetSpec to get FileSpec
        val fileSpec1 = spec["file1.txt"]
        assertEquals(RlocationPath("foo/bar1.txt"), fileSpec1.rlocationPath)
        
        // Resolve path via resolver to verify it works
        assertEquals(Paths.get("/abs/bar1.txt"), fileSpec1.resolve(mockResolver).path)

        // Verify invalid file throws exception
        assertThrows(RunfileResolutionException::class.java) {
            spec["unknown.txt"]
        }
    }
}
