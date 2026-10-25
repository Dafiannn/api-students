# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT (PBL)
## MODUL PERTEMUAN 2: REST API & HTTP DEEP DIVE (API MAHASISWA)

**Disusun Oleh:**
- **Nama:** [Nama Mahasiswa]
- **NIM:** [NIM Mahasiswa]
- **Program Studi:** D4 Teknik Informatika
- **Fakultas:** Fakultas Vokasi, Universitas Airlangga
- **Tautan Repositori GitHub:** [https://github.com/username/praktikum-backend-lanjut](https://github.com/username/praktikum-backend-lanjut)

---

## 1. Entitas dan Struktur Berkas

Proyek ini diberi nama `api-students` dengan pembagian berkas modular sesuai instruksi modul:
- `main.go`: Konfigurasi server Fiber, middleware global (`requestid`, `logger`, `cors`), middleware route-level (`requireJSON`), error handler, dan pemetaan routing.
- `model.go`: Mendefinisikan struct `Student`, `CreateStudentRequest`, `ReplaceStudentRequest`, `PatchStudentRequest`, serta amplop response seragam `WebResponse`, `Meta`, dan `ListQuery`.
- `helper.go`: Helper pembungkus response HTTP (`ok`, `okList`, `created`, `noContent`, `fail`, `failValidation`), whitelist sorting `allowedSort`, pembacaan query parameter `parseListQuery`, dan validasi parameter ID `paramID`.
- `handler.go`: Seluruh logika CRUD in-memory mahasiswa: `listStudents`, `getStudent`, `createStudent`, `replaceStudent`, `patchStudent`, dan `deleteStudent`.

### Struktur Folder:
```
api-students/
├── go.mod
├── go.sum
├── main.go
├── model.go
├── helper.go
├── handler.go
└── README.md
```

---

## 2. Tabel Kontrak API

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status HTTP | Keterangan / Contoh Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | - | - | `200 OK` | Pengecekan status server |
| **GET** | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active` | - | `200 OK` | Daftar mahasiswa berpaginasi beserta metadata |
| **GET** | `/api/v1/students/:id` | Path: `:id` (int) | - | `200 OK`<br>`404 Not Found`<br>`400 Bad Request` | Mengambil detail 1 mahasiswa |
| **POST** | `/api/v1/students` | Header: `Content-Type: application/json` | `{"nim":"18221001","name":"Ahmad Fauzi","grade":88.5}` | `201 Created`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | Menambah data baru + header `Location: /api/v1/students/1` |
| **PUT** | `/api/v1/students/:id` | Path: `:id`<br>Header: `Content-Type: application/json` | `{"nim":"18221001","name":"Ahmad Fauzi","grade":92.0,"is_active":false}` | `200 OK`<br>`404 Not Found`<br>`409 Conflict`<br>`422 Unprocessable Entity` | Mengganti seluruh isi entitas secara idempotent |
| **PATCH** | `/api/v1/students/:id` | Path: `:id`<br>Header: `Content-Type: application/json` | `{"grade":95.5}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`422 Unprocessable Entity` | Mengubah sebagian field saja |
| **DELETE** | `/api/v1/students/:id` | Path: `:id` | - | `204 No Content`<br>`404 Not Found`<br>`400 Bad Request` | Menghapus mahasiswa |

---

## 3. Bukti Pengujian Status HTTP

### A. Status 201 Created (POST Berhasil & Header Location)
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"18221001","name":"Ahmad Fauzi","grade":88.5}'
```
**Respons HTTP:**
```http
HTTP/1.1 201 Created
Date: Tue, 20 Oct 2026 02:40:00 GMT
Content-Type: application/json
Content-Length: 178
Location: /api/v1/students/1
X-Request-Id: 9e5cfa8b-03b9-478a-a4ec-1123490abc12

{
  "success": true,
  "message": "mahasiswa berhasil ditambahkan",
  "data": {
    "id": 1,
    "nim": "18221001",
    "name": "Ahmad Fauzi",
    "grade": 88.5,
    "is_active": true,
    "created_at": "2026-10-20T02:40:00.123456+07:00"
  }
}
```

### B. Status 200 OK (GET Berpaginasi & Pengurutan)
```bash
curl -i "http://localhost:3000/api/v1/students?page=1&limit=10&sort=name&order=asc"
```
**Respons HTTP:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "message": "daftar mahasiswa berhasil diambil",
  "data": [
    {
      "id": 1,
      "nim": "18221001",
      "name": "Ahmad Fauzi",
      "grade": 88.5,
      "is_active": true,
      "created_at": "2026-10-20T02:40:00.123456+07:00"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

### C. Status 409 Conflict (NIM Duplikat)
Mencoba menambahkan mahasiswa baru dengan NIM yang sama:
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"18221001","name":"Budi Santoso","grade":75.0}'
```
**Respons HTTP:**
```http
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "success": false,
  "message": "NIM sudah terdaftar dalam sistem"
}
```

