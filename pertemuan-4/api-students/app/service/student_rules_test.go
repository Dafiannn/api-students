package service

import (
	"testing"
	"api-students/app/model"
)

// TestCountTotalPages menguji logika pembagian dan pembulatan halaman secara tabel (Table-Driven Tests)
func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		name         string
		total, limit int
		want         int
	}{
		{"Total 0 Data", 0, 10, 0},
		{"Total 1 Data Limit 10", 1, 10, 1},
		{"Total 10 Data Limit 10 (Pas 1 Halaman)", 10, 10, 1},
		{"Total 11 Data Limit 10 (Membulat ke 2 Halaman)", 11, 10, 2},
		{"Total 137 Data Limit 20 (Membulat ke 7 Halaman)", 137, 20, 7},
		{"Limit 0 Menghindari Division By Zero", 50, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CountTotalPages(tc.total, tc.limit)
			if got != tc.want {
				t.Errorf("total=%d limit=%d: diharapkan %d, didapatkan %d", tc.total, tc.limit, tc.want, got)
			}
		})
	}
}

// TestValidateCreate menguji aturan bisnis validasi pembuatan data mahasiswa baru
func TestValidateCreate(t *testing.T) {
	t.Run("Valid Input Lolos Validasi", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:   "18221001",
			Name:  "Budi Santoso",
			Grade: 85.5,
		}
		errs := ValidateCreate(req)
		if len(errs) != 0 {
			t.Fatalf("seharusnya lolos validasi tanpa error, didapatkan: %v", errs)
		}
	})

	t.Run("NIM Kosong dan Nilai Negatif Ditolak", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:   "",
			Name:  "Adit",
			Grade: -5.0,
		}
		errs := ValidateCreate(req)
		if _, ok := errs["nim"]; !ok {
			t.Error("seharusnya menghasilkan error validasi pada field nim")
		}
		if _, ok := errs["grade"]; !ok {
			t.Error("seharusnya menghasilkan error validasi pada field grade")
		}
	})

	t.Run("NIM Terlalu Pendek Ditolak", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:   "123",
			Name:  "Siti",
			Grade: 90.0,
		}
		errs := ValidateCreate(req)
		if _, ok := errs["nim"]; !ok {
			t.Error("seharusnya menolak NIM dengan panjang kurang dari 5 karakter")
		}
	})
}

// TestValidateReplace menguji validasi ketat pada metode PUT
func TestValidateReplace(t *testing.T) {
	t.Run("Valid Replace Lolos Validasi", func(t *testing.T) {
		req := model.ReplaceStudentRequest{
			NIM:      "18221002",
			Name:     "Dewi Lestari",
			Grade:    95.0,
			IsActive: true,
		}
		errs := ValidateReplace(req)
		if len(errs) != 0 {
			t.Fatalf("seharusnya lolos validasi PUT tanpa error, didapatkan: %v", errs)
		}
	})

	t.Run("Nama Kosong pada PUT Ditolak", func(t *testing.T) {
		req := model.ReplaceStudentRequest{
			NIM:      "18221002",
			Name:     "   ",
			Grade:    95.0,
			IsActive: true,
		}
		errs := ValidateReplace(req)
		if _, ok := errs["name"]; !ok {
			t.Error("seharusnya menghasilkan error validasi nama pada PUT")
		}
	})
}

// TestApplyPatch menguji penerapan pembaruan sebagian tanpa mempengaruhi field lain
func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "18221001",
		Name:     "Ahmad Fauzi",
		Grade:    80.0,
		IsActive: true,
	}

	newGrade := 92.5
	newStatus := false
	patchReq := model.PatchStudentRequest{
		Grade:    &newGrade,
		IsActive: &newStatus,
	}

	result, errs := ApplyPatch(initial, patchReq)
	if len(errs) != 0 {
		t.Fatalf("tidak boleh ada error pada patch yang valid, didapatkan: %v", errs)
	}

	// Verifikasi nilai yang dimodifikasi
	if result.Grade != 92.5 {
		t.Errorf("grade seharusnya terupdate menjadi 92.5, didapat: %.2f", result.Grade)
	}
	if result.IsActive != false {
		t.Error("is_active seharusnya terupdate menjadi false")
	}

	// Verifikasi nilai yang TIDAK dikirim tidak mengalami perubahan
	if result.NIM != "18221001" {
		t.Errorf("NIM seharusnya tetap '18221001', didapat: %s", result.NIM)
	}
	if result.Name != "Ahmad Fauzi" {
		t.Errorf("Name seharusnya tetap 'Ahmad Fauzi', didapat: %s", result.Name)
	}
}

// TestIsEmptyPatch menguji deteksi payload kosong pada PATCH
func TestIsEmptyPatch(t *testing.T) {
	emptyReq := model.PatchStudentRequest{}
	if !IsEmptyPatch(emptyReq) {
		t.Error("seharusnya mendeteksi bahwa request patch kosong")
	}

	nim := "18221003"
	nonEmptyReq := model.PatchStudentRequest{NIM: &nim}
	if IsEmptyPatch(nonEmptyReq) {
		t.Error("seharusnya mendeteksi bahwa request patch TIDAK kosong")
	}
}
