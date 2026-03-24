DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'customers'
      AND column_name = 'latitude'
  ) THEN
    ALTER TABLE laundry_backend.customers
      ALTER COLUMN latitude TYPE DOUBLE PRECISION USING latitude::DOUBLE PRECISION;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'customers'
      AND column_name = 'longitude'
  ) THEN
    ALTER TABLE laundry_backend.customers
      ALTER COLUMN longitude TYPE DOUBLE PRECISION USING longitude::DOUBLE PRECISION;
  END IF;
END $$;
