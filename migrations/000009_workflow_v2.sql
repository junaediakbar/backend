-- Hanya menambah nilai enum. UPDATE data dipindah ke 000010 (PG melarang pakai nilai enum baru
-- dalam transaksi yang sama dengan ADD VALUE — SQLSTATE 55P04).
ALTER TYPE laundry_backend."WorkflowStatus" ADD VALUE IF NOT EXISTS 'rontok_done';
ALTER TYPE laundry_backend."WorkflowStatus" ADD VALUE IF NOT EXISTS 'jemur_done';
ALTER TYPE laundry_backend."WorkflowStatus" ADD VALUE IF NOT EXISTS 'downy_done';
ALTER TYPE laundry_backend."WorkflowStatus" ADD VALUE IF NOT EXISTS 'packing_done';
