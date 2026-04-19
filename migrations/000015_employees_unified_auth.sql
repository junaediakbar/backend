-- Satu baris employees = profil operasional + akun login. Tabel users dihapus.

ALTER TABLE laundry_backend.employees
  ADD COLUMN IF NOT EXISTS email TEXT,
  ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS role laundry_backend."Role";

-- Isi dari users yang terikat ke karyawan
UPDATE laundry_backend.employees e
SET
  email = u.email,
  password_hash = u.password_hash,
  role = u.role
FROM laundry_backend.users u
WHERE u.employee_id IS NOT NULL
  AND u.employee_id = e.id;

-- Akun tanpa baris karyawan -> jadi baris employees (id sama dengan users.id)
INSERT INTO laundry_backend.employees (id, name, is_active, email, password_hash, role, created_at, updated_at)
SELECT u.id, u.name, u.is_active, u.email, u.password_hash, u.role, u.created_at, u.updated_at
FROM laundry_backend.users u
WHERE u.employee_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM laundry_backend.employees e2 WHERE e2.id = u.id);

-- Karyawan tanpa akun (belum punya email): placeholder wajib ganti
UPDATE laundry_backend.employees e
SET
  email = 'pending+' || e.id || '@migrated.local',
  password_hash = CASE
    WHEN trim(COALESCE(e.password_hash, '')) = '' THEN '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW'
    ELSE e.password_hash
  END,
  role = COALESCE(e.role, 'employee'::laundry_backend."Role")
WHERE e.email IS NULL OR trim(e.email) = '';

ALTER TABLE laundry_backend.employees ALTER COLUMN email SET NOT NULL;
ALTER TABLE laundry_backend.employees ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE laundry_backend.employees ALTER COLUMN role SET NOT NULL;

DROP INDEX IF EXISTS laundry_backend_users_email_idx;

CREATE UNIQUE INDEX IF NOT EXISTS laundry_backend_employees_email_lower_idx
  ON laundry_backend.employees (lower(email));

DROP TABLE IF EXISTS laundry_backend.users;
