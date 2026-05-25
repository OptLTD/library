package formula

import (
	"math"
	"testing"
)

// 模拟业务 Combine：扁平点号键（basic.opt1）经 FormulaEnv 后，公式里用 basic.opt1 访问。

func TestBuild_IFS_nestedDotPaths(t *testing.T) {
	env := FormulaEnv(map[string]any{
		"basic.opt1": float64(3),
		"basic.opt2": float64(0),
		"other.flag": true,
	})
	// 第一个成立条件：basic.opt1 > 1 → 100
	out, err := Build(`IFS(basic.opt1 > 1, 100, basic.opt1 > 0, 50, true, 0)`, env)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-100) > 1e-6 {
		t.Fatalf("got %v want 100", out)
	}

	env2 := FormulaEnv(map[string]any{
		"basic.opt1": float64(0.5),
	})
	out2, err := Build(`IFS(basic.opt1 > 1, 100, basic.opt1 > 0, 50, true, 0)`, env2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out2.(float64)-50) > 1e-6 {
		t.Fatalf("got %v want 50", out2)
	}

	env3 := FormulaEnv(map[string]any{
		"basic.opt1": float64(-1),
	})
	out3, err := Build(`IFS(basic.opt1 > 1, 100, basic.opt1 > 0, 50, true, 0)`, env3)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out3.(float64)-0) > 1e-6 {
		t.Fatalf("got %v want 0", out3)
	}
}

func TestBuild_IF_nestedDotPaths(t *testing.T) {
	env := FormulaEnv(map[string]any{
		"basic.opt1": float64(2),
		"basic.opt2": float64(1),
	})
	out, err := Build(`IF(basic.opt1 > 1, IF(basic.opt2 > 0, 99, 0), -1)`, env)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-99) > 1e-6 {
		t.Fatalf("got %v want 99", out)
	}

	env2 := FormulaEnv(map[string]any{
		"basic.opt1": float64(0),
		"basic.opt2": float64(5),
	})
	out2, err := Build(`IF(basic.opt1 > 1, IF(basic.opt2 > 0, 99, 0), -1)`, env2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out2.(float64)-(-1)) > 1e-6 {
		t.Fatalf("got %v want -1", out2)
	}
}

func TestBuild_IFS_mixedNestedMapAndDotKeys(t *testing.T) {
	// 已有 basic 对象，再用 basic.x 合并字段（与 ToCombine 常见形态一致）
	env := FormulaEnv(map[string]any{
		"basic": map[string]any{"opt0": 1.0},
		"basic.opt1": 4.0,
	})
	out, err := Build(`IFS(basic.opt1 > basic.opt0 + 2, 1, true, 0)`, env)
	if err != nil {
		t.Fatal(err)
	}
	// 4 > 1+2 → 1
	if math.Abs(out.(float64)-1) > 1e-6 {
		t.Fatalf("got %v want 1", out)
	}
}

func TestBuild_lowercase_ifs_aliasWithNestedPath(t *testing.T) {
	env := FormulaEnv(map[string]any{"basic.opt1": 2.0})
	out, err := Build(`ifs(basic.opt1 > 1, 10, true, 0)`, env)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-10) > 1e-6 {
		t.Fatalf("got %v want 10", out)
	}
}

func TestBuild_IF_returnsStringValue(t *testing.T) {
	env := FormulaEnv(map[string]any{
		"basic.a": 1.0,
		"basic.b": 1.0,
	})
	out, err := Build(`IF(basic.a==1&&basic.b==1,"DOUBLE","SINGLE")`, env)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := out.(string)
	if !ok {
		t.Fatalf("expected string result, got %T (%v)", out, out)
	}
	if got != "DOUBLE" {
		t.Fatalf("got %q want %q", got, "DOUBLE")
	}
}
