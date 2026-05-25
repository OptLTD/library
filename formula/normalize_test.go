package formula

import "testing"

func TestNormalizeExcelIFCalls_replacesStandaloneIf(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`if(1>0, 1, 0)`, `IF(1>0, 1, 0)`},
		{`if (1>0, 1, 0)`, `IF(1>0, 1, 0)`},
		{`If(1>0, 1, 0)`, `IF(1>0, 1, 0)`},
		{`IF(1>0, 1, 0)`, `IF(1>0, 1, 0)`},
		{`IFS(a,b,c,d)`, `IFS(a,b,c,d)`},
	}
	for _, tc := range cases {
		got := NormalizeExcelIFCalls(tc.in)
		if got != tc.want {
			t.Fatalf("in %q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeExcelIFCalls_doesNotTouchSumIfAndSimilar(t *testing.T) {
	unchanged := []string{
		`SUMIF(range, crit, sum_range)`,
		`sumif(x > 0, 1, 2)`,
		`COUNTIF(a, b)`,
		`verify(something)`,
		`elseif(true, 1, 0)`, // 若将来注册 elseif，整段仍不会被改成 elseIF(
	}
	for _, s := range unchanged {
		got := NormalizeExcelIFCalls(s)
		if got != s {
			t.Fatalf("must not change %q, got %q", s, got)
		}
	}
}

func TestNormalizeExcelIFCalls_nestedIfCalls(t *testing.T) {
	in := `if(a>0, if(b>0, 1, 2), 3)`
	want := `IF(a>0, IF(b>0, 1, 2), 3)`
	if got := NormalizeExcelIFCalls(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
