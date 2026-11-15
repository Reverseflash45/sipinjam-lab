-- 004: data awal untuk mencoba API. Password semua akun: Rahasia123
INSERT INTO users (name, email, identity_no, password_hash, role) VALUES
    ('Admin Lab',   'admin@vokasi.unair.ac.id',   '198001012005011001', '$2a$10$ckvGgxDyD22cJ0JFaxae6.KR.ZxLPuEq.ePFvBgfGGu/m4dONCS.S', 'admin'),
    ('Laboran Satu','laboran@vokasi.unair.ac.id', '198502022010012002', '$2a$10$ckvGgxDyD22cJ0JFaxae6.KR.ZxLPuEq.ePFvBgfGGu/m4dONCS.S', 'laboran')
ON CONFLICT (email) DO NOTHING;

INSERT INTO rooms (code, name, building, capacity) VALUES
    ('LAB-RPL', 'Laboratorium Rekayasa Perangkat Lunak', 'Gedung Vokasi B', 40),
    ('LAB-JAR', 'Laboratorium Jaringan Komputer',        'Gedung Vokasi B', 30),
    ('LAB-MM',  'Laboratorium Multimedia',               'Gedung Vokasi C', 25)
ON CONFLICT (code) DO NOTHING;

INSERT INTO items (code, name, category, stock) VALUES
    ('PRJ-01', 'Proyektor Epson EB-X51',  'presentasi', 4),
    ('LPT-01', 'Laptop Lenovo ThinkPad',  'komputer',   10),
    ('RTR-01', 'Router MikroTik hEX',     'jaringan',   8),
    ('CAM-01', 'Kamera Sony A6400',       'multimedia', 2)
ON CONFLICT (code) DO NOTHING;
