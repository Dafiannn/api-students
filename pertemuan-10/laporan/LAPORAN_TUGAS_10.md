# LAPORAN PRAKTIKUM PEMROGRAMAN BACKEND LANJUT (PBL)
## MODUL 10: POLYGLOT PERSISTENCE — REST API DENGAN MONGODB

**Disusun Oleh:**
- **Nama:** [Nama Mahasiswa]
- **NIM:** [NIM Mahasiswa]
- **Program Studi:** D4 Teknik Informatika
- **Fakultas:** Fakultas Vokasi, Universitas Airlangga
- **Dosen Pengampu:** Tim Dosen Praktikum Pemrograman Backend Lanjut
- **Tautan Repositori GitHub:** [https://github.com/username/praktikum-backend-lanjut](https://github.com/username/praktikum-backend-lanjut)

---

## DAFTAR ISI
1. [Ringkasan Eksekutif & Konsep Polyglot Persistence](#1-ringkasan-eksekutif--konsep-polyglot-persistence)
2. [Arsitektur Sistem (PostgreSQL + MongoDB Atlas)](#2-arsitektur-sistem-postgresql--mongodb-atlas)
3. [C.1 Rancangan Dokumen Portofolio Mahasiswa](#c1-rancangan-dokumen-portofolio-mahasiswa)
4. [C.2 Implementasi Endpoint & Potongan Kode Penting](#c2-implementasi-endpoint--potongan-kode-penting)
5. [C.3 Seluruh Bukti Pengujian (Verification & Test Proofs)](#c3-seluruh-bukti-pengujian-verification--test-proofs)
6. [C.4 Jawaban Analisis Kritis](#c4-jawaban-analisis-kritis)
7. [Kesimpulan](#7-kesimpulan)

---

## 1. Ringkasan Eksekutif & Konsep Polyglot Persistence

Dalam pengembangan sistem backend modern berskala besar (*enterprise architecture*), tidak ada satu jenis database pun yang merupakan solusi terbaik untuk seluruh bentuk data (*"No One-Size-Fits-All Database"*). Pola perancangan yang memadukan lebih dari satu jenis basis data dalam satu ekosistem aplikasi disebut **Polyglot Persistence**.

Pada proyek ini:
- **Relational Database (PostgreSQL - Supabase Cloud)**: Digunakan untuk data transaksional terstruktur yang membutuhkan jaminan integritas data tingkat tinggi, *ACID transaction*, relasi foreign key, validasi batas SKS, dan pencegahan *over-allocation* kuota (tabel `users`, `students`, `courses`, `enrollments`, `supervisors`, `permissions`).
- **Document NoSQL Database (MongoDB Atlas Cloud)**: Digunakan untuk data semi-terstruktur yang bersifat dokumen utuh, dinamis, bersarang (*embedded*), serta memiliki variasi array yang fleksibel (collection `course_reviews` dan `student_portfolios`).

### Keuntungan Arsitektur Polyglot Persistence:
1. **Penyimpanan Alami Data Bersarang**: Dokumen ulasan kelas dan portofolio mahasiswa beserta array proyek dan sertifikat tersimpan sebagai satu dokumen JSON/BSON utuh tanpa perlu melakukan operasi JOIN yang mahal di 4 tabel relasional.
2. **Kemandirian Skema (Schema Flexibility)**: Struktur proyek portofolio dapat memiliki daftar tag teknologi yang bervariasi tanpa memerlukan migrasi skema tabel DDL.
3. **Ketahanan Parsial (*Fault Isolation*)**: Matinya salah satu engine database tidak meruntuhkan seluruh fungsionalitas sistem. Jika MongoDB mengalami gangguan, fitur inti pendaftaran KRS dan login di PostgreSQL tetap beroperasi normal.

---

## 2. Arsitektur Sistem (PostgreSQL + MongoDB Atlas)

### A. Struktur Folder Proyek Modul 10
Sesuai prinsip Clean Architecture, layer model tidak mengenal MongoDB driver v2 (`go.mongodb.org/mongo-driver/v2`) maupun driver PostgreSQL (`pgx/v5`). Seluruh tipe BSON (seperti `bson.ObjectID`) diisolasi di layer Repository.

```
pertemuan-10/
├── api-students/
│   ├── app/
│   │   ├── model/
│   │   │   ├── auth.go
│   │   │   ├── course.go
│   │   │   ├── portfolio.go            <- Entitas Murni Portofolio & Proyek (Tanpa BSON)
│   │   │   ├── review.go               <- Entitas Murni Ulasan Kelas (Tanpa BSON)
│   │   │   ├── supervisor.go
│   │   │   └── user.go
│   │   ├── repository/
│   │   │   ├── course_repository.go    <- PostgreSQL (Menambahkan IsEnrolled)
│   │   │   ├── dbtx.go
│   │   │   ├── portfolio_repository.go <- MongoDB (student_portfolios, BSON mapping)
│   │   │   ├── review_repository.go    <- MongoDB (course_reviews, $facet, atomik cap 50)
│   │   │   ├── role_repository.go
│   │   │   ├── student_repository.go
│   │   │   ├── supervisor_repository.go
│   │   │   ├── token_repository.go
│   │   │   ├── tx.go
│   │   │   └── user_repository.go
│   │   └── service/
│   │       ├── auth_service.go
│   │       ├── course_rules.go
│   │       ├── course_service.go
│   │       ├── portfolio_service.go    <- Koordinasi Dual DB (Cek Mahasiswa di PG)
│   │       ├── review_rules.go         <- Normalisasi Tag (Pure Function)
│   │       ├── review_rules_test.go    <- Unit Test Normalisasi (100% Pass)
│   │       ├── review_service.go       <- Koordinasi Dual DB (Cek KRS di PG, Ulasan di Mongo)
│   │       ├── student_rules.go
│   │       ├── student_service.go
│   │       └── supervisor_service.go
│   ├── config/
│   │   ├── app.go
│   │   ├── env.go
│   │   └── logger.go
│   ├── database/
│   │   ├── mongo.go                    <- Driver MongoDB v2 Client & Ping Validator
│   │   └── postgres.go                 <- Driver pgxpool PostgreSQL
│   ├── helper/
│   │   ├── errors.go
│   │   ├── validator.go                <- Validasi slice/array item (min/max item)
│   │   └── ...
│   ├── migrations/
│   │   ├── 005_courses_and_supervisors.sql
│   │   ├── 006_seed_loadtest.sql
│   │   └── 007_review_permissions.sql  <- Permission moderasi ulasan di PG
│   ├── route/
│   │   └── route.go                    <- Dual Database Health Check & Endpoints
│   ├── .env
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   └── main.go                         <- Inisialisasi Dual Database & Index Auto-Creation
└── laporan/
    └── LAPORAN_TUGAS_10.md
```

### B. Dual Database Health Check
Endpoint `GET /api/v1/health` dirancang untuk memverifikasi kesiapan kedua database secara simultan:
- Memeriksa koneksi `pool.Ping(ctx)` ke PostgreSQL (Supabase).
- Memeriksa koneksi `mongoClient.Ping(ctx, nil)` ke MongoDB Atlas.
- Bila salah satu database mati, server mengembalikan status `503 Service Unavailable` secara spesifik mengidentifikasi database yang mengalami gangguan, tanpa menyebabkan crash pada aplikasi.

---

## 3. C.1 Rancangan Dokumen Portofolio Mahasiswa

Fitur Tugas Mandiri Modul 10 membangun koleksi `student_portfolios` untuk menyimpan data portofolio karir, riwayat proyek rekayasa perangkat lunak, dan sertifikasi keahlian mahasiswa.

### A. Contoh Dokumen Lengkap di MongoDB (`student_portfolios`)
Berikut adalah struktur dokumen BSON/JSON yang tersimpan di MongoDB Atlas:

```json
{
  "_id": {
    "$oid": "6ae0e8bea6ab379023bc4334"
  },
  "student_id": 1,
  "headline": "Backend Engineer & Distributed Systems Enthusiast",
  "bio": "Mahasiswa D4 Teknik Informatika UNAIR dengan minat mendalam pada Go, PostgreSQL, dan MongoDB.",
  "projects": [
    {
      "_id": {
        "$oid": "6ae0e89f29de89b81103d9a4"
      },
      "title": "Sistem Informasi Akademik",
      "tech": [
        "Go",
        "PostgreSQL",
        "Docker"
      ],
      "url": "https://github.com/mhs/siakad",
      "year": 2025
    },
    {
      "_id": {
        "$oid": "6ae0e89f29de89b81103d9a5"
      },
      "title": "E-Commerce Microservices",
      "tech": [
        "Go",
        "MongoDB",
        "Redis",
        "Kafka"
      ],
      "url": "https://github.com/mhs/ecommerce",
      "year": 2026
    }
  ],
  "certificates": [
    {
      "_id": {
        "$oid": "6ae0e89f29de89b81103d9a6"
      },
      "name": "Certified Go Developer",
      "issuer": "Linux Foundation",
      "date": {
        "$date": "2025-06-15T00:00:00.000Z"
      }
    }
  ],
  "created_at": {
    "$date": "2026-10-27T18:08:31.541Z"
  },
  "updated_at": {
    "$date": "2026-10-27T18:08:31.541Z"
  }
}
```

### B. Penjelasan Desain Dokumen & Indeks
1. **Relasi Antar-Database (`student_id`)**:
   - `student_id` bertipe integer yang merujuk pada `students.id` di PostgreSQL.
   - Diberikan indeks unik di MongoDB (`student_portfolio_unique`):
     ```go
     Keys: bson.D{{Key: "student_id", Value: 1}}, Options: options.Index().SetUnique(true)
     ```
   - Indeks unik ini secara mutlak menjamin aturan bisnis: **Satu mahasiswa hanya boleh memiliki satu dokumen portofolio**.
2. **Pola Embedding (Penyarangan) Array `projects` dan `certificates`**:
   - Seluruh entitas proyek dan sertifikasi disimpan di dalam dokumen portofolio induk.
   - Karena portofolio dibaca sebagai satu profil utuh pada halaman profil mahasiswa, embedding mengeliminasi kebutuhan query terpisah atau `$lookup` yang lambat.
   - Setiap sub-dokumen proyek memiliki `_id` bertipe `ObjectID` unik sehingga dapat dimutasi atau dihapus secara spesifik.
3. **Indeks Multikey (`portfolio_project_tech`)**:
   - Dibuat indeks multikey pada field array `projects.tech`:
     ```go
     Keys: bson.D{{Key: "projects.tech", Value: 1}}
     ```
   - Indeks ini mengindeks setiap elemen string teknologi dalam array proyek, memungkinkan eksekusi cepat pada aggregation pipeline pencarian statistik teknologi terpopuler.

---

## 4. C.2 Implementasi Endpoint & Potongan Kode Penting

### A. Endpoint yang Diimplementasikan
| HTTP Method | Path Endpoint | Akses & Middleware | Deskripsi Fungsi |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/health` | Publik | Memeriksa ketersediaan koneksi PostgreSQL dan MongoDB Atlas secara simultan. |
| `GET` | `/api/v1/courses/:id/reviews` | Auth JWT | Mengambil daftar ulasan kelas dengan paginasi kursor (`before`) dan filter tag. |
| `GET` | `/api/v1/courses/:id/reviews/summary` | Auth JWT | Menghasilkan ringkasan rating ulasan, sebaran bintang 1-5, dan 5 tag teratas via `$facet`. |
| `POST` | `/api/v1/courses/:id/reviews` | Auth JWT | Membuat ulasan baru (hanya jika mahasiswa terdaftar di kelas pada PostgreSQL). |
| `GET` | `/api/v1/reviews/:id` | Auth JWT | Mengambil satu ulasan lengkap beserta seluruh balasannya. |
| `PATCH` | `/api/v1/reviews/:id` | Auth JWT | Mengubah ulasan milik sendiri secara atomik (pemeriksaan `user_id` di filter). |
| `DELETE` | `/api/v1/reviews/:id` | Auth JWT | Menghapus ulasan (dapat dihapus pemiliknya atau admin dengan `review:moderate`). |
| `POST` | `/api/v1/reviews/:id/replies` | Auth JWT | Menambah balasan ulasan dengan batas atomik maksimal 50 balasan. |
| `POST` | `/api/v1/reviews/:id/helpful` | Auth JWT | Menandai ulasan bermanfaat (atomik `$addToSet` dan `$inc`, maksimal 1x per user). |
| `GET` | `/api/v1/students/:id/portfolio` | Auth JWT | Membaca portofolio mahasiswa (memverifikasi eksistensi mahasiswa di PostgreSQL). |
| `PUT` | `/api/v1/students/:id/portfolio` | Auth JWT | Membuat atau mengganti portofolio lengkap (*upsert*). |
| `POST` | `/api/v1/students/:id/portfolio/projects` | Auth JWT | Menambah proyek baru dengan pembatasan kuota atomik maksimal 20 proyek. |
| `DELETE` | `/api/v1/students/:id/portfolio/projects/:projectId` | Auth JWT | Menghapus proyek dari array menggunakan operator `$pull`. |
| `GET` | `/api/v1/portfolios/stats/technologies` | Auth JWT | Aggregation pipeline menampilkan 10 teknologi paling populer di seluruh portofolio. |

---

### B. Potongan Kode Penting

#### 1. Pembatasan Atomik 20 Proyek (Pola `"projects.19": {"$exists": false}`)
Untuk mencegah *race condition* penambahan proyek melebihi batas 20 tanpa membaca panjang array terlebih dahulu di Go, pengecekan batas diletakkan langsung di dalam predikat filter MongoDB:

```go
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.CreateProjectRequest) (model.Portfolio, error) {
    projDoc := projectDocument{
        ID:    bson.NewObjectID(),
        Title: p.Title,
        Tech:  p.Tech,
        URL:   p.URL,
        Year:  p.Year,
    }

    // projects.19 adalah elemen ke-20 (indeks 0..19).
    // Syarat ini hanya bernilai true jika portofolio belum memiliki 20 proyek.
    limitField := fmt.Sprintf("projects.%d", MaxProjectsPerPortfolio-1)
    filter := bson.M{
        "student_id": studentID,
        limitField:   bson.M{"$exists": false},
    }

    update := bson.M{
        "$push": bson.M{"projects": projDoc},
        "$set":  bson.M{"updated_at": now()},
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    var doc portfolioDocument
    err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
    if err == nil {
        return doc.toModel(), nil
    }

    if !errors.Is(err, mongo.ErrNoDocuments) {
        return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
    }

    // Membedakan apakah portofolio tidak ada atau kuota sudah penuh
    count, _ := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
    if count == 0 {
        return model.Portfolio{}, ErrNotFound
    }
    return model.Portfolio{}, ErrProjectLimit
}
```

#### 2. Penghapusan Proyek dengan Operator `$pull`
Menghapus sub-dokumen proyek bersarang secara langsung tanpa perlu mengganti seluruh array:

```go
func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error) {
    oid, err := parseID(projectID)
    if err != nil {
        return model.Portfolio{}, err
    }

    filter := bson.M{
        "student_id":   studentID,
        "projects._id": oid,
    }
    update := bson.M{
        "$pull": bson.M{"projects": bson.M{"_id": oid}},
        "$set":  bson.M{"updated_at": now()},
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    var doc portfolioDocument
    err = r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
    if err == nil {
        return doc.toModel(), nil
    }
    return model.Portfolio{}, ErrProjectNotFound
}
```

#### 3. Aggregation Pipeline Dua Kali `$unwind` untuk Statistik Teknologi
Untuk mengelompokkan dan menghitung frekuensi setiap string teknologi dari seluruh portofolio:

```go
func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context) ([]model.TechStat, error) {
    pipeline := mongo.Pipeline{
        // Unwind pertama: memecah array projects (1 dokumen -> N dokumen proyek)
        {{Key: "$unwind", Value: "$projects"}},
        // Unwind kedua: memecah array string tech (1 proyek -> M dokumen teknologi)
        {{Key: "$unwind", Value: "$projects.tech"}},
        // Grouping berdasarkan nama teknologi dan hitung frekuensinya
        {{Key: "$group", Value: bson.M{
            "_id":   "$projects.tech",
            "total": bson.M{"$sum": 1},
        }}},
        // Urutkan berdasarkan total terbanyak secara descending
        {{Key: "$sort", Value: bson.D{
            {Key: "total", Value: -1},
            {Key: "_id", Value: 1},
        }}},
        {{Key: "$limit", Value: 10}},
    }

    cursor, err := r.coll.Aggregate(ctx, pipeline)
    // ... decode hasil ...
}
```

---

## 5. C.3 Seluruh Bukti Pengujian (Verification & Test Proofs)

Pengujian end-to-end dilakukan terhadap server backend `api-students` yang terhubung secara nyata ke **PostgreSQL Supabase Cloud** dan **MongoDB Atlas Cloud**.

### A. Bukti Eksekusi Uji Otomatis
```text
=================================================================
       PENGUJIAN OTOMATIS FITUR MODUL 10 (MONGODB ATLAS)         
=================================================================

--- 1. Persiapan Pendaftaran KRS di PostgreSQL ---
Enroll mhs001 ke Course 1 -> Status: 409 (Sudah terdaftar)
Enroll mhs002 ke Course 1 -> Status: 201 (Berhasil terdaftar)

--- 2. Uji Mahasiswa Tidak Terdaftar Menulis Ulasan ---
mhs003 menulis ulasan -> Status: 403 | Body: {"success":false,"code":"NOT_ENROLLED","message":"hanya peserta kelas yang boleh menulis ulasan"}

--- 3. Uji Mahasiswa Menulis Ulasan Pertama Kali + Tag Normalization ---
mhs001 menulis ulasan -> Status: 201 | Body: {"success":true,"message":"ulasan berhasil disimpan","data":{"id":"6ae0e89c29de89b81103d9a1","course_id":1,"user_id":1,"username":"mhs001","rating":5,"comment":"Materi Go-nya jelas, tugasnya cukup banyak tapi seru","tags":["tugas-banyak","dosen-jelas"],"replies":[],"helpful_count":0}}

--- 4. Uji Ulasan Ganda oleh Mahasiswa yang Sama ---
mhs001 menulis ulasan kedua -> Status: 409 | Body: {"success":false,"code":"ALREADY_REVIEWED","message":"Anda sudah menulis ulasan untuk kelas ini, ubah ulasan yang ada"}

--- 5. Uji Validasi Deklaratif ---
Rating 9, komentar pendek -> Status: 422 | Body: {"success":false,"code":"VALIDATION_ERROR","message":"validasi gagal","fields":{"Comment":"minimal 10 karakter","Rating":"nilai maksimal 5","Tags":"maksimal 5 item"}}

--- 6. Uji Ketahanan NoSQL Injection ---
Kirim operator NoSQL di body -> Status: 400 | Body: {"success":false,"code":"BAD_REQUEST","message":"body harus berupa JSON yang valid"}

--- 7. Uji Menilai Ulasan Sendiri ---
mhs001 menilai ulasannya sendiri -> Status: 409 | Body: {"success":false,"code":"OWN_REVIEW","message":"tidak dapat menilai ulasan sendiri"}

--- 8. Uji Menilai Ulasan Orang Lain & Duplikat Vote ---
mhs002 memberi helpful -> Status: 200 | Body: {"success":true,"message":"terima kasih atas penilaiannya","data":{"helpful_count":1}}
mhs002 memberi helpful kedua kali -> Status: 409 | Body: {"success":false,"code":"ALREADY_VOTED","message":"Anda sudah menilai ulasan ini"}

--- 9. Uji Menambah Balasan Ulasan ---
mhs002 membalas ulasan -> Status: 201 | Body: {"success":true,"message":"balasan berhasil ditambahkan"}

--- 10. Uji Aggregation Summary Course Reviews ---
Course 1 Summary -> Status: 200 | Body: {"success":true,"message":"ringkasan ulasan berhasil diambil","data":{"course_id":1,"count":1,"average_rating":5,"distribution":{"1":0,"2":0,"3":0,"4":0,"5":1},"top_tags":[{"tag":"dosen-jelas","total":1},{"tag":"tugas-banyak","total":1}]}}
```

---

### B. Bukti Pengujian Tugas Mandiri: Portofolio Mahasiswa
```text
=================================================================
             UJI TUGAS MANDIRI: PORTOFOLIO MAHASISWA             
=================================================================

--- A. Simpan/Update Portofolio Mahasiswa (PUT /students/1/portfolio) ---
PUT /students/1/portfolio -> Status: 200 | Body: {"success":true,"message":"portofolio berhasil disimpan","data":{"student_id":1,"headline":"Backend Engineer & Distributed Systems Enthusiast","projects":[{"title":"Sistem Informasi Akademik"},{"title":"E-Commerce Microservices"}],"certificates":[{"name":"Certified Go Developer"}]}}

--- B. Ambil Portofolio Mahasiswa (GET /students/1/portfolio) ---
GET /students/1/portfolio -> Status: 200 | Body: {"success":true,"message":"portofolio berhasil diambil","data":{"student_id":1,"headline":"Backend Engineer & Distributed Systems Enthusiast"}}

--- C. Ambil Portofolio Mahasiswa Fiktif (ID 99999) ---
GET /students/99999/portfolio -> Status: 404 | Body: {"success":false,"code":"NOT_FOUND","message":"mahasiswa tidak ditemukan di PostgreSQL"}

--- D. Mengisi hingga 18 Proyek ---
Berhasil mengisi hingga 18 proyek.

--- E. Uji Konkurensi: 10 Request Serentak Menambah Proyek (Batas Kuota 20) ---
Hasil Konkurensi Proyek -> Berhasil: 2 (Status 201) | Ditolak Batas: 8 (Status 409)
[PASSED] Integritas Terjamin: Tepat 2 proyek berhasil, tepat mencapai batas 20 proyek!

--- F. Uji Aggregation Pipeline 10 Teknologi Terpopuler (GET /portfolios/stats/technologies) ---
Top 10 Tech -> Status: 200 | Body: {"success":true,"message":"statistik 10 teknologi paling banyak digunakan","data":[{"technology":"Go","total":20},{"technology":"PostgreSQL","total":17},{"technology":"Python","total":2},{"technology":"Docker","total":1},{"technology":"Kafka","total":1},{"technology":"MongoDB","total":1},{"technology":"Redis","total":1}]}

--- G. Uji Admin Menghapus Ulasan (Moderasi) ---
Admin DELETE ulasan -> Status: 204 (No Content)
```

---

### C. Analisis Hasil Explain (`executionStats`)
Saat query dijalankan di MongoDB dengan indeks `student_portfolio_unique`:
- **Stage**: `IXSCAN` $\to$ `FETCH`.
- **docsExamined**: 1
- **nReturned**: 1
- **Rasio Pemeriksaan**: 1:1 (*Perfect Efficiency*). MongoDB langsung menemukan dokumen portofolio mahasiswa melalui indeks B-Tree tanpa melakukan pemindaian seluruh koleksi (*Zero COLLSCAN*).

---

## 6. C.4 Jawaban Analisis Kritis

### Pertanyaan 10: Mengapa `projects` disarangkan (*embedded*) dan bukan disimpan di collection tersendiri? Sebutkan satu perubahan kebutuhan yang akan membuat keputusan itu harus dibalik!

**Jawaban:**
Keputusan menyarangkan (*embedding*) array `projects` di dalam dokumen `student_portfolios` diambil berdasarkan prinsip pola akses baca (*read access pattern*): portofolio mahasiswa selalu ditampilkan sebagai satu halaman profil yang utuh, di mana daftar proyek tidak pernah dibaca secara mandiri tanpa profil mahasiswanya. Dengan *embedding*, database MongoDB hanya perlu membaca satu dokumen tunggal dalam satu operasi I/O, tanpa memerlukan query tambahan atau operasi `$lookup` yang berat. Selain itu, jumlah proyek yang dimiliki seorang mahasiswa dibatasi secara ketat maksimal 20 item, sehingga ukuran dokumen tidak akan pernah mendekati batas fisik MongoDB (16 MB).

**Satu perubahan kebutuhan yang mewajibkan keputusan tersebut dibalik (menjadi *Referencing*):**
Keputusan ini harus dibalik jika sistem kampus mengembangkan **Platform Marketplace Proyek Mahasiswa Terbuka**, di mana proyek mahasiswa dapat diakses, dicari, dan difilter secara independen oleh pihak industri, memiliki kolom komentar dan ulasan dari publik yang tumbuh tanpa batas, atau satu proyek dapat dikerjakan secara kolaboratif oleh multi-mahasiswa (*many-to-many relationship*). Pada kebutuhan tersebut, entitas `projects` harus dipisahkan menjadi collection tersendiri, dan dokumen portofolio hanya menyimpan array ID rujukan (*referencing*).

---

### Pertanyaan 11: Bila seorang mahasiswa dihapus dari PostgreSQL, apa yang terjadi pada portofolionya di MongoDB? Usulkan dua cara menanganinya beserta kelemahan masing-masing!

**Jawaban:**
Karena PostgreSQL dan MongoDB adalah dua sistem database yang terpisah secara fisik (*polyglot architecture*), MongoDB tidak memiliki mekanisme relasi *Foreign Key* otomatis (`ON DELETE CASCADE`) yang terhubung ke PostgreSQL. Akibatnya, jika record mahasiswa dengan `id = 1` dihapus dari PostgreSQL, dokumen portofolio mahasiswa tersebut di MongoDB akan menjadi **Data Yatim (*Orphaned Document*)** — dokumen tetap tersimpan dan memakai kapasitas storage, namun tidak lagi memiliki entitas pemilik yang sah di database utama.

#### Dua Cara Penanganan Beserta Kelemahannya:
1. **Pendekatan Synchronous Dual-Delete pada Application Service**:
   - *Mekanisme*: Pada handler `DeleteStudent` di service layer, setelah berhasil mengeksekusi `DELETE FROM students WHERE id = $1` di PostgreSQL, aplikasi langsung memanggil `portfolioRepo.DeleteByStudentID(ctx, id)` ke MongoDB.
   - *Kelemahan*: Menimbulkan masalah *Dual-Write Failure*. Jika server PostgreSQL berhasil menghapus data namun server MongoDB mati atau mengalami timeout jaringan tepat sebelum instruksi hapus dijalankan, transaksi di PostgreSQL sudah terlanjur di-commit sehingga data yatim di MongoDB tetap tertinggal tanpa ada mekanisme *rollback* lintas database.
2. **Pendekatan Asynchronous Event-Driven Cleanup (Transactional Outbox Pattern)**:
   - *Mekanisme*: Penghapusan mahasiswa mencatat event `student.deleted` ke tabel `outbox` di PostgreSQL dalam satu transaksi ACID yang sama. Sebuah background worker/relay secara asinkron membaca event tersebut dan mengeksekusi penghapusan portofolio di MongoDB dengan jaminan *at-least-once delivery*.
   - *Kelemahan*: Membutuhkan infrastruktur tambahan (Message Broker/Queue seperti Redis/Asynq atau Kafka), serta menimbulkan kondisi *Eventual Consistency* sementara di mana dokumen portofolio masih dapat diakses selama beberapa detik sebelum background worker selesai memproses event penghapusan.

---

### Pertanyaan 12: Endpoint mana di aplikasi Anda yang tetap berjalan bila MongoDB mati, dan mana yang tidak? Apakah perilaku itu yang Anda inginkan?

**Jawaban:**
Berdasarkan arsitektur *polyglot persistence* yang diimplementasikan:
1. **Endpoint yang TETAP Berjalan Normal saat MongoDB Mati**:
   - Seluruh endpoint Core Academic & Authentication yang 100% bergantung pada PostgreSQL:
     - `POST /api/v1/auth/login`, `POST /api/v1/auth/register`, `POST /api/v1/auth/refresh`
     - `GET /api/v1/students`, `POST /api/v1/students`, `PUT /api/v1/students/:id`
     - `GET /api/v1/courses`, `POST /api/v1/courses/:id/enroll`, `POST /api/v1/courses/swap`
     - `POST /api/v1/supervisors/:id/assign` (Pendaftaran Pembimbing TA)
2. **Endpoint yang GAGAL (Mengembalikan HTTP 500 / 503)**:
   - `GET /api/v1/health` (melaporkan `503 Service Unavailable: MongoDB tidak dapat dihubungi`).
   - Seluruh endpoint ulasan kelas: `GET/POST /api/v1/courses/:id/reviews`, `GET /api/v1/reviews/:id`, dsb.
   - Seluruh endpoint portofolio mahasiswa: `GET/PUT /api/v1/students/:id/portfolio`, dsb.

**Evaluasi Arsitektural**:
**Ya, perilaku ini adalah perilaku ideal yang memang dirancang secara sengaja (*Graceful Degradation*)**. Fitur ulasan kelas dan portofolio karir adalah fitur pelengkap (*secondary features*), sedangkan pendaftaran KRS dan penetapan dosen pembimbing adalah fungsi misi-kritis kampus (*mission-critical core business*). Dengan memisahkan database, kegagalan pada cluster MongoDB Atlas tidak melumpuhkan sistem operasional perkuliahan utama. Mahasiswa tetap dapat melakukan perwalian dan War KRS tanpa terganggu oleh matinya layanan ulasan atau portofolio.

---

## 7. Kesimpulan

Implementasi Modul 10 berhasil membuktikan keunggulan arsitektur **Polyglot Persistence** pada ekosistem Go Fiber:
1. Relasi transaksional kritis tetap dilindungi oleh integritas ACID PostgreSQL.
2. Data ulasan kelas dan portofolio karir tersimpan secara fleksibel dan efisien di MongoDB Atlas dalam format dokumen JSON/BSON bersarang (*embedded*).
3. Batasan bisnis (seperti kuota maksimal 20 proyek dan 50 balasan) terbukti kebal dari *race condition* melalui pemanfaatan operator atomik bersyarat MongoDB (`$push` dengan predikat filter `$exists: false`).
4. Fitur pencarian analitik statistik teknologi berhasil diakselerasi menggunakan *Aggregation Pipeline* dengan teknik dekonstruksi ganda (*two-stage `$unwind`*).
