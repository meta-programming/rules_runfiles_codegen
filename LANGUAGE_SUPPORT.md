# Language Support Plan: rules_runfile_codegen

This document outlines the plan for extending `rules_runfile_codegen` to support additional programming languages beyond the initial Go and Kotlin implementations. It surveys Bazel's ecosystem, analyzes the available runfiles runtime libraries for each language (at both Bazel and language-package-manager levels), sketches proposed APIs, and addresses challenges like multiple ruleset implementations.

---

## 1. Overview & Core Principles

Regardless of the target language, all implementations should adhere to our core design principles:
1.  **Type-Safety**: Expose runfiles as strongly-typed symbols, not raw strings.
2.  **Resolution Strategy**: Support both eager (fail-fast at startup) and lazy (explicit/deferred) resolution, choosing the most idiomatic approach for each language (e.g., lazy for Go to avoid `init()` side-effects and improve testability, eager for Kotlin/Java to ensure fail-fast safety).
3.  **Rich Object Wrappers**: Distinguish between regular files (`Runfile`) and executables (`ExecutableRunfile`), providing subprocess execution helpers that automatically propagate the Bazel runfiles environment.
4.  **Minimal Dependencies**: Package each language as a separate Bzlmod module (e.g., `rules_runfile_codegen_python`) so users only pull in what they need.

---

## 2. Runfiles Runtime Libraries Analysis

To generate code that resolves runfiles, we must rely on the runtime libraries provided by the Bazel ecosystem. We must analyze these dependencies at two levels:
1.  **Bazel Level (Bzlmod)**: What Bazel modules must be loaded in `MODULE.bazel`.
2.  **Language Level (Package Managers)**: What dependencies must be declared in language-specific package managers (e.g., `go.mod`, `package.json`, `Cargo.toml`, Maven/Gradle) for IDE resolution and compilation.

### Runtime Library Comparison

| Language | Runtime Library | Bazel-Level Dependency | Language-Level Dependency | Env Propagation | Platform Robustness |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Go** | `@rules_go//go/runfiles` | **Zero** (part of `rules_go`) | **Zero** (implicitly resolved by `rules_go` import mapping) | **Excellent** (`Env()`) | **Excellent** |
| **Java / Kotlin** | `@bazel_tools//tools/java/runfiles` | **Zero** (built into `@bazel_tools`) | **Zero** (compiled against local jar; Maven artifact `com.google.devtools.build:runfiles` exists but is optional for Bazel-only builds) | **Excellent** (`envVars()`) | **Excellent** |
| **Python** | `@rules_python//python/runfiles` | **Zero** (part of `rules_python`) | **Zero** (distributed as source within `rules_python`, no `pip` package needed) | **Excellent** (`EnvVars()`) | **Excellent** |
| **C++** | `@bazel_tools//tools/cpp/runfiles` | **Zero** (built into `@bazel_tools`) | **Zero** (header-only/local library) | **Excellent** (`EnvVars()`) | **Excellent** |
| **Rust** | `runfiles` (crate from `rules_rust`) | **Zero** (part of `rules_rust`) | **Optional** (crate `runfiles` exists on crates.io, but Bazel resolves it internally via `@rules_rust//tools/runfiles`) | **Good** (`env_vars()`) | **Good** |
| **TypeScript / JS** | `@bazel/runfiles` | **Medium** (requires `rules_js` setup) | **Required** (must add `@bazel/runfiles` to `package.json`) | **Good** (`envVars()`) | **Good** |

---

### Detailed Dependency Analysis

#### 2.1. Go, Python, and C++ (Zero-Overhead baseline)
These languages represent the ideal integration state:
*   **Bazel**: The runfiles libraries are bundled with the core rulesets (`rules_go`, `rules_python`) or Bazel itself (`@bazel_tools` for C++).
*   **Language-Level**: **No action required.** 
    *   In Go, `rules_go` automatically handles the import path `github.com/bazelbuild/rules_go/go/runfiles` without requiring it in the user's `go.mod`.
    *   In Python, the library is imported directly from the `rules_python` package space, requiring no `pip` installation.
    *   In C++, it is linked as a local Bazel target (`@bazel_tools//tools/cpp/runfiles`), requiring no external package manager.

