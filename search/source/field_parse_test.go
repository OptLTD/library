package source

import (
	"testing"
	"time"
)

func TestParseTimeOnlyMonthYYYYMM(t *testing.T) {
	f := Field{}
	f.Extra.DataType = "ONLYMONTH"

	got := f.ParseTime("2026-07")
	if got == nil {
		t.Fatal("ParseTime(YYYY-MM) returned nil")
	}
	if got.Year() != 2026 || got.Month() != time.July || got.Day() != 1 {
		t.Fatalf("ParseTime(YYYY-MM) = %v, want 2026-07-01", got)
	}

	got = f.ParseTime("2026-07-01")
	if got == nil {
		t.Fatal("ParseTime(YYYY-MM-DD) returned nil")
	}
	if got.Year() != 2026 || got.Month() != time.July || got.Day() != 1 {
		t.Fatalf("ParseTime(YYYY-MM-DD) = %v, want 2026-07-01", got)
	}
}

func TestParseTimeOnlyDateAlsoAcceptsMonthViaFallback(t *testing.T) {
	f := Field{}
	f.Extra.DataType = "ONLYDATE"
	// ParseDate 兼容 YYYY-MM 后，ONLYDATE 也会落到月初
	got := f.ParseTime("2026-07")
	if got == nil {
		t.Fatal("expected YYYY-MM fallback")
	}
	if got.Year() != 2026 || got.Month() != time.July {
		t.Fatalf("got %v", got)
	}
}
