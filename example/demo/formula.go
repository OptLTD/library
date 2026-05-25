package demo

import (
	"fmt"

	formulalib "github.com/OptLTD/library/formula"
)

func Formula() error {
	fmt.Println("\n--- formula ---")

	combine := map[string]any{
		"basic.qty":   3.0,
		"basic.price": 100.0,
	}
	env := formulalib.FormulaEnv(combine)

	total, err := formulalib.Build(`basic.qty * basic.price`, env)
	if err != nil {
		return err
	}
	fmt.Printf("qty * price = %v\n", total)

	level, err := formulalib.Build(
		`IFS(basic.qty >= 3, "bulk", basic.qty >= 1, "retail", true, "none")`,
		env,
	)
	if err != nil {
		return err
	}
	fmt.Printf("price level = %v\n", level)

	return nil
}
