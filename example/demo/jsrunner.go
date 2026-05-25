package demo

import (
	"context"
	"fmt"
	"time"

	js "jsrunner"
)

func JSRunner() error {
	fmt.Println("\n--- jsrunner ---")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value, err := js.RunString(ctx, `1 + 2 * 3`)
	if err != nil {
		return err
	}
	fmt.Printf("RunString: 1 + 2 * 3 = %v\n", value.Export())

	module, err := js.CompileModule("greet", `export default (name) => "hello, " + name`)
	if err != nil {
		return err
	}
	greeting, err := js.RunModule(ctx, module, "world")
	if err != nil {
		return err
	}
	fmt.Printf("RunModule: %v\n", greeting.Export())

	return nil
}
