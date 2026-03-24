DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'delivery_plans'
      AND column_name = 'end_address'
  ) THEN
    ALTER TABLE laundry_backend.delivery_plans ADD COLUMN end_address TEXT;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'delivery_plans'
      AND column_name = 'end_lat'
  ) THEN
    ALTER TABLE laundry_backend.delivery_plans ADD COLUMN end_lat DECIMAL(9,6);
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'delivery_plans'
      AND column_name = 'end_lng'
  ) THEN
    ALTER TABLE laundry_backend.delivery_plans ADD COLUMN end_lng DECIMAL(9,6);
  END IF;
END $$;
