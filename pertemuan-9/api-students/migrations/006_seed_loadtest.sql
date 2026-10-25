-- Migration 006: Akun mahasiswa untuk uji beban serentak
-- Password bcrypt default: "password123" ($2a$12$c5v7D1qjlQnyHfuWLoREh.1WjWjhIGiibgQ3O6bGyBsfbXj66fvBW)
INSERT INTO users (username, email, password, role)
SELECT 'mhs' || LPAD(g::TEXT, 3, '0'),
       'mhs' || LPAD(g::TEXT, 3, '0') || '@student.unair.ac.id',
       '$2a$12$c5v7D1qjlQnyHfuWLoREh.1WjWjhIGiibgQ3O6bGyBsfbXj66fvBW', 'user'
FROM generate_series(1, 60) AS g
ON CONFLICT DO NOTHING;