#### 2.2. Java / Kotlin (Optional Maven)
*   **Bazel**: Uses `@bazel_tools//tools/java/runfiles`, which is always available.
*   **Language-Level**: While Bazel compiles against the local jar provided by `@bazel_tools`, the library is also published to Maven Central as `com.google.devtools.build:runfiles`. 
    *   *IDE Integration*: If users want their IDE (like IntelliJ) to resolve the classes without red squigglies in hybrid projects, they *may* want to add it to their Maven/Gradle dependencies, but it is **not required** for the Bazel build to succeed.

#### 2.3. Rust (Optional Cargo)
*   **Bazel**: Resolves the `runfiles` crate internally from `rules_rust`.
*   **Language-Level**: The `runfiles` crate is published on crates.io. 
    *   *IDE Integration*: Similar to Java, if users use Cargo-based tooling (like `rust-analyzer` outside of Bazel-integrated editors), they may need to declare the `runfiles` dependency in `Cargo.toml`, but Bazel handles it internally.

#### 2.4. TypeScript / JavaScript (Strict NPM Requirement)
TypeScript/JavaScript is the only language that **requires** a language-level dependency.
*   **Bazel & Language-Level**: Because of how `rules_js` works (mapping all node dependencies through the virtual `node_modules` directory), the generated code *must* import from `@bazel/runfiles`.
*   **User Action**: The user **must** run `npm install @bazel/runfiles` (or equivalent) to add it to their `package.json`. 
*   **Generator Constraint**: The generated library's `BUILD.bazel` must link against the NPM dependency target (typically `//:node_modules/@bazel/runfiles` or via a helper macro), which means our generator must be aware of the user's NPM workspace structure.

---

## 3. Detailed Language Designs

### 3.1. Python

Python has two major rulesets: the official `rules_python` and Aspect's `aspect_rules_py`. 

#### The Python Ruleset Duality & Compatibility:
*   **`rules_python` (Official)**: The standard ruleset.
*   **`aspect_rules_py` (Aspect)**: A high-performance alternative designed as a drop-in replacement.
*   **Co-existence**: `aspect_rules_py` is built on top of `rules_python` and does not replace it entirely. In fact, projects using `aspect_rules_py` still use `rules_python` in their `MODULE.bazel` to configure and register Python toolchains. Therefore, adding `rules_python` as a dependency in our module (`rules_runfile_codegen_python`) does not introduce any new or unwanted dependencies for Aspect users.
*   **Runtime Compatibility**: Both rulesets are fully compatible with the official `@rules_python//python/runfiles` library. Our generated Python code will always import from `rules_python.python.runfiles`, which is the recommended approach for both ecosystems.
*   **Build-time Rule (`py_library`)**: Our public macro `py_runfile_library` will, by default, generate a target using `rules_python`'s `py_library`. Because `aspect_rules_py` targets (like `py_binary`) are designed to be drop-in replacements, they can seamlessly depend on `rules_python` targets. 
*   **Future-Proofing & Bzlmod Constraints**: Bzlmod does not support optional dependencies. If our `rules_runfile_codegen_python` module loaded `aspect_rules_py` definitions, it would force a dependency on `aspect_rules_py` (and its transitive chain) for all users. To avoid this and maintain **REQ-3 (Minimal Dependencies)**, we have two designs if we ever need to support Aspect-specific library rules:
    *   **Design 1: Separate Bzlmod Modules (Recommended)**: We split the support into `rules_runfile_codegen_python` (depending only on `rules_python`) and `rules_runfile_codegen_aspect_py` (depending on `aspect_rules_py`). Both would share a common Starlark generator core module (`rules_runfile_codegen_core`), ensuring zero code duplication.
    *   **Design 2: Starlark Dependency Injection**: We can design the `py_runfile_library` macro to accept the `py_library` rule as an optional parameter (`py_library_rule = py_library`). Aspect users can then load `aspect_py_library` themselves and pass it in, avoiding any Bzlmod-level dependency in our module.
    *   *Note*: Due to the high interoperability between the two rulesets, the standard `rules_python` `py_library` is highly likely to be sufficient for all users, making these designs a contingency.

#### BUILD.bazel Usage:
```python
load("@rules_runfile_codegen_python//python:defs.bzl", "py_runfile_library", "py_runfile")

py_runfile_library(
    name = "my_resources",
    importpath = "path.to.myapp.resources",
    entries = [
        py_runfile(
            name = "ConfigJSON",
            target = "//path/to:config.json",
            doc = "A configuration file.",
        ),
        py_runfile(
            name = "HelperTool",
            target = "//path/to/tools:helper",
            doc = "A helper tool.",
        ),
    ],
)
```

