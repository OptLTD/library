package demo

import (
	"context"
	"fmt"
	"time"

	_ "jsmodule/buffer"
	_ "jsmodule/encoding"

	js "jsrunner"
)

func JSModule() error {
	fmt.Println("\n--- jsmodule ---")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	encoded, err := js.RunString(ctx, `btoa("hello")`)
	if err != nil {
		return err
	}
	fmt.Printf("buffer.btoa: %v\n", encoded.Export())

	decoded, err := js.RunString(ctx, `atob("`+encoded.String()+`")`)
	if err != nil {
		return err
	}
	fmt.Printf("buffer.atob: %v\n", decoded.Export())

	text, err := js.RunString(ctx, `new TextEncoder().encode("你好").length`)
	if err != nil {
		return err
	}
	fmt.Printf("encoding.TextEncoder byte length: %v\n", text.Export())

	return nil
}
