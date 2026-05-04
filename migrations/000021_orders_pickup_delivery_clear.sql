-- Kosongkan pilihan antar jemput untuk semua nota (NULL = "-"/belum tahu di UI).
UPDATE laundry_backend.orders
SET pickup_delivery = NULL;
