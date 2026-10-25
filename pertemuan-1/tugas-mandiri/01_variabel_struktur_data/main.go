package main

import (
	"fmt"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  TUGAS 1.2: VARIABEL DAN STRUKTUR DATA (GO)")
	fmt.Println("==================================================")

	// 1. Deklarasi 5 variabel dengan tipe berbeda: string, int, float64, bool, dan slice
	var studentName string = "Muhammad Bintang"
	var studentAge int = 20
	var studentGPA float64 = 3.85
	var isEnrolled bool = true
	var courseList []string = []string{"Pemrograman Backend Lanjut", "Basis Data", "Jaringan Komputer"}

	fmt.Println("\n--- 1. Deklarasi 5 Variabel Berbeda Tipe ---")
	fmt.Printf("1. String  (studentName) : %s\n", studentName)
	fmt.Printf("2. Integer (studentAge)  : %d tahun\n", studentAge)
	fmt.Printf("3. Float64 (studentGPA)  : %.2f\n", studentGPA)
	fmt.Printf("4. Boolean (isEnrolled)  : %t\n", isEnrolled)
	fmt.Printf("5. Slice   (courseList)  : %v (panjang: %d, kapasitas: %d)\n", courseList, len(courseList), cap(courseList))

	// 2. Map untuk menyimpan data mahasiswa (Nama sebagai kunci, Nilai float64 sebagai isinya)
	fmt.Println("\n--- 2. Map Data Mahasiswa ---")
	studentGrades := make(map[string]float64)

	// Operasi a: Menambah data ke dalam map
	studentGrades["Budi Pratama"] = 87.5
	studentGrades["Siti Aisyah"] = 92.0
	studentGrades["Rian Hidayat"] = 78.5
	studentGrades["Dewi Lestari"] = 85.0
	fmt.Println("Data mahasiswa berhasil ditambahkan ke dalam map.")

	// Operasi b: Membaca data dengan pengecekan keberadaan (two-value assignment)
	fmt.Println("\n--- Pengecekan Keberadaan Data (Read with Existence Check) ---")
	targetStudent := "Siti Aisyah"
	if grade, exists := studentGrades[targetStudent]; exists {
		fmt.Printf("Status: Ditemukan! Nilai %s adalah %.2f\n", targetStudent, grade)
	} else {
		fmt.Printf("Status: Mahasiswa atas nama %s TIDAK ditemukan dalam sistem.\n", targetStudent)
	}

	targetNotFound := "Andi Kurniawan"
	if grade, exists := studentGrades[targetNotFound]; exists {
		fmt.Printf("Status: Ditemukan! Nilai %s adalah %.2f\n", targetNotFound, grade)
	} else {
		fmt.Printf("Status: Mahasiswa atas nama '%s' TIDAK ditemukan dalam sistem.\n", targetNotFound)
	}

	// Operasi c: Menghapus data dari map
	fmt.Println("\n--- Menghapus Data Mahasiswa dari Map ---")
	deleteTarget := "Rian Hidayat"
	fmt.Printf("Menghapus data mahasiswa: %s...\n", deleteTarget)
	delete(studentGrades, deleteTarget)

	// Pastikan data telah terhapus
	if _, exists := studentGrades[deleteTarget]; !exists {
		fmt.Printf("Verifikasi: Data %s berhasil dihapus dari map.\n", deleteTarget)
	}

	// Operasi d: Menelusuri seluruh isi map (Iterasi dengan range)
	fmt.Println("\n--- Daftar Lengkap Seluruh Mahasiswa dalam Map ---")
	nomor := 1
	for name, grade := range studentGrades {
		fmt.Printf("%d. Nama: %-15s | Nilai Akhir: %.2f\n", nomor, name, grade)
		nomor++
	}
	fmt.Println("==================================================")
}
