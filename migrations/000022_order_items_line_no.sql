-- Urutan baris item di nota (stabil untuk mapping gambar item saat create).
ALTER TABLE laundry_backend.order_items
  ADD COLUMN IF NOT EXISTS line_no INTEGER;

UPDATE laundry_backend.order_items oi
SET line_no = ranked.rn
FROM (
  SELECT
    id,
    ROW_NUMBER() OVER (
      PARTITION BY order_id
      ORDER BY created_at ASC, id ASC
    )::int AS rn
  FROM laundry_backend.order_items
) ranked
WHERE oi.id = ranked.id
  AND (oi.line_no IS NULL OR oi.line_no = 0);

UPDATE laundry_backend.order_items
SET line_no = 1
WHERE line_no IS NULL;

ALTER TABLE laundry_backend.order_items
  ALTER COLUMN line_no SET NOT NULL;

CREATE INDEX IF NOT EXISTS order_items_order_id_line_no_idx
  ON laundry_backend.order_items (order_id, line_no);
