package com.github.metaprogramming.runfiles

import com.google.devtools.build.runfiles.Runfiles
import java.io.File as JFile
import java.io.IOException
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
 * Interface for looking up runfiles.
 */
fun interface Resolver {
    /**
     * Resolves the given runfile path to an absolute path, or returns null if it cannot be found.
     */
    fun rlocation(path: RlocationPath): String?
}

/**
 * Interface for providing environment variables.
 */
interface EnvProvider {
    /**
     * Returns the environment variables that should be propagated to subprocesses.
     */
    val envVars: Map<String, String>
}

/**
 * Default runfiles resolver that wraps the official Bazel Java runfiles library.
 * Can be overridden for testing by modifying [resolver] and [envProvider].
 */
object RunfileResolver : Resolver, EnvProvider {
    private val runfiles: Runfiles by lazy {
        try {
            Runfiles.create()
        } catch (e: IOException) {
            throw RunfileResolutionException("Failed to initialize default Bazel Runfiles", e)
        }
    }

    private val defaultResolver = Resolver { path ->
        try {
            runfiles.rlocation(path.value)
        } catch (e: Exception) {
            throw RunfileResolutionException("Error resolving path ${path.value}", e)
        }
    }

    private val defaultEnvProvider = object : EnvProvider {
        override val envVars: Map<String, String>
            get() = try {
                runfiles.envVars
            } catch (e: Exception) {
                emptyMap()
            }
    }

    @Volatile
    var resolver: Resolver = defaultResolver

    @Volatile
    var envProvider: EnvProvider = defaultEnvProvider

    override fun rlocation(path: RlocationPath): String? = resolver.rlocation(path)

    override val envVars: Map<String, String>
        get() = envProvider.envVars
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
    fun resolve(resolver: Resolver = RunfileResolver): File {
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
    fun resolve(resolver: Resolver = RunfileResolver): Executable {
        val file = fileSpec.resolve(resolver)
        val env = (resolver as? EnvProvider) ?: RunfileResolver
        return Executable(file.rlocationPath, file.path, env)
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
    private val envProvider: EnvProvider = RunfileResolver
) : File(rlocationPath, path) {

    /**
     * Returns a [ProcessBuilder] pre-configured to run this executable,
     * with Bazel runfiles environment variables already propagated.
     */
    fun processBuilder(vararg args: String): ProcessBuilder {
        return ProcessBuilder(path, *args).apply {
            environment().putAll(envProvider.envVars)
        }
    }
}