#### Proposed Python API Sketch:

```python
# Generated code: resources.py
from rules_python.python.runfiles import runfiles
import os
import subprocess
import sys

class Runfile:
    def __init__(self, rlocation_path: str):
        self._rlocation_path = rlocation_path
        # Eager resolution at import time
        self._abs_path = _resolver.Rlocation(rlocation_path)
        if not self._abs_path:
            print(f"FATAL: failed to resolve runfile '{rlocation_path}'", file=sys.stderr)
            sys.exit(1)

    def path(self) -> str:
        return self._abs_path

class ExecutableRunfile(Runfile):
    def cmd(self, args: list[str], **kwargs) -> subprocess.Popen:
        """Returns a Popen object pre-configured with runfiles env."""
        env = os.environ.copy()
        env.update(_resolver.EnvVars())
        return subprocess.Popen([self.path()] + args, env=env, **kwargs)

# Initialize the global resolver
try:
    _resolver = runfiles.Create()
except Exception as e:
    print(f"FATAL: failed to initialize runfiles registry: {e}", file=sys.stderr)
    sys.exit(1)

# Strongly-typed symbols
ConfigJSON = Runfile("_main/path/to/config.json")
HelperTool = ExecutableRunfile("_main/path/to/helper")
```

---

### 3.2. C++

C++ uses the native Bazel `cc_*` rules and the built-in `@bazel_tools//tools/cpp/runfiles` library.

#### BUILD.bazel Usage:
```python
load("@rules_runfile_codegen_cc//cc:defs.bzl", "cc_runfile_library", "cc_runfile")

cc_runfile_library(
    name = "my_resources",
    namespace = "myapp::resources",
    entries = [
        cc_runfile(
            name = "ConfigJSON",
            target = "//path/to:config.json",
            doc = "A configuration file.",
        ),
        cc_runfile(
            name = "HelperTool",
            target = "//path/to/tools:helper",
            doc = "A helper tool.",
        ),
    ],
)
```

#### Proposed C++ API Sketch:
C++ lacks `init` blocks, but we can achieve eager startup resolution by leveraging **global static initialization**. The constructors of the global `Runfile` instances will run before `main()`, ensuring the program fails fast if any runfile is missing.

```cpp
// my_resources.h
#pragma once
#include <string>
#include <vector>
#include <memory>

class Runfile {
public:
    Runfile(const std::string& rlocation_path);
    std::string Path() const;
protected:
    std::string abs_path_;
};

class ExecutableRunfile : public Runfile {
public:
    using Runfile::Runfile;
    // Returns env vars in KEY=VALUE format for execve/posix_spawn
    std::vector<std::string> Env() const;
};

namespace my_resources {
    extern const Runfile ConfigJSON;
    extern const ExecutableRunfile HelperTool;
}
```

```cpp
// my_resources.cc
#include "my_resources.h"
#include "tools/cpp/runfiles/runfiles.h"
#include <stdexcept>

namespace {
    using bazel::tools::cpp::runfiles::Runfiles;
    std::unique_ptr<Runfiles> runfiles_registry;

    void InitRegistry() {
        if (!runfiles_registry) {
            std::string error;
            runfiles_registry.reset(Runfiles::CreateForTest(&error));
            if (!runfiles_registry) {
                fprintf(stderr, "FATAL: failed to initialize runfiles: %s\n", error.c_str());
                exit(1);
            }
        }
    }

    std::string Resolve(const std::string& path) {
        InitRegistry();
        std::string abs_path = runfiles_registry->Rlocation(path);
        if (abs_path.empty()) {
            fprintf(stderr, "FATAL: failed to resolve runfile: %s\n", path.c_str());
            exit(1);
        }
        return abs_path;
    }
}

Runfile::Runfile(const std::string& rlocation_path) : abs_path_(Resolve(rlocation_path)) {}
std::string Runfile::Path() const { return abs_path_; }

std::vector<std::string> ExecutableRunfile::Env() const {
    InitRegistry();
    std::vector<std::string> env;
    for (const auto& var : runfiles_registry->EnvVars()) {
        env.push_back(var.first + "=" + var.second);
    }
    return env;
}

namespace my_resources {
    // Global constructors run before main(), enforcing eager resolution
    const Runfile ConfigJSON("_main/path/to/config.json");
    const ExecutableRunfile HelperTool("_main/path/to/helper");
}
```

