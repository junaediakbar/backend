-- Isi nilai default untuk nota yang sudah ada sebelum fitur layanan pengiriman.
-- Tidak menghapus atau mengubah data lain (invoice, item, pembayaran, dll).
UPDATE laundry_backend.orders
SET
  delivery_service_category = 'reguler',
  delivery_estimate_days = 7
WHERE delivery_service_category IS NULL;
