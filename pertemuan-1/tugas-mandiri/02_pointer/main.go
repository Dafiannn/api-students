package main

import "fmt"

// swap menerima dua buah pointer ke integer (*int),
// lalu menukar nilai yang berada di alamat memori tersebut secara langsung.
func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

// swapByValue adalah contoh fungsi pass-by-value biasa.
// Fungsi ini hanya menerima salinan nilai integer dari pemanggil.
// Perubahan di dalam fungsi ini TIDAK akan mengubah variabel di pemanggil.
func swapByValue(a, b int) {
	temp := a
	a = b
	b = temp
}

// updateSlice menerima pointer ke slice (*[]string) dan item baru string.
// Penggunaan pointer di sini sangat penting karena ketika fungsi append dipanggil,
// kapasitas slice lama mungkin terlampaui sehingga Go mengalokasikan array dasar baru di memori
// dan memperbarui header slice (pointer data, len, cap). Tanpa pointer (*[]string),
// header baru hasil append hanya disimpan di salinan lokal fungsi dan hilang setelah kembali ke pemanggil.
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// updateSliceByValue adalah contoh fungsi yang menerima slice secara pass-by-value.
// Menambahkan elemen dengan append di sini tidak akan mengubah slice di fungsi pemanggil
// jika pemanggil mengandalkan slice header yang sama.
func updateSliceByValue(s []string, newItem string) {
	s = append(s, newItem)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("             TUGAS 1.3: POINTER DI GO             ")
	fmt.Println("==================================================")

	// --- Bagian 1: Perbandingan Swap (Pass by Value vs Pass by Pointer) ---
	fmt.Println("\n>>> DEMONSTRASI 1: SWAP DUA INTEGER")
	val1, val2 := 10, 99
	fmt.Printf("[Kondisi Awal]           : val1 = %d, val2 = %d\n", val1, val2)

	// Percobaan 1: Pass by Value
	swapByValue(val1, val2)
	fmt.Printf("[Sesudah swapByValue]    : val1 = %d, val2 = %d (TIDAK BERUBAH)\n", val1, val2)

	// Percobaan 2: Pass by Pointer
	swap(&val1, &val2)
	fmt.Printf("[Sesudah swap(&val1, &val2)]: val1 = %d, val2 = %d (BERHASIL DITUKAR!)\n", val1, val2)

	// --- Bagian 2: Demonstrasi updateSlice dengan Pointer ---
	fmt.Println("\n>>> DEMONSTRASI 2: UPDATESLICE DENGAN POINTER")
	hobi := []string{"Membaca", "Coding"}
	fmt.Printf("[Slice Awal]                   : %v (len: %d, cap: %d)\n", hobi, len(hobi), cap(hobi))

	// Percobaan 1: Pass by Value (updateSliceByValue)
	updateSliceByValue(hobi, "Berenang")
	fmt.Printf("[Sesudah updateSliceByValue]   : %v (item baru 'Berenang' TIDAK masuk)\n", hobi)

	// Percobaan 2: Pass by Pointer (*[]string)
	updateSlice(&hobi, "Desain Grafis")
	updateSlice(&hobi, "Bermain Musik")
	fmt.Printf("[Sesudah updateSlice (*pointer)]: %v (len: %d, cap: %d) (BERHASIL DITAMBAHKAN!)\n", hobi, len(hobi), cap(hobi))

	// --- Bagian 3: Penjelasan Teknis Perbedaan Pass by Value vs Pass by Pointer ---
	fmt.Println("\n==================================================")
	fmt.Println("PENJELASAN TEKNIS:")
	fmt.Println("1. Pass by Value:")
	fmt.Println("   Go membuat salinan (copy) dari nilai variabel ke stack frame fungsi baru.")
	fmt.Println("   Setiap mutasi hanya berdampak pada variabel salinan lokal dan dibuang saat fungsi selesai.")
	fmt.Println("2. Pass by Pointer:")
	fmt.Println("   Go mengirimkan alamat memori (&) tempat data asli disimpan.")
	fmt.Println("   Melalui dereferensi (*), fungsi dapat memanipulasi langsung isi memori asli.")
	fmt.Println("3. Kenapa updateSlice butuh *[]string?")
	fmt.Println("   Slice header terdiri dari pointer array, len, dan cap.")
	fmt.Println("   Saat append memicu realokasi memori baru, alamat pointer internal dan nilai len berubah.")
	fmt.Println("   Tanpa pointer ke struct slice itu sendiri, perubahan header ini tidak terdengar oleh pemanggil.")
	fmt.Println("==================================================")
}
