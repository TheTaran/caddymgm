package main

import (
	"testing"
	"time"
)

func TestSecurityTrendSpecForPeriod(t *testing.T) {
	tests := []struct {
		period    string
		canonical string
		window    time.Duration
		bucket    time.Duration
	}{
		{"", "1d", 24 * time.Hour, time.Hour},
		{"1h", "1h", time.Hour, 5 * time.Minute},
		{"6h", "6h", 6 * time.Hour, 15 * time.Minute},
		{"1d", "1d", 24 * time.Hour, time.Hour},
		{"7d", "7d", 7 * 24 * time.Hour, 6 * time.Hour},
		{"30d", "30d", 30 * 24 * time.Hour, 24 * time.Hour},
	}
	for _, test := range tests {
		period, spec, err := securityTrendSpecForPeriod(test.period)
		if err != nil || period != test.canonical || spec.window != test.window || spec.bucketDuration != test.bucket {
			t.Fatalf("period %q: got period=%q spec=%+v err=%v", test.period, period, spec, err)
		}
	}
	if _, _, err := securityTrendSpecForPeriod("2h"); err == nil {
		t.Fatal("expected an invalid period error")
	}
}
