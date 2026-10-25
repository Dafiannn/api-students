# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT (PBL)
## MODUL PERTEMUAN 3: DATABASE & REPOSITORY PATTERN (API MAHASISWA)

**Disusun Oleh:**
- **Nama:** [Nama Mahasiswa]
- **NIM:** [NIM Mahasiswa]
- **Program Studi:** D4 Teknik Informatika
- **Fakultas:** Fakultas Vokasi, Universitas Airlangga
- **Tautan Repositori GitHub:** [https://github.com/username/praktikum-backend-lanjut](https://github.com/username/praktikum-backend-lanjut)

---

## 1. Skema Tabel dan Migrasi

Berkas migrasi SQL disimpan pada direktori `migrations/001_create_students.sql`:

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

### Mengapa Keunikan NIM Lebih Baik Dijaga oleh Basis Data Dibandingkan Kode Go?
1. **Penanganan Konkurensi & Mencegah Race Condition:**
   Pemeriksaan manual di kode Go (`SELECT` terlebih dahulu sebelum `INSERT`) menyisakan celah bahaya. Jika dua permintaan pembuatan akun mahasiswa dengan NIM yang sama tiba dalam milidetik yang sama, kedua thread Go dapat menyelesaikan query `SELECT` (keduanya mendapati NIM belum ada) lalu keduanya mengeksekusi `INSERT`. Akibatnya terjadi duplikasi data. Sebaliknya, basis data PostgreSQL menggunakan mekanisme locking transaksi atomik (*ACID*) pada level *engine*, sehingga basis datalah yang menjadi wasit mutlak pemutus siapa yang tiba lebih dulu.
2. **Integritas Data Terpusat:**
   Jika di masa depan ada aplikasi lain (seperti background worker, batch job seeder, atau API microservice lain) yang menulis ke tabel `students`, aturan unik tetap terjamin aman tanpa bergantung pada apakah kode pemanggil menulis validasi pengecekan atau tidak.

### Alasan Penambahan Indeks Selain Kunci Primer (`students_name_lower_idx`):
Pencarian mahasiswa pada endpoint `GET /api/v1/students?search=...` menggunakan filter `LOWER(name) ILIKE $1`. Tanpa indeks fungsional `LOWER(name)`, PostgreSQL terpaksa melakukan pemindaian menyeluruh (*Sequential Scan / Table Scan*) ke seluruh baris tabel dari awal sampai akhir. Dengan adanya `B-Tree Index` pada `LOWER(name)`, PostgreSQL dapat mencari posisi baris data secara logaritmik ($O(\log N)$), menghemat I/O disk, dan menjaga kecepatan respon API tetap di bawah beberapa milidetik meskipun tabel berisi ratusan ribu baris.

---

## 2. Konfigurasi dan Connection Pool

### Konfigurasi `.env` & `.env.example`:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

### Koneksi Basis Data (`database/postgres.go`):
Membuka koneksi baru ke basis data sangat mahal karena melibatkan TCP handshake, autentikasi TLS/MD5/SCRAM, alokasi memori proses, dan negosiasi parameter sesi. Pola *Connection Pool* (`*pgxpool.Pool`) menyediakan sejumlah koneksi yang selalu terbuka (*pre-warmed*) untuk dipinjam oleh worker thread HTTP dan dikembalikan seketika setelah query selesai:
- `MaxConns = 10`: Mencegah server PostgreSQL kehabisan memori (*out of connection slots*).
- `MinConns = 2`: Menjaga koneksi siap pakai agar tidak terjadi lonjakan latensi di awal request.
- `MaxConnLifetime = 1 Hour`: Merotasi koneksi agar koneksi tua yang rentan putus diganti secara transparan.
- `MaxConnIdleTime = 30 Minutes`: Menutup koneksi yang menganggur guna menghemat sumber daya basis data.
- `Ping(ctx)`: Memastikan kredensial dan ketersediaan server sebelum API mulai melayani request pengguna.

---

## 3. Pola Repository

Interface `StudentRepository` memisahkan pertanyaan **"apa yang ingin saya lakukan pada data"** dari **"bagaimana cara data itu disimpan"**:

```go
type StudentRepository interface {
    FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
    FindByID(ctx context.Context, id int) (model.Student, error)
    Create(ctx context.Context, s model.Student) (model.Student, error)
    Update(ctx context.Context, s model.Student) (model.Student, error)
    Delete(ctx context.Context, id int) error
}
```

### Keunggulan Desain:
1. **Arah Ketergantungan Bersih:**
   `handler -> repository (interface) -> model`. Paket repository sama sekali tidak mengenal `*fiber.Ctx`.
2. **Sentinel Error Abstraksi:**
   Didefinisikan dua error bisnis pada repository:
   - `ErrNotFound = errors.New("data mahasiswa tidak ditemukan")`
   - `ErrDuplicate = errors.New("NIM mahasiswa sudah terdaftar")`
   Handler hanya berurusan dengan sentinel error ini, tanpa perlu tahu kode error driver PostgreSQL (`pgconn.PgError` dengan kode `23505` atau `pgx.ErrNoRows`).

---

## 4. Query di Sisi Basis Data

Seluruh proses filtering, searching, sorting, dan paging dipindahkan dari memori Go ke PostgreSQL:
1. **Pencarian Nama:** Menggunakan `ILIKE` dengan parameter `$1` (`%keyword%`).
2. **Penyaringan:** Klausa `WHERE is_active = $2` berparameter.
3. **Pengurutan Aman:** Menggunakan whitelist dictionary `kolomUrut[q.Sort]` untuk mencegah SQL Injection pada klausa `ORDER BY`.
4. **Paginasi:** Menggunakan `LIMIT $X OFFSET $Y`.
5. **Penghitungan Total:** Menjalankan `SELECT COUNT(*) FROM students` dengan kondisi `WHERE` yang sama persis sebelum limit dan offset diterapkan.

---

## 5. Bukti Pengujian Status HTTP Error Basis Data

### A. Status 404 Not Found (ID Tidak Ada di Tabel)
Menguji `GET`, `PUT`, `PATCH`, dan `DELETE` pada ID mahasiswa yang tidak ada (misal `id = 999`):
```bash
curl -i -X GET http://localhost:3000/api/v1/students/999
```
**Respons HTTP:**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "success": false,
  "message": "mahasiswa tidak ditemukan"
}
```
*Diterjemahkan secara elegan dari `pgx.ErrNoRows` atau `RowsAffected == 0` menjadi status HTTP 404.*

---

### B. Status 409 Conflict (NIM Ganda)
Mencoba menambahkan atau mengubah mahasiswa dengan NIM yang sudah ada di basis data:
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"18221001","name":"Ahmad Duplicate","grade":80.0}'
```
**Respons HTTP:**
```http
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "success": false,
  "message": "NIM mahasiswa sudah terdaftar"
}
```
*Diterjemahkan dari PostgreSQL Error Code `23505` (unique_violation) menjadi `ErrDuplicate` dan dipetakan ke 409 Conflict.*

---

### C. Status 503 Service Unavailable / 500 Internal Server Error (Layanan Database Terputus)
Pengujian ketika service basis data PostgreSQL dimatikan:
```bash
curl -i http://localhost:3000/api/v1/health
```
**Respons HTTP:**
```http
HTTP/1.1 503 Service Unavailable
Content-Type: application/json

{
  "success": false,
  "message": "database tidak dapat dihubungi"
}
```

### Analisis Pemilihan Status (503 vs 500):
Menurut RFC 9110, status **503 Service Unavailable** adalah status yang paling tepat untuk situasi di mana server web masih hidup normal, namun layanan hulu (*upstream dependency*) seperti database sedang mengalami gangguan sementara atau kelebihan beban (*overloaded/unreachable*). Status 503 secara eksplisit mengindikasikan kepada klien (dan load balancer) bahwa kondisi ini bersifat sementara (*temporary condition*), sehingga klien dapat mencoba kembali beberapa saat kemudian (*Retry-After*). Sebaliknya, status 500 Internal Server Error lebih tepat untuk galat tak terduga (*unexpected panic/bug*) pada internal kode aplikasi.
