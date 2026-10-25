-- Migration 005: Kelas, KRS, dan Pembimbing Tugas Akhir (Supervisors)

-- 1. Tabel Courses & Enrollments
CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    credits SMALLINT NOT NULL CHECK (credits BETWEEN 1 AND 6),
    quota INTEGER NOT NULL CHECK (quota > 0),
    enrolled_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT courses_enrolled_within_quota
        CHECK (enrolled_count >= 0 AND enrolled_count <= quota)
);

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_course_user_unique UNIQUE (course_id, user_id)
);

CREATE INDEX IF NOT EXISTS enrollments_user_idx ON enrollments (user_id);

-- 2. Tabel Supervisors & Supervisions (Tugas Mandiri D.2)
CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    supervised_count INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT supervisors_count_within_quota
        CHECK (supervised_count >= 0 AND supervised_count <= quota)
);

CREATE TABLE IF NOT EXISTS supervisions (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    supervisor_id INTEGER NOT NULL REFERENCES supervisors(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS supervisions_supervisor_idx ON supervisions (supervisor_id);

-- 3. Seed Data Kelas
INSERT INTO courses (code, name, credits, quota) VALUES
    ('IF101', 'Pemrograman Backend Lanjut', 3, 40),
    ('IF102', 'Basis Data Lanjut', 3, 40),
    ('IF301', 'Kecerdasan Buatan', 3, 40),
    ('IF302', 'Keamanan Informasi', 3, 40),
    ('IF303', 'Komputasi Awan', 3, 40),
    ('IF304', 'Data Mining', 3, 40),
    ('IF305', 'Pemrograman Mobile', 3, 40),
    ('IF306', 'Interaksi Manusia Komputer', 3, 40),
    ('IF307', 'Sistem Terdistribusi', 3, 40),
    ('IF308', 'Visi Komputer', 3, 40),
    ('IF309', 'Pemrosesan Bahasa Alami', 3, 40),
    ('IF310', 'Manajemen Proyek TI', 3, 40),
    ('IF399', 'Kelas Rebutan', 3, 5)
ON CONFLICT (code) DO NOTHING;

-- 4. Seed Data Dosen Pembimbing
INSERT INTO supervisors (id, name, quota, supervised_count) VALUES
    (1, 'Dr. Ir. Budi Rahardjo, M.T.', 3, 0),
    (2, 'Prof. Siti Maryam, Ph.D.', 5, 0),
    (3, 'Agus Prasetyo, S.Kom., M.Kom.', 3, 0)
ON CONFLICT (id) DO NOTHING;

-- 5. Permissions RBAC
INSERT INTO permissions (name, description) VALUES
    ('course:manage', 'Mengubah kuota dan data kelas'),
    ('supervision:manage', 'Menetapkan dan mengelola pembimbing tugas akhir')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'course:manage'),
    ('admin', 'supervision:manage')
ON CONFLICT DO NOTHING;
