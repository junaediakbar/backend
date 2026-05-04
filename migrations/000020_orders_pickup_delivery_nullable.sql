-- Tiga opsi: NULL = belum tahu, false = tidak, true = ya
ALTER TABLE laundry_backend.orders
  ALTER COLUMN pickup_delivery DROP NOT NULL;

ALTER TABLE laundry_backend.orders
  ALTER COLUMN pickup_delivery SET DEFAULT NULL;
