-- Nama tugas kanonikal (UI / workflow) yang belum ada di enum WorkTaskType.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type t WHERE t.typname = 'WorkTaskType') THEN
    BEGIN
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'rontok';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'sikat';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'bilas';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'jemur';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'downy';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'rumbai';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_bensin';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_buruh_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_buruh_2';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_antar_jemput';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_bensin';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_buruh_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_buruh_2';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_antar_jemput';
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
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'rontok';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'sikat';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'bilas';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'jemur';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'downy';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'rumbai';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_bensin';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_buruh_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_buruh_2';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_antar_jemput';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_bensin';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_buruh_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_buruh_2';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_antar_jemput';
  END IF;
END $$;
