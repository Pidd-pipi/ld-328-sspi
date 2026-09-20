package util

import (
	"testing"
	"time"
)

func TestFoodCalculator_RemainingDays(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	yesterday := time.Now().AddDate(0, 0, -1)
	calc := NewFoodCalculator()
	tests := []struct {
		name       string
		expiry     *time.Time
		wantRemain int
	}{
		{name: "tomorrow", expiry: &tomorrow, wantRemain: 1},
		{name: "yesterday", expiry: &yesterday, wantRemain: -1},
		{name: "nil", expiry: nil, wantRemain: 3650},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.RemainingDays(tt.expiry); got != tt.wantRemain {
				t.Fatalf("RemainingDays() = %d, want %d", got, tt.wantRemain)
			}
		})
	}
}

func TestFoodCalculator_ComputeFreshness(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1)
	tenDays := time.Now().AddDate(0, 0, 10)
	yesterday := time.Now().AddDate(0, 0, -1)
	calc := NewFoodCalculator()
	tests := []struct {
		name   string
		status string
		expiry *time.Time
		want   string
	}{
		{name: "consumed wins", status: "consumed", expiry: &tenDays, want: "consumed"},
		{name: "expiring within 3 days", status: "fresh", expiry: &tomorrow, want: "expiring"},
		{name: "fresh far away", status: "fresh", expiry: &tenDays, want: "fresh"},
		{name: "expired", status: "fresh", expiry: &yesterday, want: "expired"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calc.ComputeFreshness(tt.status, tt.expiry); got != tt.want {
				t.Fatalf("ComputeFreshness() = %s, want %s", got, tt.want)
			}
		})
	}
}
