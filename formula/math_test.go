package formula

import (
	"math"
	"testing"
)

func TestArgToFloat64(t *testing.T) {
	v, err := argToFloat64(float64(3.5))
	if err != nil || v != 3.5 {
		t.Fatalf("float64: %v %v", v, err)
	}
	v, err = argToFloat64(int64(7))
	if err != nil || v != 7 {
		t.Fatalf("int64: %v %v", v, err)
	}
	v, err = argToFloat64("2.25")
	if err != nil || math.Abs(v-2.25) > 1e-9 {
		t.Fatalf("string: %v %v", v, err)
	}
	_, err = argToFloat64(struct{}{})
	if err == nil {
		t.Fatal("expect error for struct")
	}
}

func TestSliceToFloat64_nil(t *testing.T) {
	s, err := sliceToFloat64(nil)
	if err != nil || s != nil || len(s) != 0 {
		t.Fatalf("nil slice: %v %v", s, err)
	}
}

func TestTruncFormat(t *testing.T) {
	var m MathFuncs
	v, err := m.truncFormat(1.239, 2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v-1.23) > 1e-9 {
		t.Fatalf("got %v", v)
	}
}

func TestCeilFloorDecimal(t *testing.T) {
	var m MathFuncs
	c, err := m.CeilDecimal(1.231, 2)
	if err != nil || math.Abs(c-1.24) > 1e-9 {
		t.Fatalf("ceil: %v %v", c, err)
	}
	fl, err := m.FloorDecimal(1.239, 2)
	if err != nil || math.Abs(fl-1.23) > 1e-9 {
		t.Fatalf("floor: %v %v", fl, err)
	}
}

func TestMarginalProgressive(t *testing.T) {
	var m MathFuncs
	run := func(qty float64, upper []float64, prices []float64) float64 {
		v, err := m.marginal(qty, upper, prices)
		if err != nil {
			t.Fatalf("marginal: %v", err)
		}
		return v.(float64)
	}

	if g := run(12, nil, []float64{2.5}); math.Abs(g-30) > 1e-9 {
		t.Fatalf("flat: got %v want 30", g)
	}

	if g := run(350, []float64{200, 400}, []float64{0.5, 0.6, 0.8}); math.Abs(g-190) > 1e-9 {
		t.Fatalf("350: got %v want 190", g)
	}

	if g := run(500, []float64{200, 400}, []float64{0.5, 0.6, 0.8}); math.Abs(g-300) > 1e-9 {
		t.Fatalf("500: got %v want 300", g)
	}

	if g := run(200, []float64{200, 400}, []float64{0.5, 0.6, 0.8}); math.Abs(g-100) > 1e-9 {
		t.Fatalf("200: got %v want 100", g)
	}

	if g := run(0, []float64{200}, []float64{1, 2}); g != 0 {
		t.Fatalf("0 qty: got %v", g)
	}
}

func TestMarginalViaBuild(t *testing.T) {
	out, err := Build(`MARGINAL(350, [200, 400], [0.5, 0.6, 0.8])`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-190) > 1e-6 {
		t.Fatalf("Build marginal: got %v want 190", out)
	}
}

func TestRoundCeilingFloorViaBuild(t *testing.T) {
	cases := []struct {
		code string
		want float64
	}{
		{`ROUND(2.4)`, 2},
		{`ROUND(2.6)`, 3},
		{`ROUND(1.2345, 2)`, 1.23},
		{`ROUND(1234, -2)`, 1200},
		{`CEILING(2.1)`, 3},
		{`CEILING(4.1, 2)`, 6},
		{`FLOOR(2.9)`, 2},
		{`FLOOR(4.1, 2)`, 4},
	}
	for _, c := range cases {
		out, err := Build(c.code, nil)
		if err != nil {
			t.Fatalf("%s: %v", c.code, err)
		}
		if math.Abs(out.(float64)-c.want) > 1e-6 {
			t.Fatalf("%s: got %v want %v", c.code, out, c.want)
		}
	}
}

func TestMarginalRejectsZeroInUpper(t *testing.T) {
	var m MathFuncs
	_, err := m.marginal(100.0, []any{0, 200}, []any{1.0, 2.0, 3.0})
	if err == nil {
		t.Fatal("expected error when upper contains 0")
	}
}

func TestSum_scalarAndNumericStrings(t *testing.T) {
	var m MathFuncs
	v, err := m.sum(1.0, 2.0, "3.5")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v.(float64)-6.5) > 1e-9 {
		t.Fatalf("got %v", v)
	}
}