---

### 3.3. Java

Java uses the native `java_*` rules and the built-in `@bazel_tools//tools/java/runfiles` library (same as Kotlin).

#### BUILD.bazel Usage:
```python
load("@rules_runfile_codegen_java//java:defs.bzl", "java_runfile_library", "java_runfile")

java_runfile_library(
    name = "my_resources",
    package = "com.example.project.resources",
    entries = [
        java_runfile(
            name = "ConfigJSON",
            target = "//path/to:config.json",
            doc = "A configuration file.",
        ),
        java_runfile(
            name = "HelperTool",
            target = "//path/to/tools:helper",
            doc = "A helper tool.",
        ),
    ],
)
```

#### Proposed Java API Sketch:
We use a `static` initializer block in the generated class to enforce eager resolution at class-load time.

```java
// MyResources.java
package com.example.project.resources;

import com.google.devtools.build.runfiles.Runfiles;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Map;

public final class MyResources {
    private static final Runfiles _runfiles;

    static {
        try {
            _runfiles = Runfiles.create();
        } catch (IOException e) {
            throw new RuntimeException("FATAL: Failed to create runfiles registry", e);
        }
    }

    public static class Runfile {
        private final String absPath;

        public Runfile(String rlocationPath) {
            String resolved = _runfiles.rlocation(rlocationPath);
            if (resolved == null) {
                throw new RuntimeException("FATAL: Failed to resolve runfile: " + rlocationPath);
            }
            this.absPath = resolved;
        }

        public String path() { return absPath; }
        public Path jvmPath() { return Paths.get(absPath); }
    }

    public static class ExecutableRunfile extends Runfile {
        public ExecutableRunfile(String rlocationPath) {
            super(rlocationPath);
        }

        public ProcessBuilder processBuilder(String... args) {
            ProcessBuilder pb = new ProcessBuilder(args);
            pb.command().add(0, path());
            pb.environment().putAll(_runfiles.envVars());
            return pb;
        }
    }

    // Eagerly initialized constants
    public static final Runfile ConfigJSON = new Runfile("_main/path/to/config.json");
    public static final ExecutableRunfile HelperTool = new ExecutableRunfile("_main/path/to/helper");

    private MyResources() {}
}
```

---

### 3.4. Rust

Rust uses `rules_rust` and the official `runfiles` crate.

#### BUILD.bazel Usage:
```python
load("@rules_runfile_codegen_rust//rust:defs.bzl", "rust_runfile_library", "rust_runfile")

rust_runfile_library(
    name = "my_resources",
    # Generates a module named my_resources
    entries = [
        rust_runfile(
            name = "CONFIG_JSON",
            target = "//path/to:config.json",
            doc = "A configuration file.",
        ),
        rust_runfile(
            name = "HELPER_TOOL",
            target = "//path/to/tools:helper",
            doc = "A helper tool.",
        ),
    ],
)
```

#### Proposed Rust API Sketch:
Rust does not support life before `main()` (no static constructors) without external crates. To keep dependencies minimal and idiomatic, we use `std::sync::OnceLock` for thread-safe lazy initialization, but we can also provide an explicit `validate()` function that users can call at the beginning of `main()` if they want to enforce eager startup checks.

