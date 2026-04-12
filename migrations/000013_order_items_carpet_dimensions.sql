ALTER TABLE laundry_backend.order_items
  ADD COLUMN IF NOT EXISTS carpet_length_m DECIMAL(12,2) NULL,
  ADD COLUMN IF NOT EXISTS carpet_width_m DECIMAL(12,2) NULL;
