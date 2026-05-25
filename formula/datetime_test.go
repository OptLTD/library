package formula

import (
	"math"
	"testing"
	"time"
)

func TestDurationDaysHoursMinutes(t *testing.T) {
	var dt DateTimeFuncs
	d := 48 * time.Hour
	out, err := dt.durationDays(d)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(out.(float64)-2) > 1e-9 {
		t.Fatalf("days: got %v", out)
	}
	h, err := dt.durationHours(90 * time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(h.(float64)-1.5) > 1e-9 {
		t.Fatalf("hours: got %v", h)
	}
	min, err := dt.durationMinutes(2*time.Hour + 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(min.(float64)-150) > 1e-9 {
		t.Fatalf("minutes: got %v", min)
	}
}

func TestNowReturnsTime(t *testing.T) {
	var dt DateTimeFuncs
	v, err := dt.now()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.(time.Time); !ok {
		t.Fatalf("want time.Time, got %T", v)
	}
}
