-- Migrasi 001: Membuat tabel students dan indeks penunjang kinerja
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks UNIK untuk menjamin keunikan NIM tanpa membedakan huruf besar dan kecil.
-- Batasan UNIQUE di sisi basis data mencegah race condition bila ada dua permintaan bersamaan.
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_lower_key
    ON students (LOWER(nim));

-- Indeks penunjang pencarian nama mahasiswa (ILIKE / LOWER)
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
