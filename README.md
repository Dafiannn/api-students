# Repositori Praktikum Pemrograman Backend Lanjut (PBL)
**Program Studi D4 Teknik Informatika — Fakultas Vokasi, Universitas Airlangga**

Repositori ini memuat implementasi lengkap tugas mandiri dari **Modul Pertemuan 1 hingga Pertemuan 4** secara bertahap, modular, dan teruji.

---

## 📂 Struktur Direktori Proyek

```
praktikum-backend-lanjut/
├── pertemuan-1/                 # MODUL 1: Persiapan Lingkungan & Sintaks Go dengan Fiber
│   ├── tugas-mandiri/
│   │   ├── 01_variabel_struktur_data/main.go  <- Demo 5 variabel & operasi map mahasiswa
│   │   ├── 02_pointer/main.go                 <- Demo swap & updateSlice dengan pointer
│   │   └── 03_struct_student/                 <- Struct Student (Value vs Pointer Receiver)
│   │       ├── student.go
│   │       └── main.go
│   └── laporan/
│       └── LAPORAN_TUGAS_1.md                 <- Draf laporan lengkap Tugas 1
│
├── pertemuan-2/                 # MODUL 2: REST API & HTTP Deep Dive (In-Memory)
│   ├── api-students/
│   │   ├── handler.go           <- CRUD handler mahasiswa in-memory
│   │   ├── helper.go            <- WebResponse envelope, parser query, status HTTP
│   │   ├── model.go             <- Student struct, Request DTO (POST, PUT, PATCH)
│   │   ├── main.go              <- Fiber server & routing
│   │   └── README.md            <- Dokumen Kontrak API Mahasiswa
│   └── laporan/
│       └── LAPORAN_TUGAS_2.md                 <- Draf laporan lengkap Tugas 2 (PUT vs PATCH, Status codes)
│
├── pertemuan-3/                 # MODUL 3: Database & Repository Pattern (PostgreSQL)
│   ├── api-students/
│   │   ├── app/
│   │   │   ├── model/student.go               <- Student entity & ListQuery Offset()
│   │   │   └── repository/student_repository.go <- StudentRepository interface & PostgreSQL pgxpool
│   │   ├── config/env.go                      <- Parser environment .env
│   │   ├── database/postgres.go               <- Connection pool pgxpool.Pool & Ping check
│   │   ├── migrations/001_create_students.sql <- Skema DDL tabel & UNIQUE index
│   │   ├── handler.go                         <- Handler berbasis repository & error mapper
│   │   ├── helper.go
│   │   ├── main.go                            <- Perakitan basis data & health check
│   │   ├── .env.example
│   │   └── README.md                          <- Panduan basis data & migrasi
│   └── laporan/
│       └── LAPORAN_TUGAS_3.md                 <- Draf laporan lengkap Tugas 3
│
└── pertemuan-4/                 # MODUL 4: Clean Architecture (Struktur Baku Mata Kuliah)
    ├── api-students/
    │   ├── app/
    │   │   ├── model/student.go               <- Layer 1: Entities (Core Business Rules)
    │   │   ├── repository/student_repository.go <- Layer 3: Adapters (Gateway PostgreSQL)
    │   │   └── service/
    │   │       ├── student_rules.go           <- Layer 2: Pure Business Rules (Tanpa Fiber)
    │   │       ├── student_rules_test.go      <- Unit Tests murni (Table-driven tests)
    │   │       └── student_service.go         <- Layer 3: Controller & Use Case
    │   ├── config/
    │   │   ├── app.go                         <- Fiber app assembly & custom error handler
    │   │   ├── env.go                         <- Environment loader
    │   │   └── logger.go                      <- Structured JSON logger (slog + lumberjack)
    │   ├── database/postgres.go               <- Connection pool
    │   ├── helper/
    │   │   ├── request.go                     <- RequestContext, ParamID, ParseListQuery
    │   │   └── response.go                    <- Success, Created, Fail, FailValidation
    │   ├── logs/                              <- Berkas log rotasi (di-gitignore)
    │   ├── middleware/middleware.go           <- Security headers, RequestLogger, RequireJSON
    │   ├── migrations/001_create_students.sql
    │   ├── route/route.go                     <- Pemetaan URL endpoint murni
    │   ├── main.go                            <- Composition root & Graceful shutdown
    │   ├── .env.example
    │   └── README.md
    └── laporan/
        └── LAPORAN_TUGAS_4.md                 <- Draf laporan lengkap Tugas 4 (Peta arsitektur, kebocoran)
```

---

## 🚀 Ringkasan Cara Menjalankan Tiap Modul

### Modul 1 (Sintaks Go & Pointer & Struct):
```bash
# Uji Variabel dan Map:
cd pertemuan-1/tugas-mandiri/01_variabel_struktur_data
go run main.go

# Uji Pointer:
cd ../02_pointer
go run main.go

# Uji Struct Student:
cd ../03_struct_student
go run .
```

### Modul 2 (API Mahasiswa In-Memory):
```bash
cd pertemuan-2/api-students
go run .
```

### Modul 3 (API Mahasiswa PostgreSQL & Repository):
```bash
cd pertemuan-3/api-students
cp .env.example .env
# Jalankan migrasi SQL di PostgreSQL:
# psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
go run .
```

### Modul 4 (API Mahasiswa Clean Architecture & Unit Test):
```bash
cd pertemuan-4/api-students

# 1. Jalankan Unit Test (100% lolos):
go test ./app/service/... -v

# 2. Jalankan Build & Static Analysis:
go build ./...
go vet ./...

# 3. Jalankan Server:
go run .
```

---

## 📑 Format Berkas Laporan

Laporan lengkap untuk setiap modul telah disusun rapi dalam berkas markdown yang siap diekspor menjadi format PDF dengan penamaan berkas resmi:
1. `pertemuan-1/laporan/LAPORAN_TUGAS_1.md` -> diekspor menjadi **`Tugas1_NIM.pdf`**
2. `pertemuan-2/laporan/LAPORAN_TUGAS_2.md` -> diekspor menjadi **`Tugas2_NIM.pdf`**
3. `pertemuan-3/laporan/LAPORAN_TUGAS_3.md` -> diekspor menjadi **`Tugas3_NIM.pdf`**
4. `pertemuan-4/laporan/LAPORAN_TUGAS_4.md` -> diekspor menjadi **`Tugas4_NIM.pdf`**
