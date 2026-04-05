-- Sesuaikan persentase upah pekerja pada semua penugasan yang ada dengan nilai standar terbaru
-- dan hitung ulang amount dari subtotal order_items.
-- Produksi: Rontok 4%, Sikat 5%, Bilas 6%, Downy 2%, Rumbai 2%, Finishing 3%, Jemur/spin 8%
-- Jemput/Antar: tetap 7,5% per grup (3 + 1,5 + 0 + 3)

DO $$
BEGIN
  IF to_regclass('laundry_backend.work_assignments') IS NULL THEN
    RETURN;
  END IF;

  -- Produksi & alias Inggris
  UPDATE laundry_backend.work_assignments wa
  SET percent = 4.00,
      amount = ROUND(oi.total::numeric * (4.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('rontok', 'dust_removal');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 5.00,
      amount = ROUND(oi.total::numeric * (5.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('sikat', 'brushing');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 6.00,
      amount = ROUND(oi.total::numeric * (6.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('bilas', 'rinse_sprayer');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 8.00,
      amount = ROUND(oi.total::numeric * (8.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('jemur', 'spin_dry', 'spin_dry_1', 'spin_dry_2');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 2.00,
      amount = ROUND(oi.total::numeric * (2.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'downy';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 2.00,
      amount = ROUND(oi.total::numeric * (2.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'rumbai';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 3.00,
      amount = ROUND(oi.total::numeric * (3.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('finishing_1', 'finishing_2', 'finishing_packing');

  -- Jemput
  UPDATE laundry_backend.work_assignments wa
  SET percent = 7.50,
      amount = ROUND(oi.total::numeric * (7.50 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'pickup_antar_jemput';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 3.00,
      amount = ROUND(oi.total::numeric * (3.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'pickup_driver';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 1.50,
      amount = ROUND(oi.total::numeric * (1.50 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('pickup_buruh_1', 'pickup_worker_1');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 0.00,
      amount = ROUND(oi.total::numeric * (0.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('pickup_buruh_2', 'pickup_worker_2');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 3.00,
      amount = ROUND(oi.total::numeric * (3.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('pickup_bensin', 'pickup_fuel');

  -- Antar
  UPDATE laundry_backend.work_assignments wa
  SET percent = 7.50,
      amount = ROUND(oi.total::numeric * (7.50 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'dropoff_antar_jemput';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 3.00,
      amount = ROUND(oi.total::numeric * (3.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text = 'dropoff_driver';

  UPDATE laundry_backend.work_assignments wa
  SET percent = 1.50,
      amount = ROUND(oi.total::numeric * (1.50 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('dropoff_buruh_1', 'dropoff_worker_1');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 0.00,
      amount = ROUND(oi.total::numeric * (0.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('dropoff_buruh_2', 'dropoff_worker_2');

  UPDATE laundry_backend.work_assignments wa
  SET percent = 3.00,
      amount = ROUND(oi.total::numeric * (3.00 / 100.0), 2)
  FROM laundry_backend.order_items oi
  WHERE wa.order_item_id = oi.id
    AND wa.task_type::text IN ('dropoff_bensin', 'dropoff_fuel');
END $$;
