-- Skema Tabel Mahasiswa untuk PostgreSQL
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks unik untuk keunikan NIM (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks performa untuk pencarian nama mahasiswa
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
