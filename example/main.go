package main

import (
	"fmt"
	"os"

	"example/ability"
	"example/search"
)

func main() {
	fmt.Println("=== option-library example ===")

	runners := []struct {
		name string
		run  func() error
	}{
		{"formula", ability.Formula},
		{"search", search.Run},
		{"github.com/OptLTD/library/jsrunner", ability.JSRunner},
		{"github.com/OptLTD/library/jsmodule", ability.JSModule},
	}

	for _, item := range runners {
		if err := item.run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s example failed: %v\n", item.name, err)
			os.Exit(1)
		}
	}

	fmt.Println("\nall examples passed")
}
