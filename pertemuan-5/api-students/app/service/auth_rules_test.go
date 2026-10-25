package service

import (
	"testing"
	"api-students-v5/app/model"
)

func TestCheckPasswordStrength(t *testing.T) {
	cases := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Password Terlalu Pendek", "abc1", false},
		{"Password Tanpa Angka", "hanyahurufaja", false},
		{"Password Tanpa Huruf", "1234567890", false},
		{"Password Terlalu Umum", "password123", false},
		{"Password Kuat dan Sah", "MahasiswaS1TI2026!", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := CheckPasswordStrength(tc.password)
			if tc.valid && msg != "" {
				t.Errorf("seharusnya valid, tetapi dapat error: %s", msg)
			}
			if !tc.valid && msg == "" {
				t.Errorf("seharusnya ditolak, tetapi lolos: %s", tc.password)
			}
		})
	}
}

func TestValidateRegister(t *testing.T) {
	t.Run("Valid Registration", func(t *testing.T) {
		req := model.RegisterRequest{
			Username: "andi_ti",
			Email:    "andi@student.unair.ac.id",
			Password: "KuatPassword999",
		}
		errs := ValidateRegister(req)
		if len(errs) != 0 {
			t.Fatalf("seharusnya lolos, dapat: %v", errs)
		}
	})

	t.Run("Username Tidak Valid", func(t *testing.T) {
		req := model.RegisterRequest{
			Username: "an di!",
			Email:    "andi@student.unair.ac.id",
			Password: "KuatPassword999",
		}
		errs := ValidateRegister(req)
		if _, ok := errs["username"]; !ok {
			t.Error("seharusnya menolak username dengan spasi atau karakter khusus")
		}
	})
}
