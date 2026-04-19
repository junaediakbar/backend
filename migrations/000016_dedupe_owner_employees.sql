-- Satukan beberapa akun Owner menjadi satu baris employees yang valid.
-- Prioritas baris yang dipertahankan: role = owner, email "nyata" (bukan placeholder migrasi/seed), lalu created_at paling awal.

DO $$
DECLARE
  keep_id TEXT;
BEGIN
  SELECT e.id INTO keep_id
  FROM laundry_backend.employees e
  WHERE e.role = 'owner'::laundry_backend."Role"
     OR lower(trim(e.name)) = 'owner'
  ORDER BY
    CASE WHEN e.role = 'owner'::laundry_backend."Role" THEN 0 ELSE 1 END,
    CASE
      WHEN e.email IS NOT NULL
        AND e.email NOT ILIKE '%@migrated.local'
        AND e.email NOT ILIKE '%@seed.local'
        AND e.email NOT ILIKE 'pending+%@%'
      THEN 0
      ELSE 1
    END,
    e.created_at ASC NULLS LAST,
    e.id ASC
  LIMIT 1;

  IF keep_id IS NULL THEN
    RETURN;
  END IF;

  -- Alihkan penugasan dari duplikat Owner ke akun kanonik
  UPDATE laundry_backend.work_assignments wa
  SET employee_id = keep_id
  WHERE wa.employee_id IN (
    SELECT e.id
    FROM laundry_backend.employees e
    WHERE e.id <> keep_id
      AND (
        e.role = 'owner'::laundry_backend."Role"
        OR lower(trim(e.name)) = 'owner'
      )
  );

  -- Hapus baris Owner duplikat (bukan kanonik)
  DELETE FROM laundry_backend.employees e
  WHERE e.id <> keep_id
    AND (
      e.role = 'owner'::laundry_backend."Role"
      OR lower(trim(e.name)) = 'owner'
    );

  -- Normalisasi satu-satunya Owner
  UPDATE laundry_backend.employees
  SET
    name = 'Owner',
    role = 'owner'::laundry_backend."Role",
    updated_at = now()
  WHERE id = keep_id;
END $$;

-- Hanya boleh ada satu baris ber-role owner
CREATE UNIQUE INDEX IF NOT EXISTS laundry_backend_employees_single_owner_idx
  ON laundry_backend.employees ((1))
  WHERE role = 'owner'::laundry_backend."Role";
