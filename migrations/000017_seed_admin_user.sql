-- Seed admin user
-- Email: jun@treeslaundry.com
-- Password: 123456

INSERT INTO laundry_backend.employees (id, name, email, password_hash, role, is_active, created_at, updated_at)
VALUES (
  'clq4h9k2o0000352h6g1j8vhd', -- CUID format ID
  'Jun',
  'jun@treeslaundry.com',
  '$2a$10$TOl1ux8tLd/yHGAlu6lKlen6QPKD8gihOQjVZfNZCbZjORn73lmp2', -- bcrypt hash for "123456"
  'admin'::laundry_backend."Role",
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO NOTHING; -- Prevent duplicate inserts if migration is re-run
