-- Jalankan setelah 000009 berhasil (nilai enum sudah ter-commit).
UPDATE laundry_backend.orders SET workflow_status = 'rontok_done'::laundry_backend."WorkflowStatus"
WHERE workflow_status::text = 'washing';

UPDATE laundry_backend.orders SET workflow_status = 'jemur_done'::laundry_backend."WorkflowStatus"
WHERE workflow_status::text = 'drying';

UPDATE laundry_backend.orders SET workflow_status = 'downy_done'::laundry_backend."WorkflowStatus"
WHERE workflow_status::text = 'ironing';

UPDATE laundry_backend.orders SET workflow_status = 'packing_done'::laundry_backend."WorkflowStatus"
WHERE workflow_status::text = 'finished';
