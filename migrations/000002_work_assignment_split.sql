DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'WorkTaskType') THEN
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_fuel';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_driver';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_1';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'pickup_worker_2';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_fuel';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_driver';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_1';
    ALTER TYPE "WorkTaskType" ADD VALUE IF NOT EXISTS 'dropoff_worker_2';
  END IF;
END $$;
