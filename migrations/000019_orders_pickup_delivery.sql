-- Antar-jemput: flag per nota (Ya/Tidak di UI).
ALTER TABLE laundry_backend.orders
ADD COLUMN IF NOT EXISTS pickup_delivery BOOLEAN NOT NULL DEFAULT false;
