package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func main() {
	r, err := runfiles.New()
	if err != nil {
		log.Fatalf("helper: failed to init runfiles: %v", err)
	}

	// Try resolving with the module name (Bzlmod style)
	path, err := r.Rlocation("rules_runfile_codegen_go_tests/data/helper_data.txt")
	if err != nil {
		// Fallback to _main just in case
		path, err = r.Rlocation("_main/data/helper_data.txt")
	}
	if err != nil {
		log.Fatalf("helper: failed to resolve runfile: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("helper: failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("helper: failed to read file: %v", err)
	}

	fmt.Print(string(content))
}