### D. Status 415 Unsupported Media Type (Tanpa Header Content-Type)
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -d '{"nim":"18221002","name":"Dewi"}'
```
**Respons HTTP:**
```http
HTTP/1.1 415 Unsupported Media Type
Content-Type: application/json

{
  "success": false,
  "message": "Content-Type harus application/json"
}
```

### E. Status 422 Unprocessable Entity (Validasi Gagal)
```bash
curl -i -X POST http://localhost:3000/api/v1/students \
  -H "Content-Type: application/json" \
  -d '{"nim":"","name":"","grade":150}'
```
**Respons HTTP:**
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "success": false,
  "message": "validasi gagal",
  "errors": {
    "grade": "nilai harus berada di antara 0 dan 100",
    "name": "wajib diisi",
    "nim": "wajib diisi"
  }
}
```

### F. Status 404 Not Found (Data Tidak Ditemukan)
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

### G. Status 204 No Content (DELETE Berhasil)
```bash
curl -i -X DELETE http://localhost:3000/api/v1/students/1
```
**Respons HTTP:**
```http
HTTP/1.1 204 No Content
Date: Tue, 20 Oct 2026 02:42:00 GMT
```
*(Perhatikan tidak ada body JSON yang dikembalikan, sesuai spesifikasi RFC 9110).*

---

## 4. Perbandingan Nyata PUT vs PATCH pada Data yang Sama

Berikut adalah bukti pengujian yang memperlihatkan perbedaan fundamental antara metode **PUT** (mengganti seluruh entitas) dan **PATCH** (hanya memodifikasi field yang ditentukan).

### Tahap 1: Kondisi Awal Data Mahasiswa
```bash
curl -i http://localhost:3000/api/v1/students/1
```
```json
{
  "id": 1,
  "nim": "18221001",
  "name": "Ahmad Fauzi",
  "grade": 88.5,
  "is_active": true,
  "created_at": "2026-10-20T02:40:00+07:00"
}
```

### Tahap 2: Eksekusi PUT (Seluruh Field Diganti)
Pada PUT, seluruh field (`nim`, `name`, `grade`, `is_active`) wajib dikirimkan. Jika tidak dikirimkan, request akan ditolak (422).
```bash
curl -i -X PUT http://localhost:3000/api/v1/students/1 \
  -H "Content-Type: application/json" \
  -d '{"nim":"18221001","name":"Ahmad Fauzi, S.Tr.Kom","grade":94.0,"is_active":false}'
```
**Hasil setelah PUT:**
```json
{
  "id": 1,
  "nim": "18221001",
  "name": "Ahmad Fauzi, S.Tr.Kom",
  "grade": 94.0,
  "is_active": false,
  "created_at": "2026-10-20T02:40:00+07:00"
}
```
*Seluruh field berubah sesuai payload PUT.*

### Tahap 3: Eksekusi PATCH (Hanya Field Tertentu yang Dikirim)
Pada PATCH, klien hanya mengirimkan field `grade` untuk mengoreksi nilai. Field lain (`nim`, `name`, `is_active`) tidak dikirimkan (`nil` pada struct pointer Go) dan dipertahankan apa adanya.
```bash
curl -i -X PATCH http://localhost:3000/api/v1/students/1 \
  -H "Content-Type: application/json" \
  -d '{"grade":98.5}'
```
**Hasil setelah PATCH:**
```json
{
  "id": 1,
  "nim": "18221001",
  "name": "Ahmad Fauzi, S.Tr.Kom",
  "grade": 98.5,
  "is_active": false,
  "created_at": "2026-10-20T02:40:00+07:00"
}
```
*Hanya field `grade` yang berubah dari 94.0 menjadi 98.5. Nama dan status `is_active: false` tetap tidak tersentuh.*

---

## 5. Fitur Query String: Keamanan dan Efisiensi

1. **Pencegahan Denial of Service (DoS) pada Limit:**
   Diatur nilai default limit = 10 dan batas atas maksimum (`q.Limit > 100`) ditetapkan secara tegas di `helper.go`. Hal ini mencegah penyerang mengirimkan parameter seperti `?limit=999999999` yang berpotensi membebani memori RAM server.
2. **Whitelist Sort Kolom:**
   Hanya kolom tertentu (`id`, `nim`, `name`, `grade`) yang diperbolehkan diurutkan melalui map `allowedSort`. Jika klien mencoba mengirimkan kolom arbitrary atau SQL Injection seperti `?sort=passwordadmin`, server secara otomatis menetapkan pengurutan kembali ke default `id`.
