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
      ALTER COLUMN latitude TYPE DECIMAL(10,8) USING latitude::DECIMAL(10,8);
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'laundry_backend'
      AND table_name = 'customers'
      AND column_name = 'longitude'
  ) THEN
    ALTER TABLE laundry_backend.customers
      ALTER COLUMN longitude TYPE DECIMAL(11,8) USING longitude::DECIMAL(11,8);
  END IF;
END $$;
