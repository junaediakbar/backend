DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'users'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'laundry_backend'
        AND table_name = 'users'
        AND column_name = 'password_hash'
    ) THEN
      ALTER TABLE laundry_backend.users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'laundry_backend'
        AND table_name = 'users'
        AND column_name = 'is_active'
    ) THEN
      ALTER TABLE laundry_backend.users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
    END IF;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS laundry_backend_users_email_idx ON laundry_backend.users(email);
