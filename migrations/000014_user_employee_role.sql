-- Role for karyawan (akses terbatas); hubungan user login ke master karyawan untuk performa.
DO $migrate$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_enum e
    JOIN pg_type t ON t.oid = e.enumtypid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'laundry_backend'
      AND t.typname = 'Role'
      AND e.enumlabel = 'employee'
  ) THEN
    ALTER TYPE laundry_backend."Role" ADD VALUE 'employee';
  END IF;
END
$migrate$;

ALTER TABLE laundry_backend.users
  ADD COLUMN IF NOT EXISTS employee_id TEXT REFERENCES laundry_backend.employees(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS laundry_backend_users_employee_id_idx
  ON laundry_backend.users(employee_id)
  WHERE employee_id IS NOT NULL;
