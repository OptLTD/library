package main

import (
	"fmt"
	"os"

	"example/demo"
)

func main() {
	fmt.Println("=== option-library example ===")

	runners := []struct {
		name string
		run  func() error
	}{
		{"formula", demo.Formula},
		{"search", demo.Search},
		{"jsrunner", demo.JSRunner},
		{"jsmodule", demo.JSModule},
	}

	for _, item := range runners {
		if err := item.run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s example failed: %v\n", item.name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\nall examples passed")
}
