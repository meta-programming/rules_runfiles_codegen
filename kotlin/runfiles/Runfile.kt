package com.github.metaprogramming.runfiles

import java.nio.file.Path
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
     * The environment variables necessary to propagate runfiles location information
     * to subprocesses. Defaults to an empty map (useful for mocks/custom resolvers).
     *
     * For the default resolver, this contains variables like `RUNFILES_DIR` or
     * `RUNFILES_MANIFEST_FILE` which Bazel uses to locate runfiles.
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
}

/**
 * Represents an unresolved runfile specification.
 */
class FileSpec(
    /** The logical, runfiles-root-relative path for this runfile. */
    val rlocationPath: RlocationPath
) {
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
        return File(rlocationPath, Paths.get(resolvedPath))
    }
}

/**
 * Represents an unresolved executable runfile specification.
 */
class ExecutableSpec(
    /** The logical, runfiles-root-relative path for this executable runfile. */
    val rlocationPath: RlocationPath
) {
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
    /** The logical, runfiles-root-relative path that was resolved. */
    val rlocationPath: RlocationPath,
    /** The physical, absolute path to the runfile on disk. */
    val path: Path
)

/**
 * Represents an executable runfile successfully located on disk.
 */
class Executable internal constructor(
    rlocationPath: RlocationPath,
    path: Path,
    /**
     * The environment variables necessary to propagate runfiles location information
     * to this executable when run as a subprocess.
     *
     * These are automatically applied when using [processBuilder].
     */
    val envVars: Map<String, String>
) : File(rlocationPath, path) {

    /**
     * Returns a [ProcessBuilder] pre-configured to run this executable,
     * with Bazel runfiles environment variables already propagated.
     */
    fun processBuilder(vararg args: String): ProcessBuilder {
        return ProcessBuilder(path.toString(), *args).apply {
            environment().putAll(envVars)
        }
    }
}
