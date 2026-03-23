ALTER TABLE laundry_backend.orders
ADD COLUMN IF NOT EXISTS public_token TEXT;

UPDATE laundry_backend.orders
SET public_token = md5(random()::text || clock_timestamp()::text || id)
WHERE public_token IS NULL OR public_token = '';

ALTER TABLE laundry_backend.orders
ALTER COLUMN public_token SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS laundry_backend_orders_public_token_uq
ON laundry_backend.orders(public_token);
