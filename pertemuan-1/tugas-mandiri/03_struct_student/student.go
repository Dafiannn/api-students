package main

import "fmt"

// Student merepresentasikan entitas mahasiswa.
type Student struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// GetInfo menggunakan VALUE RECEIVER (s Student)
// Alasan: Method ini hanya membaca data (read-only) untuk diformat menjadi representasi string.
// Method ini tidak melakukan mutasi status apapun pada struct Student, sehingga aman dan efisien
// menggunakan salinan (copy) tanpa risiko efek samping yang tidak diinginkan.
func (s Student) GetInfo() string {
	status := "Non-Aktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("[ID: %d] Nama: %-15s | Nilai: %5.2f | Status: %s", s.ID, s.Name, s.Grade, status)
}

// UpdateGrade menggunakan POINTER RECEIVER (s *Student)
// Alasan: Method ini bertugas mengubah nilai field 'Grade'. Jika menggunakan value receiver,
// nilai Grade yang diubah hanya berada pada salinan struct lokal dan akan hilang begitu method selesai.
// Dengan pointer (*Student), modifikasi langsung dilakukan ke instance asli di memori.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate menggunakan POINTER RECEIVER (s *Student)
// Alasan: Membutuhkan mutasi terhadap field 'IsActive' menjadi true pada instance asli mahasiswa.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate menggunakan POINTER RECEIVER (s *Student)
// Alasan: Membutuhkan mutasi terhadap field 'IsActive' menjadi false pada instance asli mahasiswa.
func (s *Student) Deactivate() {
	s.IsActive = false
}
