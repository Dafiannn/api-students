# API Mahasiswa (api-students) - Pertemuan 3: Database & Repository Pattern

Aplikasi REST API mahasiswa yang menyimpan data secara persisten ke dalam basis data PostgreSQL menggunakan pola arsitektur **Repository Pattern** dan driver performa tinggi **pgx/v5**.

---

## 1. Skema Tabel Basis Data

Skema tabel mahasiswa didefinisikan pada berkas `migrations/001_create_students.sql`:

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks unik untuk NIM (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks penunjang pencarian nama mahasiswa
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

### Mengapa Keunikan NIM Dijaga di Sisi Basis Data?
Pemeriksaan keunikan dengan perulangan kode Go atau query `SELECT` sebelum `INSERT` memiliki kelemahan kritis: **Race Condition**. Bila dua permintaan HTTP dengan NIM yang sama tiba secara bersamaan (*concurrent*), keduanya dapat lolos validasi `SELECT` sebelum salah satunya sempat melakukan `INSERT`. Dengan indeks `UNIQUE` di sisi basis data, PostgreSQL bertindak sebagai otoritas tunggal yang menjamin integritas atomik (*ACID*) dan menolak duplikasi secara mutlak.

---

## 2. Cara Menyiapkan Basis Data dari Nol

1. Masuk ke terminal PostgreSQL / psql:
   ```bash
   psql -U postgres
   ```
2. Buat database baru bernama `praktikum_backend`:
   ```sql
   CREATE DATABASE praktikum_backend;
   \q
   ```
3. Jalankan berkas migrasi:
   ```bash
   psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
   ```
4. Verifikasi struktur tabel yang terbentuk:
   ```bash
   psql -U postgres -d praktikum_backend -c "\d students"
   ```

---

## 3. Daftar Variabel Environment

Salin berkas `.env.example` menjadi `.env` lalu sesuaikan dengan kredensial PostgreSQL Anda:
```bash
cp .env.example .env
```

Isi variabel konfigurasi:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=rahasia_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

> **Catatan Keamanan:** Berkas `.env` telah didaftarkan pada `.gitignore` dan tidak boleh di-commit ke repositori Git publik.

---

## 4. Cara Menjalankan Aplikasi

1. Unduh modul dependensi:
   ```bash
   go mod init api-students-v3
   go get github.com/gofiber/fiber/v2
   go get github.com/jackc/pgx/v5
   go get github.com/joho/godotenv
   go mod tidy
   ```
2. Jalankan server:
   ```bash
   go run .
   ```
3. Uji endpoint kesehatan (*health check*):
   ```bash
   curl -i http://localhost:3000/api/v1/health
   ```