```rust
// my_resources.rs
use std::path::{Path, PathBuf};
use std::process::Command;
use std::sync::OnceLock;
use runfiles::Runfiles;

static RUNFILES: OnceLock<Runfiles> = OnceLock::new();

fn get_runfiles() -> &'static Runfiles {
    RUNFILES.get_or_init(|| {
        Runfiles::create().expect("FATAL: Failed to initialize runfiles registry")
    })
}

fn resolve(rlocation_path: &str) -> PathBuf {
    get_runfiles()
        .rlocation(rlocation_path)
        .unwrap_or_else(|| panic!("FATAL: Failed to resolve runfile: {}", rlocation_path))
}

pub struct Runfile {
    rlocation_path: &'static str,
    abs_path: OnceLock<PathBuf>,
}

impl Runfile {
    const fn new(rlocation_path: &'static str) -> Self {
        Self {
            rlocation_path,
            abs_path: OnceLock::new(),
        }
    }

    pub fn path(&self) -> &Path {
        self.abs_path.get_or_init(|| resolve(self.rlocation_path))
    }
}

pub struct ExecutableRunfile {
    runfile: Runfile,
}

impl ExecutableRunfile {
    const fn new(rlocation_path: &'static str) -> Self {
        Self {
            runfile: Runfile::new(rlocation_path),
        }
    }

    pub fn path(&self) -> &Path {
        self.runfile.path()
    }

    pub fn cmd(&self) -> Command {
        let mut cmd = Command::new(self.path());
        for (k, v) in get_runfiles().env_vars() {
            cmd.env(k, v);
        }
        cmd
    }
}

// Generated symbols
pub static CONFIG_JSON: Runfile = Runfile::new("_main/path/to/config.json");
pub static HELPER_TOOL: ExecutableRunfile = ExecutableRunfile::new("_main/path/to/helper");

/// Optional: Eager validation helper that users can call in main()
pub fn validate() {
    let _ = CONFIG_JSON.path();
    let _ = HELPER_TOOL.path();
}
```

---

### 3.5. TypeScript / JavaScript

TS/JS uses Aspect's `rules_js`/`rules_ts` and the `@bazel/runfiles` NPM package.

#### BUILD.bazel Usage:
```python
load("@rules_runfile_codegen_ts//ts:defs.bzl", "ts_runfile_library", "ts_runfile")

ts_runfile_library(
    name = "my_resources",
    # Generates my_resources.ts
    entries = [
        ts_runfile(
            name = "ConfigJSON",
            target = "//path/to:config.json",
            doc = "A configuration file.",
        ),
        ts_runfile(
            name = "HelperTool",
            target = "//path/to/tools:helper",
            doc = "A helper tool.",
        ),
    ],
)
```

#### Proposed TypeScript API Sketch:
In JS/TS, top-level module code runs immediately when the module is imported, giving us eager resolution by default.

```typescript
// my_resources.ts
import { Runfiles } from '@bazel/runfiles';
import * as child_process from 'child_process';

let runfiles: Runfiles;
try {
    runfiles = new Runfiles();
} catch (e) {
    console.error("FATAL: Failed to initialize runfiles registry", e);
    process.exit(1);
}

export class Runfile {
    readonly absPath: string;

    constructor(rlocationPath: string) {
        const resolved = runfiles.rlocation(rlocationPath);
        if (!resolved) {
            console.error(`FATAL: Failed to resolve runfile: ${rlocationPath}`);
            process.exit(1);
        }
        this.absPath = resolved;
    }
}

export class ExecutableRunfile extends Runfile {
    spawn(args: string[], options?: child_process.SpawnOptions): child_process.ChildProcess {
        const env = { ...process.env, ...runfiles.envVars() };
        return child_process.spawn(this.absPath, args, { ...options, env });
    }
}

// Eagerly resolved at module load time
export const ConfigJSON = new Runfile('_main/path/to/config.json');
export const HelperTool = new ExecutableRunfile('_main/path/to/helper');
```

---

## 4. Extensibility Architecture (Generator Refactoring)

To make adding new languages as easy as possible, we should refactor the Starlark generator logic. 

Currently, the path resolution and validation logic are mixed with the Go/Kotlin code generation in `internal/rules.bzl`. 

### Proposed Refactoring: Separating Core Logic from Emitters

We can create a shared Starlark helper library (e.g., `//internal:common.bzl`) that handles the heavy lifting:
1.  **Input Validation**: Checking for duplicates, empty names, and basic identifier rules.
2.  **Target Analysis**: Detecting if a target is executable (using `is_source` fixes) and ensuring it produces exactly one output.
3.  **Path Resolution**: Computing the correct `rlocation` path (handling external repos and main repo prefixes).

This helper library will return a structured list of provider-like objects (structs) containing:
*   `name`: The sanitized symbol name.
*   `rlocation_path`: The resolved runfile path.
*   `is_executable`: Boolean flag.
*   `doc`: The docstring.

Each language-specific generator rule (e.g., `py_runfile_gen`, `cc_runfile_gen`) will then:
1.  Call the shared helper to get the resolved entries.
2.  Focus **only** on emitting the language-specific source code (string templating).

This drastically reduces the effort to onboard a new language (reducing it to a simple string-template emitter) and ensures that any future bug fixes in path resolution or validation (like the `is_source` bug) are automatically applied to all languages.
