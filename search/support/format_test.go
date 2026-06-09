package support

import "testing"

func TestCoerceNumber(t *testing.T) {
	tests := []struct {
		in   any
		want float64
		ok   bool
	}{
		{0, 0, true},
		{int(0), 0, true},
		{float64(0), 0, true},
		{100, 100, true},
		{"0.5", 0.5, true},
		{"", 0, false},
		{nil, 0, false},
	}
	for _, tc := range tests {
		got, ok := CoerceNumber(tc.in)
		if ok != tc.ok || (tc.ok && got != tc.want) {
			t.Fatalf("CoerceNumber(%v) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
