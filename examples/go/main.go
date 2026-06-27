package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/example/project/examples/go/resources"
)

func main() {
	fmt.Println("Data file path:", resources.DataFile)
	data, err := os.ReadFile(resources.DataFile)
	if err != nil {
		log.Fatalf("Failed to read data file: %v", err)
	}
	fmt.Printf("Data file contents: %s\n", string(data))

	fmt.Println("Helper tool path:", resources.HelperTool)
	cmd := exec.Command(resources.HelperTool, "arg1", "arg2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run helper tool: %v", err)
	}
}
