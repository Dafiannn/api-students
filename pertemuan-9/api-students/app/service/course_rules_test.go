package service

import "testing"

func TestWithinCreditLimit(t *testing.T) {
	tests := []struct {
		name   string
		taken  int
		adding int
		want   bool
	}{
		{"bawah batas", 18, 3, true},
		{"pas 24 SKS", 21, 3, true},
		{"melebihi batas", 22, 3, false},
		{"nol SKS", 0, 3, true},
		{"langsung 25 SKS", 0, 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinCreditLimit(tt.taken, tt.adding); got != tt.want {
				t.Errorf("WithinCreditLimit(%d, %d) = %v; want %v", tt.taken, tt.adding, got, tt.want)
			}
		})
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	tests := []struct {
		name    string
		taken   int
		leaving int
		joining int
		want    int
	}{
		{"swap sama sks", 20, 3, 3, 20},
		{"swap dari 3 ke 2 sks", 20, 3, 2, 19},
		{"swap dari 2 ke 4 sks", 20, 2, 4, 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreditsAfterSwap(tt.taken, tt.leaving, tt.joining); got != tt.want {
				t.Errorf("CreditsAfterSwap(%d, %d, %d) = %d; want %d", tt.taken, tt.leaving, tt.joining, got, tt.want)
			}
		})
	}
}
