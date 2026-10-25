package main

import "fmt"

func main() {
	fmt.Println("==================================================")
	fmt.Println("             TUGAS 1.4: STRUCT STUDENT            ")
	fmt.Println("==================================================")

	// Inisialisasi mahasiswa
	mhs1 := Student{
		ID:       101,
		Name:     "Muhammad Danu",
		Grade:    82.5,
		IsActive: true,
	}

	mhs2 := Student{
		ID:       102,
		Name:     "Adelia Putri",
		Grade:    76.0,
		IsActive: false,
	}

	fmt.Println("\n--- Data Awal Mahasiswa (Menggunakan GetInfo: Value Receiver) ---")
	fmt.Println(mhs1.GetInfo())
	fmt.Println(mhs2.GetInfo())

	fmt.Println("\n--- Mengubah Nilai Mahasiswa (UpdateGrade: Pointer Receiver) ---")
	fmt.Println("Memperbarui nilai Muhammad Danu dari 82.50 menjadi 91.75...")
	mhs1.UpdateGrade(91.75)
	fmt.Println(mhs1.GetInfo())

	fmt.Println("\n--- Mengubah Status Keaktifan (Activate / Deactivate: Pointer Receiver) ---")
	fmt.Println("Mengaktifkan mahasiswa Adelia Putri...")
	mhs2.Activate()
	fmt.Println(mhs2.GetInfo())

	fmt.Println("Menonaktifkan mahasiswa Muhammad Danu...")
	mhs1.Deactivate()
	fmt.Println(mhs1.GetInfo())

	fmt.Println("\n==================================================")
	fmt.Println("PENJELASAN VALUE RECEIVER VS POINTER RECEIVER:")
	fmt.Println("1. GetInfo() -> Value Receiver (s Student)")
	fmt.Println("   - Hanya membaca informasi dan tidak mengubah isi struct.")
	fmt.Println("   - Menjamin sifat immutability sehingga data asli terlindungi dari perubahan tak sengaja.")
	fmt.Println("2. UpdateGrade(), Activate(), Deactivate() -> Pointer Receiver (s *Student)")
	fmt.Println("   - Perlu memodifikasi state internal struct (Grade dan IsActive).")
	fmt.Println("   - Dengan pointer (*), mutasi data terjadi langsung pada memori objek yang bersangkutan.")
	fmt.Println("==================================================")
}
