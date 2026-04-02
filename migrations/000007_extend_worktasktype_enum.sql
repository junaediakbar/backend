DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type t WHERE t.typname = 'WorkTaskType') THEN
    BEGIN
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry_2';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_2';
    EXCEPTION WHEN undefined_object THEN
      NULL;
    END;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'WorkTaskType' AND n.nspname = 'laundry_backend'
  ) THEN
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry_2';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_2';
  END IF;
END $$;

