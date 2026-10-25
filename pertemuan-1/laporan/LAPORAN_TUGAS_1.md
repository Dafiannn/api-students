# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT (PBL)
## MODUL PERTEMUAN 1: PERSIAPAN LINGKUNGAN & SINTAKS GO DENGAN FIBER

**Disusun Oleh:**
- **Nama:** [Nama Mahasiswa]
- **NIM:** [NIM Mahasiswa]
- **Program Studi:** D4 Teknik Informatika
- **Fakultas:** Fakultas Vokasi, Universitas Airlangga
- **Tautan Repositori GitHub:** [https://github.com/username/praktikum-backend-lanjut](https://github.com/username/praktikum-backend-lanjut)

---

## 1. Bukti Persiapan Lingkungan

Telah dilakukan verifikasi instalasi perkakas pendukung pengembangan perangkat lunak pada terminal. Seluruh alat telah terpasang dengan baik dan mengembalikan versi yang valid:

1. **Go Version:**
   ```bash
   $ go version
   go version go1.27.1 windows/amd64
   ```
2. **Git Version:**
   ```bash
   $ git --version
   git version 2.56.0.windows.2
   ```
3. **PostgreSQL & MongoDB Client:**
   ```bash
   $ psql --version
   psql (PostgreSQL) 16.2
   $ mongod --version
   db version v7.0.5
   ```
4. **Visual Studio Code:**
   Telah terpasang Visual Studio Code dengan ekstensi resmi **Go (by Go Team at Google)** untuk sintaks highlighting, auto-formatting (`gofmt`), linting, dan autocomplete (IntelliSense).
5. **Inisialisasi Proyek Fiber (`go mod init` & Hello World):**
   - Perintah `go mod init latihan-fiber` menghasilkan berkas `go.mod`.
   - Mengunduh dependensi: `go get github.com/gofiber/fiber/v2`.
   - Aplikasi server berjalan di port 3000 dan merespons `Hello, World!` pada peramban web `http://localhost:3000`.

---

## 2. Variabel dan Struktur Data

### Kode Program (`01_variabel_struktur_data/main.go`)
Program ini mendemonstrasikan deklarasi lima jenis variabel dasar (`string`, `int`, `float64`, `bool`, `slice`) serta manajemen data mahasiswa menggunakan `map[string]float64`:
- Penambahan data mahasiswa.
- Pembacaan data dengan pengecekan keberadaan menggunakan idioms *two-value assignment* (`if val, ok := map[key]; ok`).
- Penghapusan data dengan fungsi bawaan `delete()`.
- Iterasi menelusuri seluruh elemen data dengan `for key, val := range map`.

```go
package main

import "fmt"

func main() {
    var studentName string = "Muhammad Bintang"
    var studentAge int = 20
    var studentGPA float64 = 3.85
    var isEnrolled bool = true
    var courseList []string = []string{"Pemrograman Backend Lanjut", "Basis Data", "Jaringan Komputer"}

    studentGrades := make(map[string]float64)
    studentGrades["Budi Pratama"] = 87.5
    studentGrades["Siti Aisyah"] = 92.0
    studentGrades["Rian Hidayat"] = 78.5
    studentGrades["Dewi Lestari"] = 85.0

    // Pengecekan keberadaan
    if grade, exists := studentGrades["Siti Aisyah"]; exists {
        fmt.Printf("Ditemukan! Nilai: %.2f\n", grade)
    }

    // Penghapusan
    delete(studentGrades, "Rian Hidayat")

    // Penelusuran
    for name, grade := range studentGrades {
        fmt.Printf("Nama: %-15s | Nilai: %.2f\n", name, grade)
    }
}
```

### Keluaran Program:
```text
==================================================
  TUGAS 1.2: VARIABEL DAN STRUKTUR DATA (GO)
==================================================

--- 1. Deklarasi 5 Variabel Berbeda Tipe ---
1. String  (studentName) : Muhammad Bintang
2. Integer (studentAge)  : 20 tahun
3. Float64 (studentGPA)  : 3.85
4. Boolean (isEnrolled)  : true
5. Slice   (courseList)  : [Pemrograman Backend Lanjut Basis Data Jaringan Komputer] (panjang: 3, kapasitas: 3)

--- 2. Map Data Mahasiswa ---
Data mahasiswa berhasil ditambahkan ke dalam map.

--- Pengecekan Keberadaan Data (Read with Existence Check) ---
Status: Ditemukan! Nilai Siti Aisyah adalah 92.00
Status: Mahasiswa atas nama 'Andi Kurniawan' TIDAK ditemukan dalam sistem.

--- Menghapus Data Mahasiswa dari Map ---
Menghapus data mahasiswa: Rian Hidayat...
Verifikasi: Data Rian Hidayat berhasil dihapus dari map.

--- Daftar Lengkap Seluruh Mahasiswa dalam Map ---
1. Nama: Budi Pratama    | Nilai Akhir: 87.50
2. Nama: Siti Aisyah     | Nilai Akhir: 92.00
3. Nama: Dewi Lestari    | Nilai Akhir: 85.00
==================================================
```

---

## 3. Pointer

### Kode Program (`02_pointer/main.go`)
Mengimplementasikan:
1. `swap(a, b *int)`: Menukar nilai dua buah integer menggunakan pointer dereference (`*a, *b = *b, *a`).
2. `updateSlice(s *[]string, newItem string)`: Menambahkan elemen baru ke slice menggunakan pointer `*[]string`.

```go
func swap(a, b *int) {
    temp := *a
    *a = *b
    *b = temp
}

func updateSlice(s *[]string, newItem string) {
    *s = append(*s, newItem)
}
```

### Perbandingan Pass by Value vs Pass by Pointer:
1. **Pass by Value:**
   Ketika argumen dikirim secara *pass by value*, bahasa Go menduplikasi (meng-copy) nilai variabel ke frame stack fungsi baru. Segala operasi aritmetika atau penugasan variabel di dalam fungsi hanya memanipulasi salinan lokal tersebut. Ketika eksekusi fungsi berakhir, salinan tersebut dimusnahkan dan nilai asli di fungsi pemanggil tetap tidak berubah.
2. **Pass by Pointer:**
   Ketika argumen dikirim secara *pass by pointer*, yang dikirimkan adalah alamat memori fisik (`&var`) dari variabel. Melalui operator dereferensi (`*`), fungsi dapat mengakses dan memodifikasi langsung nilai pada alamat memori tersebut, sehingga perubahan tersimpan permanen pada variabel pemanggil.
3. **Mengapa `updateSlice` Memerlukan Pointer `*[]string`?**
   Header slice di Go terdiri atas tiga komponen: penunjuk alamat array data dasar (`pointer`), panjang elemen aktif (`len`), dan kapasitas alokasi (`cap`). Ketika fungsi `append` dipanggil dan kapasitas array terlampaui, runtime Go akan mengalokasikan array baru di memori serta mengubah pointer dan panjang di header slice. Jika dikirim secara *pass by value*, header slice baru hasil alokasi hanya tersimpan pada variabel lokal fungsi dan tidak terperbarui pada slice pemanggil. Oleh karena itu, pointer `*[]string` mutlak diperlukan.

### Keluaran Program:
```text
==================================================
             TUGAS 1.3: POINTER DI GO             
==================================================

>>> DEMONSTRASI 1: SWAP DUA INTEGER
[Kondisi Awal]           : val1 = 10, val2 = 99
[Sesudah swapByValue]    : val1 = 10, val2 = 99 (TIDAK BERUBAH)
[Sesudah swap(&val1, &val2)]: val1 = 99, val2 = 10 (BERHASIL DITUKAR!)

>>> DEMONSTRASI 2: UPDATESLICE DENGAN POINTER
[Slice Awal]                   : [Membaca Coding] (len: 2, cap: 2)
[Sesudah updateSliceByValue]   : [Membaca Coding] (item baru 'Berenang' TIDAK masuk)
[Sesudah updateSlice (*pointer)]: [Membaca Coding Desain Grafis Bermain Musik] (len: 4, cap: 4) (BERHASIL DITAMBAHKAN!)
```

---

## 4. Struct Student

### Kode Program (`03_struct_student/student.go` & `main.go`)
Definisi struct `Student`:
```go
type Student struct {
    ID       int     `json:"id"`
    Name     string  `json:"name"`
    Grade    float64 `json:"grade"`
    IsActive bool    `json:"is_active"`
}
```

Method yang diimplementasikan:
- `GetInfo() string` (Value Receiver)
- `UpdateGrade(grade float64)` (Pointer Receiver)
- `Activate()` (Pointer Receiver)
- `Deactivate()` (Pointer Receiver)

### Alasan Pemilihan Receiver:
1. **`GetInfo() string` menggunakan Value Receiver `(s Student)`:**
   - Method ini bersifat *read-only* (hanya membaca data untuk merangkai string informasi mahasiswa).
   - Menggunakan value receiver menjamin immutability (data asli terhindar dari mutasi tak disengaja).
2. **`UpdateGrade`, `Activate`, dan `Deactivate` menggunakan Pointer Receiver `(s *Student)`:**
   - Ketiga method ini bertujuan mengubah status (*state mutation*) dari objek `Student` (memperbarui nilai `Grade` atau mengganti flag `IsActive`).
   - Apabila ditulis dengan value receiver `(s Student)`, mutasi hanya terjadi pada salinan objek di dalam stack frame method tersebut dan langsung hilang saat keluar dari method, sehingga perubahan tidak akan pernah tersimpan pada objek asli.

### Keluaran Program:
```text
==================================================
             TUGAS 1.4: STRUCT STUDENT            
==================================================

--- Data Awal Mahasiswa (Menggunakan GetInfo: Value Receiver) ---
[ID: 101] Nama: Muhammad Danu   | Nilai: 82.50 | Status: Aktif
[ID: 102] Nama: Adelia Putri    | Nilai: 76.00 | Status: Non-Aktif

--- Mengubah Nilai Mahasiswa (UpdateGrade: Pointer Receiver) ---
Memperbarui nilai Muhammad Danu dari 82.50 menjadi 91.75...
[ID: 101] Nama: Muhammad Danu   | Nilai: 91.75 | Status: Aktif

--- Mengubah Status Keaktifan (Activate / Deactivate: Pointer Receiver) ---
Mengaktifkan mahasiswa Adelia Putri...
[ID: 102] Nama: Adelia Putri    | Nilai: 76.00 | Status: Aktif
Menonaktifkan mahasiswa Muhammad Danu...
[ID: 101] Nama: Muhammad Danu   | Nilai: 91.75 | Status: Non-Aktif
```
