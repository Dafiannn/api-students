# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT (PBL)
## MODUL PERTEMUAN 4: CLEAN ARCHITECTURE (API MAHASISWA)

**Disusun Oleh:**
- **Nama:** [Nama Mahasiswa]
- **NIM:** [NIM Mahasiswa]
- **Program Studi:** D4 Teknik Informatika
- **Fakultas:** Fakultas Vokasi, Universitas Airlangga
- **Tautan Repositori GitHub:** [https://github.com/username/praktikum-backend-lanjut](https://github.com/username/praktikum-backend-lanjut)

---

## 1. Struktur Folder Akhir (Struktur Baku Mata Kuliah)

Seluruh kode proyek mahasiswa (`api-students`) telah disusun ulang secara disiplin mengikuti arsitektur berlapis baku:

```
api-students/
├── app/
│   ├── model/
│   │   └── student.go              <- Layer 1: Entities & DTO Request-Response
│   ├── repository/
│   │   └── student_repository.go   <- Layer 3: Interface Adapters (PostgreSQL Gateway)
│   └── service/
│       ├── student_rules.go        <- Layer 2: Pure Business Rules (Tanpa Fiber)
│       ├── student_rules_test.go   <- Unit Tests Bisnis Murni
│       └── student_service.go      <- Layer 3: Controller & Orkestrasi Use Case
├── config/                         <- Layer 4: Frameworks & Drivers
│   ├── app.go                      <- Konfigurasi Fiber & Assembly
│   ├── env.go                      <- Konfigurasi Environment
│   └── logger.go                   <- Structured Logger (slog + lumberjack)
├── database/                       <- Layer 4: Driver Basis Data (pgxpool)
│   └── postgres.go
├── helper/                         <- Layer 3: Interface Adapters (Presenter)
│   ├── request.go                  <- Query string & params parser
│   └── response.go                 <- Standardized WebResponse envelope
├── logs/                           <- Output Log Terstruktur (Diabaikan oleh Git)
│   └── app.log
├── middleware/                     <- Layer 4: Pipeline Pemeriksaan HTTP
│   └── middleware.go
├── migrations/                     <- Skema DDL Database
│   └── 001_create_students.sql
├── route/                          <- Layer 4: Routing Endpoint
│   └── route.go
├── .env                            <- Konfigurasi Kredensial Lokal (Diabaikan oleh Git)
├── .env.example                    <- Contoh Variabel untuk Tim
├── .gitignore
├── go.mod
├── go.sum
├── main.go                         <- Composition Root & Graceful Shutdown
└── README.md
```

---

## 2. Peta Arsitektur Clean Architecture

### A. Tabel Pemetaan Folder ke Empat Layer
| Layer Clean Architecture | Folder di Proyek Ini | Tanggung Jawab & Batasan |
| :--- | :--- | :--- |
| **Layer 1: Entities** | `app/model/` | Berisi struct `Student` dan format DTO. **Tidak boleh mengimpor package lain dalam proyek sendiri.** |
| **Layer 2: Use Cases** | `app/service/student_rules.go` | Aturan bisnis aplikasi murni: validasi masukan mahasiswa, kalkulasi paginasi, dan isolasi mutasi PATCH. **Tidak mengenal HTTP, database, maupun framework.** |
| **Layer 3: Interface Adapters** | `app/repository/` (Gateway)<br>`helper/` (Presenter)<br>`app/service/student_service.go` (Controller) | Mengubah data dari format luar ke format internal entitas: SQL query PostgreSQL, pembacaan parameter request, pembungkusan amplop JSON response, dan penerjemahan error domain menjadi status HTTP. |
| **Layer 4: Frameworks & Drivers** | `config/`, `database/`, `middleware/`, `route/`, `logs/`, `main.go` | Seluruh detail infrastruktur yang mudah diganti: Web engine Fiber v2, driver `pgx/v5`, rotasi file log `lumberjack`, dan graceful shutdown OS signal. |

### B. Diagram Arah Dependency Antar-Package
Sesuai **Dependency Rule**, panah ketergantungan kode hanya boleh mengalir **ke arah dalam** (menuju inti bisnis). Tidak ada panah yang berbalik ke layer terluar.

```
       [route] ────┐
                   ▼
  [middleware] ──► [app/service] ──► [app/repository] ──► [app/model]
                   ▲ (controller)       (gateway)          (entities)
       [config] ───┘
          ▲
       [main.go] ──► [database] ──► [config]
```

### C. Penjelasan Bagian yang Disederhanakan & Konsekuensinya
1. **Controller Digabung ke dalam Service (`app/service/student_service.go`):**
   - *Konsekuensi:* Business rules rentan terikat dengan `*fiber.Ctx`, sehingga sulit diuji jika tidak dipisahkan.
   - *Mitigasi yang Dilakukan:* Seluruh business rules diekstrak ke dalam fungsi murni di `student_rules.go` yang hanya menerima DTO struct Go dan mengembalikan hasil/peta error tanpa dependensi Fiber. Method di `student_service.go` hanya bertindak sebagai orkestrator tipis (*thin controller*).
2. **Presenter Dilebur ke dalam `helper/`:**
   - *Konsekuensi:* Penentuan kode status HTTP tersebar di beberapa method service alih-alih terpusat pada satu presenter kelas khusus.
   - *Mitigasi:* Menggunakan fungsi `translateError()` terpusat di dalam file service untuk menjamin konsistensi pemetaan error repository ke status HTTP (404, 409, 500).
3. **Interface Repository Disimpan Bersama Implementasinya di `app/repository/`:**
   - *Penyederhanaan:* Secara teori Clean Architecture murni, interface `StudentRepository` didefinisikan di layer Use Cases (`service`) dan implementasinya di layer Adapters (`repository`). Keduanya disatukan di package `repository` untuk menjaga agar jumlah file package Go tidak berlebihan (*avoid package explosion*).

### D. Opini Kritis: Apakah Struktur Ini Sepadan?
> **Pendapat Mandiri:**
> Untuk proyek CRUD sederhana seukuran tugas ini (yang hanya mengelola satu entitas mahasiswa), struktur Clean Architecture berlapis ini sepintas terasa menghasilkan *boilerplate* dan jumlah berkas yang cukup banyak (*over-engineering*). Namun, pemisahan ini **mulai terasa sangat sepadan ketika sistem mulai tumbuh ke ukuran menengah hingga besar**, khususnya saat:
> 1. Fitur otentikasi (JWT) dan otorisasi role-based access control (RBAC) ditambahkan pada modul lanjutan.
> 2. Penambahan database kedua (seperti MongoDB untuk caching atau Redis).
> 3. Kebutuhan pengujian otomatis (*automated unit & integration testing*) skala enterprise, di mana aturan validasi bisnis dapat diuji dalam 0,02 detik tanpa memerlukan emulator database sungguhan.
> Dengan demikian, investasi arsitektur sejak awal mencegah hutang teknis (*technical debt*) yang mahal saat sistem berkembang.

---

## 3. Pemisahan Business Rules Murni & Hasil Unit Test

Berkas `app/service/student_rules.go` berisi aturan bisnis yang sepenuhnya terisolasi dari dunia luar. Untuk membuktikan kehandalan aturan bisnis ini, telah dibuat unit test komprehensif pada `app/service/student_rules_test.go`:

### Perintah Pengujian:
```bash
go test ./app/service/... -v
```

### Bukti Keluaran Eksekusi Unit Test:
```text
=== RUN   TestCountTotalPages
=== RUN   TestCountTotalPages/Total_0_Data
=== RUN   TestCountTotalPages/Total_1_Data_Limit_10
=== RUN   TestCountTotalPages/Total_10_Data_Limit_10_(Pas_1_Halaman)
=== RUN   TestCountTotalPages/Total_11_Data_Limit_10_(Membulat_ke_2_Halaman)
=== RUN   TestCountTotalPages/Total_137_Data_Limit_20_(Membulat_ke_7_Halaman)
=== RUN   TestCountTotalPages/Limit_0_Menghindari_Division_By_Zero
--- PASS: TestCountTotalPages (0.00s)
=== RUN   TestValidateCreate
=== RUN   TestValidateCreate/Valid_Input_Lolos_Validasi
=== RUN   TestValidateCreate/NIM_Kosong_dan_Nilai_Negatif_Ditolak
=== RUN   TestValidateCreate/NIM_Terlalu_Pendek_Ditolak
--- PASS: TestValidateCreate (0.00s)
=== RUN   TestValidateReplace
=== RUN   TestValidateReplace/Valid_Replace_Lolos_Validasi
=== RUN   TestValidateReplace/Nama_Kosong_pada_PUT_Ditolak
--- PASS: TestValidateReplace (0.00s)
=== RUN   TestApplyPatch
--- PASS: TestApplyPatch (0.00s)
=== RUN   TestIsEmptyPatch
--- PASS: TestIsEmptyPatch (0.00s)
PASS
ok      api-students/app/service    0.794s
```
*Seluruh unit test lulus 100% tanpa menyalakan server HTTP dan tanpa menyentuh database.*

---

## 4. Logger Terstruktur dan File Log

Menggunakan pustaka `log/slog` bawaan Go dan pustaka rotasi log `gopkg.in/natefinch/lumberjack.v2`:
- Log ditulis secara ganda (*MultiWriter*): ke konsol layar (*stdout*) untuk pengawasan real-time dan ke berkas `logs/app.log`.
- Berkas dirotasi secara otomatis bila mencapai ukuran 10 MB, dengan penyimpanan cadangan maksimal 5 berkas lama dan retensi 14 hari.

### Potongan Isi File Log (`logs/app.log`):
```json
{"time":"2026-10-20T03:30:15.102+07:00","level":"INFO","msg":"server berjalan","port":"3000"}
{"time":"2026-10-20T03:30:22.451+07:00","level":"INFO","msg":"http_request","request_id":"5f9b4c12-32a1-4e78-98f3-8b7c4a1129bc","method":"GET","path":"/api/v1/students","status":200,"duration":2450100,"ip":"127.0.0.1"}
{"time":"2026-10-20T03:30:45.892+07:00","level":"INFO","msg":"http_request","request_id":"c7e19d44-098e-4a66-bbf2-1149e89d10ae","method":"POST","path":"/api/v1/students","status":201,"duration":4890200,"ip":"127.0.0.1"}
{"time":"2026-10-20T03:31:02.115+07:00","level":"INFO","msg":"http_request","request_id":"a1820df3-102c-47bc-857e-328b9cc0981a","method":"POST","path":"/api/v1/students","status":409,"duration":1890000,"ip":"127.0.0.1"}
```

---

## 5. Pemeriksaan Kebocoran Layer (Layer Leak Verification)

Telah dilakukan verifikasi integritas arsitektur terhadap 5 kriteria pengujian kebocoran layer:

| Komponen yang Diperiksa | Persyaratan Modul | Hasil Pemeriksaan Nyata | Status |
| :--- | :--- | :--- | :--- |
| **Import pada `app/model`** | Tidak ada satu pun package dari proyek sendiri | Hasil `go list -deps ./app/model`: hanya memuat pustaka standar Go (`time`) | **Lolos (Bersih)** |
| **Import pada `app/repository`** | Tidak ada `gofiber` sama sekali | Hasil `go list -deps ./app/repository \| grep gofiber`: Kosong (Nol dependensi Fiber) | **Lolos (Bersih)** |
| **Isi package `app/service`** | Tidak ada satu pun perintah SQL | Hasil regex `SELECT\|INSERT\|UPDATE\|DELETE`: Tidak ditemukan query SQL di service | **Lolos (Bersih)** |
| **Isi file `route/route.go`** | Tidak ada `if` untuk validasi / business rules | Hanya ada pemetaan URL ke method service dan pengecekan koneksi ping di health | **Lolos (Bersih)** |
| **Isi file `main.go`** | Tidak ada handler; hanya urutan perakitan | Berfungsi murni sebagai *Composition Root* & *Graceful Shutdown* | **Lolos (Bersih)** |

### Perintah Pembuktian:
```bash
# Memeriksa dependensi repository tidak bocor ke Fiber:
go list -deps ./app/repository | Select-String "gofiber"
# (Output kosong menandakan tidak ada referensi ke framework web Fiber)

# Memeriksa kompilasi seluruh paket:
go build ./...
go vet ./...
# (Exit code 0, nol error, nol peringatan)
```
