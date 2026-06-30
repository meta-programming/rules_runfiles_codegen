/**
 * Package com.github.metaprogramming.runfiles provides type-safe, explicit (non-eager) resolution of Bazel runfiles.
 *
 * This library is the runtime companion for the [rules_runfiles_codegen](https://github.com/meta-programming/rules_runfiles_codegen)
 * Bazel rules. The code generator emits structures that utilize the types in this package (such as [FileSpec]
 * and [ExecutableSpec]) to allow users to safely resolve runfiles at runtime.
 */
package com.github.metaprogramming.runfiles

import java.io.File as JFile
import java.nio.file.Path as JPath
import java.nio.file.Paths

/**
 * Exception thrown when a runfile cannot be resolved.
 */
class RunfileResolutionException(message: String, cause: Throwable? = null) : RuntimeException(message, cause)

/**
 * Represents a logical, runfiles-root-relative path used to locate a data dependency at runtime.
 */
@JvmInline
value class RlocationPath(val value: String) {
    init {
        require(value.isNotEmpty()) { "rlocation path cannot be empty" }
    }
    override fun toString(): String = value
}

/**
 * Interface for looking up runfiles and retrieving environment variables.
 */
fun interface Resolver {
    /**
     * Resolves the given runfile path to an absolute path, or returns null if it cannot be found.
     */
    fun rlocation(path: RlocationPath): String?

    /**
     * The environment variables to propagate to executables.
     * Defaults to empty map.
     */
    val envVars: Map<String, String>
        get() = emptyMap()

    /**
     * Default runfiles resolver that wraps the official Bazel Java runfiles library.
     * Can be overridden globally for testing by modifying [global].
     */
    object Default : Resolver {
        private val runfiles by lazy {
            try {
                com.google.devtools.build.runfiles.Runfiles.create()
            } catch (e: java.io.IOException) {
                throw RunfileResolutionException("Failed to initialize default Bazel Runfiles", e)
            }
        }

        private val systemResolver = object : Resolver {
            override fun rlocation(path: RlocationPath): String? {
                try {
                    return runfiles.rlocation(path.value)
                } catch (e: Exception) {
                    throw RunfileResolutionException("Error resolving path ${path.value}", e)
                }
            }

            override val envVars: Map<String, String>
                get() = try {
                    runfiles.envVars
                } catch (e: Exception) {
                    emptyMap()
                }
        }

        /**
         * The global resolver instance. Can be replaced with a mock in tests.
         */
        @Volatile
        var global: Resolver = systemResolver

        override fun rlocation(path: RlocationPath): String? = global.rlocation(path)
        override val envVars: Map<String, String> get() = global.envVars
    }
}

/**
 * Represents an unresolved runfile specification.
 */
class FileSpec(val rlocationPath: RlocationPath) {
    /**
     * Attempts to find the runfile on disk.
     *
     * @param resolver The resolver to use. Defaults to the default Bazel resolver.
     * @return A resolved [File] if successful.
     * @throws RunfileResolutionException If the runfile cannot be resolved.
     */
    fun resolve(resolver: Resolver = Resolver.Default): File {
        val resolvedPath = resolver.rlocation(rlocationPath)
            ?: throw RunfileResolutionException("Failed to resolve runfile: ${rlocationPath.value}")
        return File(rlocationPath, resolvedPath)
    }
}

/**
 * Represents an unresolved executable runfile specification.
 */
class ExecutableSpec(val rlocationPath: RlocationPath) {
    private val fileSpec = FileSpec(rlocationPath)

    /**
     * Attempts to find the executable on disk.
     */
    fun resolve(resolver: Resolver = Resolver.Default): Executable {
        val file = fileSpec.resolve(resolver)
        return Executable(file.rlocationPath, file.path, resolver.envVars)
    }
}

/**
 * Represents a runfile that has been successfully located on disk.
 */
open class File internal constructor(
    val rlocationPath: RlocationPath,
    val path: String
) {
    /**
     * The physical path to the runfile as a [JPath].
     */
    val jvmPath: JPath by lazy { Paths.get(path) }

    /**
     * The physical path to the runfile as a [JFile].
     */
    val file: JFile by lazy { JFile(path) }
}

/**
 * Represents an executable runfile successfully located on disk.
 */
class Executable internal constructor(
    rlocationPath: RlocationPath,
    path: String,
    val envVars: Map<String, String>
) : File(rlocationPath, path) {

    /**
     * Returns a [ProcessBuilder] pre-configured to run this executable,
     * with Bazel runfiles environment variables already propagated.
     */
    fun processBuilder(vararg args: String): ProcessBuilder {
        return ProcessBuilder(path, *args).apply {
            environment().putAll(envVars)
        }
    }
}
