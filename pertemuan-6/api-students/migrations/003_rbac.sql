-- Migration 003: Role-Based Access Control (RBAC)
CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh terhadap seluruh data dan konfigurasi sistem'),
    ('staff', 'Melihat seluruh data mahasiswa dan membuat data baru'),
    ('user',  'Hanya mengelola profil dan data miliknya sendiri')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
    ('user:list', 'Melihat daftar seluruh pengguna'),
    ('user:read:any', 'Melihat data pengguna lain'),
    ('user:update:any', 'Mengubah data pengguna lain'),
    ('user:delete', 'Menghapus pengguna'),
    ('role:assign', 'Mengubah role pengguna lain'),
    ('student:list', 'Melihat daftar mahasiswa'),
    ('student:read:any', 'Melihat detail mahasiswa siapapun'),
    ('student:create', 'Menambahkan mahasiswa baru'),
    ('student:update:any', 'Mengubah data mahasiswa siapapun'),
    ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'user:list'),
    ('admin', 'user:read:any'),
    ('admin', 'user:update:any'),
    ('admin', 'user:delete'),
    ('admin', 'role:assign'),
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'user:list'),
    ('staff', 'user:read:any'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- Tambahkan foreign key constraint pada tabel users
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_fkey;
ALTER TABLE users
    ADD CONSTRAINT users_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

-- Tambahkan kolom owner_id pada tabel students
ALTER TABLE students ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);
