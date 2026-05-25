package formula

import (
	"math"
	"testing"

	"github.com/expr-lang/expr"
)

func TestExcelIF(t *testing.T) {
	var lg LogicFuncs
	run := func(args ...any) any {
		v, err := lg.excelIF(args...)
		if err != nil {
			t.Fatalf("IF: %v", err)
		}
		return v
	}
	if run(1.0, 10.0, 20.0) != 10.0 {
		t.Fatal("true branch")
	}
	if run(0.0, 10.0, 20.0) != 20.0 {
		t.Fatal("false branch for 0")
	}
	if run(false, "a", "b") != "b" {
		t.Fatal("false bool")
	}
	if run("", 1.0, 2.0) != 2.0 {
		t.Fatal("empty string false")
	}
	if run("0", 1.0, 2.0) != 1.0 {
		t.Fatal("non-empty string true in Excel")
	}
}

func TestExcelIFS(t *testing.T) {
	var lg LogicFuncs
	v, err := lg.excelIFS(false, 1.0, true, 2.0, false, 3.0)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2.0 {
		t.Fatalf("got %v want 2", v)
	}
	_, err = lg.excelIFS(false, 1.0, false, 2.0)
	if err == nil {
		t.Fatal("expect error when no match")
	}
}

func TestExcelSwitch(t *testing.T) {
	var lg LogicFuncs
	v, err := lg.excelSwitch(2.0, 1.0, "a", 2.0, "b", "z")
	if err != nil {
		t.Fatal(err)
	}
	if v != "b" {
		t.Fatalf("got %v want b", v)
	}
	v2, err := lg.excelSwitch(99.0, 1.0, "a", 2.0, "b", "default")
	if err != nil {
		t.Fatal(err)
	}
	if v2 != "default" {
		t.Fatalf("default: got %v", v2)
	}
	_, err = lg.excelSwitch(99.0, 1.0, "a", 2.0, "b")
	if err == nil {
		t.Fatal("expect error without default")
	}
}

func TestExcelCondViaExprCompile(t *testing.T) {
	env := map[string]any{"x": 2.0, "y": 10.0}
	opts := ExprOptions()
	prog, err := expr.Compile(`IF(x > 0, y, 0)`, opts...)
	if err != nil {
		t.Fatal(err)
	}
	out, err := expr.Run(prog, env)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-10) > 1e-9 {
		t.Fatalf("IF expr: got %v", out)
	}

	prog2, err := expr.Compile(`IFS(x < 0, 1, x < 5, 2, true, 3)`, opts...)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := expr.Run(prog2, env)
	if err != nil {
		t.Fatal(err)
	}
	switch v := out2.(type) {
	case float64:
		if v != 2 {
			t.Fatalf("IFS expr: got %v want 2", out2)
		}
	case int:
		if v != 2 {
			t.Fatalf("IFS expr: got %v want 2", out2)
		}
	default:
		t.Fatalf("IFS expr: got %v (%T) want 2", out2, out2)
	}

	prog3, err := expr.Compile(`SWITCH(x, 1.0, "one", 2.0, "two", "other")`, opts...)
	if err != nil {
		t.Fatal(err)
	}
	out3, err := expr.Run(prog3, env)
	if err != nil {
		t.Fatal(err)
	}
	if out3 != "two" {
		t.Fatalf("SWITCH expr: got %v", out3)
	}
}
