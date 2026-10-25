# API Mahasiswa (api-students) - Pertemuan 4: Clean Architecture

Implementasi RESTful API Mahasiswa menerapkan prinsip **Clean Architecture (Robert C. Martin)**, pemisahan tanggung jawab berlapis (*separation of concerns*), *inversion of control*, pencatatan log terstruktur (*structured logging* dengan `log/slog` & `lumberjack`), serta pengujian unit terisolasi (*pure unit tests*).

---

## 1. Struktur Folder Baku Mata Kuliah

```
api-students/
├── app/
│   ├── model/                  # Layer 1: Entities (Struct model & request-response)
│   ├── repository/             # Layer 3: Interface Adapters - Gateway (SQL Query PostgreSQL)
│   └── service/                # Layer 2: Use Cases & Controller
│       ├── student_rules.go    # Pure Business Rules (Validasi & Patch tanpa Fiber)
│       ├── student_rules_test.go # Unit Test Bisnis Murni
│       └── student_service.go  # Controller Fiber & Orkestrasi
├── config/                     # Layer 4: Frameworks & Drivers (Konfigurasi & Logger)
│   ├── app.go                  # Perakitan instance Fiber, middleware, dan route
│   ├── env.go                  # Parser environment (.env)
│   └── logger.go               # Logger terstruktur slog dengan rotasi lumberjack
├── database/                   # Layer 4: Koneksi & Pool PostgreSQL (pgx/v5)
│   └── postgres.go
├── helper/                     # Layer 3: Presenter (WebResponse & request parser)
│   ├── request.go
│   └── response.go
├── logs/                       # Folder output log berkas (di-gitignore)
│   └── app.log
├── middleware/                 # Layer 4: Middleware pipeline (Security, Log, RequireJSON)
│   └── middleware.go
├── migrations/                 # Skema migrasi basis data SQL
│   └── 001_create_students.sql
├── route/                      # Layer 4: Pemetaan alamat URL endpoint
│   └── route.go
├── .env                        # Konfigurasi lokal rahasia (di-gitignore)
├── .env.example                # Templat konfigurasi rekan tim
├── go.mod
├── go.sum
├── main.go                     # Composition Root & Graceful Shutdown
└── README.md
```

---

## 2. Pemetaan Empat Layer Clean Architecture

| Layer Clean Architecture | Folder di Proyek Ini | Tanggung Jawab Utama |
| :--- | :--- | :--- |
| **1. Entities** | `app/model/` | Aturan bisnis inti (*core business rules*). Tidak boleh mengimpor package lain dalam proyek. |
| **2. Use Cases** | `app/service/student_rules.go` | Aturan bisnis aplikasi (*pure business rules*): validasi input, kalkulasi paginasi, penerapan patch. Bebas dari ketergantungan HTTP/Fiber. |
| **3. Interface Adapters** | `app/repository/` (Gateway)<br>`helper/` (Presenter)<br>`app/service/student_service.go` (Controller) | Penerjemah antara representasi luar dengan model internal usecase: SQL driver, format JSON HTTP, dan pembacaan parameter. |
| **4. Frameworks & Drivers** | `config/`, `database/`, `middleware/`, `route/`, `logs/`, `main.go` | Alat bantu teknis, web framework Fiber, koneksi basis data, file I/O logging, dan rute. |

---

## 3. Menjalankan Unit Test Terisolasi

Salah satu keunggulan terbesar Clean Architecture adalah **Business Rules dapat diuji dalam hitungan milidetik** tanpa perlu koneksi database aktif atau server web:

```bash
go test ./app/service/... -v
```

Output:
```text
=== RUN   TestCountTotalPages
--- PASS: TestCountTotalPages (0.00s)
=== RUN   TestValidateCreate
--- PASS: TestValidateCreate (0.00s)
=== RUN   TestValidateReplace
--- PASS: TestValidateReplace (0.00s)
=== RUN   TestApplyPatch
--- PASS: TestApplyPatch (0.00s)
=== RUN   TestIsEmptyPatch
--- PASS: TestIsEmptyPatch (0.00s)
PASS
ok      api-students/app/service    0.024s
```

---

## 4. Cara Menjalankan Server API

1. Pastikan PostgreSQL telah siap dan berkas migrasi `migrations/001_create_students.sql` telah dieksekusi.
2. Salin dan sesuaikan konfigurasi environment:
   ```bash
   cp .env.example .env
   ```
3. Unduh seluruh dependensi:
   ```bash
   go mod tidy
   ```
4. Jalankan aplikasi:
   ```bash
   go run .
   ```
5. Log aktivitas request HTTP akan dicetak ke konsol dan tersimpan otomatis ke `logs/app.log` berformat JSON terstruktur.
