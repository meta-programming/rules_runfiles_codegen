package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Helper running!")
	if len(os.Args) > 1 {
		fmt.Printf("Args: %v\n", os.Args[1:])
	}
}
