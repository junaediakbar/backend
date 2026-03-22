DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = current_schema()
      AND table_name = 'customers'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'phone'
    ) THEN
      ALTER TABLE customers ADD COLUMN phone TEXT;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'address'
    ) THEN
      ALTER TABLE customers ADD COLUMN address TEXT;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'email'
    ) THEN
      ALTER TABLE customers ADD COLUMN email TEXT;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'notes'
    ) THEN
      ALTER TABLE customers ADD COLUMN notes TEXT;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'latitude'
    ) THEN
      ALTER TABLE customers ADD COLUMN latitude DECIMAL(9,6);
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'customers'
          AND column_name = 'lat'
      ) THEN
        EXECUTE 'UPDATE customers SET latitude = lat WHERE latitude IS NULL';
      END IF;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'longitude'
    ) THEN
      ALTER TABLE customers ADD COLUMN longitude DECIMAL(9,6);
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'customers'
          AND column_name = 'lng'
      ) THEN
        EXECUTE 'UPDATE customers SET longitude = lng WHERE longitude IS NULL';
      END IF;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'created_at'
    ) THEN
      ALTER TABLE customers ADD COLUMN created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'customers'
        AND column_name = 'updated_at'
    ) THEN
      ALTER TABLE customers ADD COLUMN updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP;
    END IF;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS customers_name_idx ON customers(name);
