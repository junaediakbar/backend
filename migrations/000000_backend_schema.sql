CREATE SCHEMA IF NOT EXISTS laundry_backend;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'Role' AND n.nspname = 'laundry_backend'
  ) THEN
    CREATE TYPE laundry_backend."Role" AS ENUM ('owner', 'admin', 'cashier');
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'PaymentStatus' AND n.nspname = 'laundry_backend'
  ) THEN
    CREATE TYPE laundry_backend."PaymentStatus" AS ENUM ('unpaid', 'partial', 'paid');
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'WorkflowStatus' AND n.nspname = 'laundry_backend'
  ) THEN
    CREATE TYPE laundry_backend."WorkflowStatus" AS ENUM ('received', 'washing', 'drying', 'ironing', 'finished', 'picked_up');
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'WorkTaskType' AND n.nspname = 'laundry_backend'
  ) THEN
    CREATE TYPE laundry_backend."WorkTaskType" AS ENUM ('pickup_fuel', 'pickup_driver', 'pickup_worker_1', 'pickup_worker_2', 'dropoff_fuel', 'dropoff_driver', 'dropoff_worker_1', 'dropoff_worker_2', 'dust_removal', 'brushing', 'rinse_sprayer', 'spin_dry', 'finishing_packing');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS laundry_backend.users (
  id TEXT PRIMARY KEY,
  auth_user_id TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  role laundry_backend."Role" NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS laundry_backend.employees (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS laundry_backend_employees_name_idx ON laundry_backend.employees(name);

CREATE TABLE IF NOT EXISTS laundry_backend.customers (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  phone TEXT,
  address TEXT,
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  email TEXT,
  notes TEXT,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS laundry_backend_customers_name_idx ON laundry_backend.customers(name);

CREATE TABLE IF NOT EXISTS laundry_backend.service_types (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  unit TEXT NOT NULL,
  default_price DECIMAL(12,2) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS laundry_backend.orders (
  id TEXT PRIMARY KEY,
  invoice_number TEXT NOT NULL UNIQUE,
  customer_id TEXT NOT NULL,
  total DECIMAL(12,2) NOT NULL,
  payment_status laundry_backend."PaymentStatus" NOT NULL DEFAULT 'unpaid',
  workflow_status laundry_backend."WorkflowStatus" NOT NULL DEFAULT 'received',
  received_date TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_date TIMESTAMP(3),
  pickup_date TIMESTAMP(3),
  note TEXT,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES laundry_backend.customers(id) ON DELETE RESTRICT ON UPDATE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS laundry_backend_orders_invoice_number_key ON laundry_backend.orders(invoice_number);
CREATE INDEX IF NOT EXISTS laundry_backend_orders_invoice_number_idx ON laundry_backend.orders(invoice_number);
CREATE INDEX IF NOT EXISTS laundry_backend_orders_payment_status_workflow_status_idx ON laundry_backend.orders(payment_status, workflow_status);

CREATE TABLE IF NOT EXISTS laundry_backend.order_items (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  service_type_id TEXT NOT NULL,
  quantity DECIMAL(12,2) NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  discount DECIMAL(12,2) NOT NULL DEFAULT 0,
  total DECIMAL(12,2) NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES laundry_backend.orders(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_order_items_service_type_id_fkey FOREIGN KEY (service_type_id) REFERENCES laundry_backend.service_types(id) ON DELETE RESTRICT ON UPDATE CASCADE
);
CREATE INDEX IF NOT EXISTS laundry_backend_order_items_order_id_idx ON laundry_backend.order_items(order_id);
CREATE INDEX IF NOT EXISTS laundry_backend_order_items_service_type_id_idx ON laundry_backend.order_items(service_type_id);

CREATE TABLE IF NOT EXISTS laundry_backend.work_assignments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  order_item_id TEXT NOT NULL,
  employee_id TEXT NOT NULL,
  task_type laundry_backend."WorkTaskType" NOT NULL,
  percent DECIMAL(5,2) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_work_assignments_order_id_fkey FOREIGN KEY (order_id) REFERENCES laundry_backend.orders(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_work_assignments_order_item_id_fkey FOREIGN KEY (order_item_id) REFERENCES laundry_backend.order_items(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_work_assignments_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES laundry_backend.employees(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_work_assignments_order_item_id_task_type_key UNIQUE (order_item_id, task_type)
);
CREATE INDEX IF NOT EXISTS laundry_backend_work_assignments_order_id_idx ON laundry_backend.work_assignments(order_id);
CREATE INDEX IF NOT EXISTS laundry_backend_work_assignments_employee_id_idx ON laundry_backend.work_assignments(employee_id);

CREATE TABLE IF NOT EXISTS laundry_backend.payments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  method TEXT NOT NULL,
  paid_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  note TEXT,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_payments_order_id_fkey FOREIGN KEY (order_id) REFERENCES laundry_backend.orders(id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE INDEX IF NOT EXISTS laundry_backend_payments_order_id_idx ON laundry_backend.payments(order_id);

CREATE TABLE IF NOT EXISTS laundry_backend.order_attachments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  file_path TEXT NOT NULL,
  mime_type TEXT,
  size_bytes INTEGER,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_order_attachments_order_id_fkey FOREIGN KEY (order_id) REFERENCES laundry_backend.orders(id) ON DELETE CASCADE ON UPDATE CASCADE
);
CREATE INDEX IF NOT EXISTS laundry_backend_order_attachments_order_id_idx ON laundry_backend.order_attachments(order_id);

CREATE TABLE IF NOT EXISTS laundry_backend.delivery_plans (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  planned_date TIMESTAMP(3) NOT NULL,
  start_address TEXT,
  start_lat DECIMAL(9,6),
  start_lng DECIMAL(9,6),
  end_address TEXT,
  end_lat DECIMAL(9,6),
  end_lng DECIMAL(9,6),
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS laundry_backend_delivery_plans_planned_date_idx ON laundry_backend.delivery_plans(planned_date);

CREATE TABLE IF NOT EXISTS laundry_backend.delivery_stops (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL,
  customer_id TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  distance_km DECIMAL(10,2),
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT laundry_backend_delivery_stops_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES laundry_backend.delivery_plans(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_delivery_stops_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES laundry_backend.customers(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT laundry_backend_delivery_stops_plan_id_customer_id_key UNIQUE (plan_id, customer_id),
  CONSTRAINT laundry_backend_delivery_stops_plan_id_sequence_key UNIQUE (plan_id, sequence)
);
CREATE INDEX IF NOT EXISTS laundry_backend_delivery_stops_customer_id_idx ON laundry_backend.delivery_stops(customer_id);
