ALTER TABLE laundry_backend.orders
  ADD COLUMN IF NOT EXISTS delivery_service_category TEXT NULL,
  ADD COLUMN IF NOT EXISTS delivery_estimate_days INTEGER NULL;
