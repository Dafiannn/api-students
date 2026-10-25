# API Mahasiswa (api-students) - Pertemuan 2

REST API untuk pengelolaan data mahasiswa menggunakan Golang dan framework Fiber v2 dengan penyimpanan in-memory.

## Dokumen Kontrak API

Berikut adalah tabel kontrak API mahasiswa sesuai dengan standar RESTful API:

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status HTTP | Contoh Respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/health` | - | - | `200 OK` | `{"success": true, "message": "server berjalan", "data": {"timestamp": "..."}}` |
| **GET** | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active` | - | `200 OK` | `{"success": true, "message": "daftar mahasiswa berhasil diambil", "data": [...], "meta": {"page": 1, "limit": 10, "total": 1, "total_pages": 1}}` |
| **GET** | `/api/v1/students/:id` | Path: `:id` (integer) | - | `200 OK`<br>`404 Not Found`<br>`400 Bad Request` | `{"success": true, "message": "data mahasiswa ditemukan", "data": {"id": 1, "nim": "18221001", "name": "Budi", "grade": 85.5, "is_active": true}}` |
| **POST** | `/api/v1/students` | Header: `Content-Type: application/json` | `{"nim": "18221001", "name": "Budi", "grade": 85.5}` | `201 Created`<br>`400 Bad Request`<br>`409 Conflict`<br>`415 Unsupported Media Type`<br>`422 Unprocessable Entity` | Respons dengan header `Location: /api/v1/students/1`:<br>`{"success": true, "message": "mahasiswa berhasil ditambahkan", "data": {"id": 1, "nim": "18221001", ...}}` |
| **PUT** | `/api/v1/students/:id` | Path: `:id`<br>Header: `Content-Type: application/json` | `{"nim": "18221001", "name": "Budi Santoso", "grade": 90.0, "is_active": false}` | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`422 Unprocessable Entity` | `{"success": true, "message": "data mahasiswa berhasil diganti seluruhnya", "data": {...}}` |
| **PATCH** | `/api/v1/students/:id` | Path: `:id`<br>Header: `Content-Type: application/json` | `{"grade": 95.0}` *(hanya field yang ingin diubah)* | `200 OK`<br>`400 Bad Request`<br>`404 Not Found`<br>`409 Conflict`<br>`422 Unprocessable Entity` | `{"success": true, "message": "data mahasiswa berhasil diperbarui sebagian", "data": {...}}` |
| **DELETE** | `/api/v1/students/:id` | Path: `:id` | - | `204 No Content`<br>`400 Bad Request`<br>`404 Not Found` | *(Tanpa Body)* |

## Cara Menjalankan Aplikasi

1. Unduh modul dependensi:
   ```bash
   go mod init api-students
   go get github.com/gofiber/fiber/v2
   go mod tidy
   ```
2. Jalankan server:
   ```bash
   go run .
   ```
3. Akses server di `http://localhost:3000`.
