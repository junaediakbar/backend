CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'Role') THEN
    CREATE TYPE "Role" AS ENUM ('owner', 'admin', 'cashier');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'PaymentStatus') THEN
    CREATE TYPE "PaymentStatus" AS ENUM ('unpaid', 'partial', 'paid');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'WorkflowStatus') THEN
    CREATE TYPE "WorkflowStatus" AS ENUM ('received', 'washing', 'drying', 'ironing', 'finished', 'picked_up');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'WorkTaskType') THEN
    CREATE TYPE "WorkTaskType" AS ENUM ('pickup_fuel', 'pickup_driver', 'pickup_worker_1', 'pickup_worker_2', 'dropoff_fuel', 'dropoff_driver', 'dropoff_worker_1', 'dropoff_worker_2', 'dust_removal', 'brushing', 'rinse_sprayer', 'spin_dry', 'finishing_packing');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  auth_user_id TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  role "Role" NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS employees (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS employees_name_idx ON employees(name);

CREATE TABLE IF NOT EXISTS customers (
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
CREATE INDEX IF NOT EXISTS customers_name_idx ON customers(name);

CREATE TABLE IF NOT EXISTS service_types (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  unit TEXT NOT NULL,
  default_price DECIMAL(12,2) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  invoice_number TEXT NOT NULL UNIQUE,
  customer_id TEXT NOT NULL,
  service_type_id TEXT,
  quantity DECIMAL(12,2),
  unit_price DECIMAL(12,2),
  discount DECIMAL(12,2) DEFAULT 0,
  total DECIMAL(12,2) NOT NULL,
  payment_status "PaymentStatus" NOT NULL DEFAULT 'unpaid',
  workflow_status "WorkflowStatus" NOT NULL DEFAULT 'received',
  received_date TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_date TIMESTAMP(3),
  pickup_date TIMESTAMP(3),
  note TEXT,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT orders_service_type_id_fkey FOREIGN KEY (service_type_id) REFERENCES service_types(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = current_schema()
      AND table_name = 'orders'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'orders'
        AND column_name = 'invoice_number'
    ) THEN
      ALTER TABLE orders ADD COLUMN invoice_number TEXT;

      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'orders'
          AND column_name = 'invoiceNumber'
      ) THEN
        EXECUTE 'UPDATE orders SET invoice_number = "invoiceNumber" WHERE invoice_number IS NULL';
      END IF;

      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'orders'
          AND column_name = 'invoice'
      ) THEN
        EXECUTE 'UPDATE orders SET invoice_number = invoice WHERE invoice_number IS NULL';
      END IF;

      EXECUTE 'UPDATE orders SET invoice_number = COALESCE(invoice_number, ''LDR-'' || id)';
      ALTER TABLE orders ALTER COLUMN invoice_number SET NOT NULL;
    ELSE
      EXECUTE 'UPDATE orders SET invoice_number = COALESCE(invoice_number, ''LDR-'' || id) WHERE invoice_number IS NULL';
      ALTER TABLE orders ALTER COLUMN invoice_number SET NOT NULL;
    END IF;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = current_schema()
      AND table_name = 'orders'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'orders'
        AND column_name = 'payment_status'
    ) THEN
      ALTER TABLE orders ADD COLUMN payment_status "PaymentStatus" NOT NULL DEFAULT 'unpaid';
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'orders'
          AND column_name = 'paymentStatus'
      ) THEN
        EXECUTE 'UPDATE orders SET payment_status = ("paymentStatus"::text)::"PaymentStatus"';
      END IF;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'orders'
        AND column_name = 'workflow_status'
    ) THEN
      ALTER TABLE orders ADD COLUMN workflow_status "WorkflowStatus" NOT NULL DEFAULT 'received';
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'orders'
          AND column_name = 'workflowStatus'
      ) THEN
        EXECUTE 'UPDATE orders SET workflow_status = ("workflowStatus"::text)::"WorkflowStatus"';
      END IF;
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'orders'
          AND column_name = 'status'
      ) THEN
        EXECUTE 'UPDATE orders SET workflow_status = (status::text)::"WorkflowStatus" WHERE workflow_status IS NULL';
      END IF;
      EXECUTE 'UPDATE orders SET workflow_status = COALESCE(workflow_status, ''received''::"WorkflowStatus")';
    END IF;
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS orders_invoice_number_key ON orders(invoice_number);
CREATE INDEX IF NOT EXISTS orders_invoice_number_idx ON orders(invoice_number);
CREATE INDEX IF NOT EXISTS orders_payment_status_workflow_status_idx ON orders(payment_status, workflow_status);

