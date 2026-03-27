DO $$
BEGIN
  -- Support database setups where enum was created without schema qualification.
  IF EXISTS (SELECT 1 FROM pg_type t WHERE t.typname = 'WorkTaskType') THEN
    BEGIN
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_fuel';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_driver';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_2';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_fuel';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_driver';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_1';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_2';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dust_removal';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'brushing';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'rinse_sprayer';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry';
      ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_packing';
    EXCEPTION WHEN undefined_object THEN
      -- Type exists but is not visible in search_path; ignore here and handle schema-qualified below.
      NULL;
    END;
  END IF;

  -- Support database setups where everything lives under schema `laundry_backend`.
  IF EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'WorkTaskType' AND n.nspname = 'laundry_backend'
  ) THEN
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_fuel';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_driver';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_2';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_fuel';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_driver';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_1';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_2';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'dust_removal';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'brushing';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'rinse_sprayer';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'spin_dry';
    ALTER TYPE laundry_backend."WorkTaskType" ADD VALUE IF NOT EXISTS 'finishing_packing';
  END IF;
END $$;

