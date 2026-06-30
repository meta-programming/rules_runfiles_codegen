package version

import (
	"fmt"
	"strings"

	"github.com/bazelbuild/buildtools/build"
)

// Parse extracts the version string from MODULE.bazel content.
func Parse(content []byte) (string, error) {
	ast, err := build.ParseModule("MODULE.bazel", content)
	if err != nil {
		return "", err
	}
	for _, stmt := range ast.Stmt {
		call, ok := stmt.(*build.CallExpr)
		if !ok {
			continue
		}
		ident, ok := call.X.(*build.Ident)
		if !ok || ident.Name != "module" {
			continue
		}
		for _, arg := range call.List {
			assign, ok := arg.(*build.AssignExpr)
			if !ok {
				continue
			}
			key, ok := assign.LHS.(*build.Ident)
			if ok && key.Name == "version" {
				stringVal, ok := assign.RHS.(*build.StringExpr)
				if ok {
					return stringVal.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("version not found in module() declaration")
}

// Update replaces the version string in MODULE.bazel content and returns the updated, formatted content.
// It updates both the module(version = "...") and any bazel_dep(name = "rules_runfile_codegen_*", version = "...") declarations.
func Update(content []byte, newVersion string) ([]byte, error) {
	ast, err := build.ParseModule("MODULE.bazel", content)
	if err != nil {
		return nil, err
	}
	moduleUpdated := false
	for _, stmt := range ast.Stmt {
		call, ok := stmt.(*build.CallExpr)
		if !ok {
			continue
		}
		ident, ok := call.X.(*build.Ident)
		if !ok {
			continue
		}

		if ident.Name == "module" {
			for _, arg := range call.List {
				assign, ok := arg.(*build.AssignExpr)
				if !ok {
					continue
				}
				key, ok := assign.LHS.(*build.Ident)
				if ok && key.Name == "version" {
					assign.RHS = &build.StringExpr{Value: newVersion}
					moduleUpdated = true
					break
				}
			}
		} else if ident.Name == "bazel_dep" {
			// Check if this bazel_dep is one of ours
			isOurs := false
			var versionAssign *build.AssignExpr

			for _, arg := range call.List {
				assign, ok := arg.(*build.AssignExpr)
				if !ok {
					continue
				}
				key, ok := assign.LHS.(*build.Ident)
				if !ok {
					continue
				}

				if key.Name == "name" {
					strExpr, ok := assign.RHS.(*build.StringExpr)
					if ok && strings.HasPrefix(strExpr.Value, "rules_runfile_codegen_") {
						isOurs = true
					}
				} else if key.Name == "version" {
					versionAssign = assign
				}
			}

			if isOurs && versionAssign != nil {
				versionAssign.RHS = &build.StringExpr{Value: newVersion}
			}
		}
	}
	if !moduleUpdated {
		return nil, fmt.Errorf("version not found in module() declaration")
	}
	return build.Format(ast), nil
}

// ParseDeps extracts the versions of any bazel_dep declarations pointing to rules_runfile_codegen_* modules.
func ParseDeps(content []byte) (map[string]string, error) {
	ast, err := build.ParseModule("MODULE.bazel", content)
	if err != nil {
		return nil, err
	}
	deps := make(map[string]string)
	for _, stmt := range ast.Stmt {
		call, ok := stmt.(*build.CallExpr)
		if !ok {
			continue
		}
		ident, ok := call.X.(*build.Ident)
		if !ok || ident.Name != "bazel_dep" {
			continue
		}

		var name, ver string
		for _, arg := range call.List {
			assign, ok := arg.(*build.AssignExpr)
			if !ok {
				continue
			}
			key, ok := assign.LHS.(*build.Ident)
			if !ok {
				continue
			}
			if key.Name == "name" {
				strExpr, ok := assign.RHS.(*build.StringExpr)
				if ok && strings.HasPrefix(strExpr.Value, "rules_runfile_codegen_") {
					name = strExpr.Value
				}
			} else if key.Name == "version" {
				strExpr, ok := assign.RHS.(*build.StringExpr)
				if ok {
					ver = strExpr.Value
				}
			}
		}
		if name != "" && ver != "" {
			deps[name] = ver
		}
	}
	return deps, nil
}