CREATE TABLE IF NOT EXISTS order_items (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  service_type_id TEXT NOT NULL,
  quantity DECIMAL(12,2) NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  discount DECIMAL(12,2) NOT NULL DEFAULT 0,
  total DECIMAL(12,2) NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT order_items_service_type_id_fkey FOREIGN KEY (service_type_id) REFERENCES service_types(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = current_schema()
      AND table_name = 'order_items'
  ) THEN
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'order_id'
    ) THEN
      ALTER TABLE order_items ADD COLUMN order_id TEXT;
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'order_items'
          AND column_name = 'orderId'
      ) THEN
        EXECUTE 'UPDATE order_items SET order_id = "orderId" WHERE order_id IS NULL';
      END IF;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'service_type_id'
    ) THEN
      ALTER TABLE order_items ADD COLUMN service_type_id TEXT;
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'order_items'
          AND column_name = 'serviceTypeId'
      ) THEN
        EXECUTE 'UPDATE order_items SET service_type_id = "serviceTypeId" WHERE service_type_id IS NULL';
      END IF;
      IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'order_items'
          AND column_name = 'service'
      ) THEN
        EXECUTE 'UPDATE order_items SET service_type_id = service WHERE service_type_id IS NULL';
      END IF;
    END IF;

    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'quantity'
    ) THEN
      ALTER TABLE order_items ADD COLUMN quantity DECIMAL(12,2);
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'unit_price'
    ) THEN
      ALTER TABLE order_items ADD COLUMN unit_price DECIMAL(12,2);
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'discount'
    ) THEN
      ALTER TABLE order_items ADD COLUMN discount DECIMAL(12,2) DEFAULT 0;
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'total'
    ) THEN
      ALTER TABLE order_items ADD COLUMN total DECIMAL(12,2);
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'created_at'
    ) THEN
      ALTER TABLE order_items ADD COLUMN created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP;
    END IF;
    IF NOT EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = current_schema()
        AND table_name = 'order_items'
        AND column_name = 'updated_at'
    ) THEN
      ALTER TABLE order_items ADD COLUMN updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP;
    END IF;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON order_items(order_id);
CREATE INDEX IF NOT EXISTS order_items_service_type_id_idx ON order_items(service_type_id);

CREATE TABLE IF NOT EXISTS work_assignments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  order_item_id TEXT NOT NULL,
  employee_id TEXT NOT NULL,
  task_type "WorkTaskType" NOT NULL,
  percent DECIMAL(5,2) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT work_assignments_order_item_id_task_type_key UNIQUE (order_item_id, task_type)
);
CREATE INDEX IF NOT EXISTS work_assignments_order_id_idx ON work_assignments(order_id);
CREATE INDEX IF NOT EXISTS work_assignments_employee_id_idx ON work_assignments(employee_id);

CREATE TABLE IF NOT EXISTS payments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  method TEXT NOT NULL,
  paid_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  note TEXT,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS payments_order_id_idx ON payments(order_id);

CREATE TABLE IF NOT EXISTS order_attachments (
  id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL,
  file_path TEXT NOT NULL,
  mime_type TEXT,
  size_bytes INTEGER,
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS order_attachments_order_id_idx ON order_attachments(order_id);

CREATE TABLE IF NOT EXISTS delivery_plans (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  planned_date TIMESTAMP(3) NOT NULL,
  start_address TEXT,
  start_lat DECIMAL(9,6),
  start_lng DECIMAL(9,6),
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS delivery_plans_planned_date_idx ON delivery_plans(planned_date);

CREATE TABLE IF NOT EXISTS delivery_stops (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL,
  customer_id TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  distance_km DECIMAL(10,2),
  created_at TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT delivery_stops_plan_id_customer_id_key UNIQUE (plan_id, customer_id),
  CONSTRAINT delivery_stops_plan_id_sequence_key UNIQUE (plan_id, sequence)
);
CREATE INDEX IF NOT EXISTS delivery_stops_customer_id_idx ON delivery_stops(customer_id);
