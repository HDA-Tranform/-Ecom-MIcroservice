TRUNCATE TABLE loyalty_points CASCADE;
TRUNCATE TABLE vouchers CASCADE;
TRUNCATE TABLE customer_vouchers CASCADE; 
TRUNCATE TABLE usage_records CASCADE; 

-- Loyalty points
INSERT INTO "loyalty_points" ("customer_id", "points", "source", "reference_id", "created_at")
VALUES
  (1, 100, 'signup_bonus', NULL, NOW()),
  (2, 250, 'order', 101, NOW()),
  (3, 50,  'promotion', 202, NOW());

-- Vouchers
INSERT INTO "vouchers" ("code", "description", "discount_type", "discount_value", "start_date", "end_date", "usage_limit", "created_at")
VALUES
  ('WELCOME10', '10% off for new customers', 'PERCENTAGE', 10.00, '2025-09-01', '2025-12-31', 1, NOW()),
  ('FREESHIP50K', 'Free shipping up to 50,000đ', 'FIXED', 50000.00, '2025-09-01', '2025-10-15', 2, NOW()),
  ('VIP20', '20% discount for VIP customers', 'PERCENTAGE', 20.00, '2025-09-01', '2026-01-01', 3, NOW());

-- Customer vouchers
INSERT INTO "customer_vouchers" ("customer_id", "voucher_id", "status", "used_at")
VALUES
  (1, 1, 'unused', NULL),
  (2, 2, 'unused', NULL),
  (3, 3, 'unused', NULL),
  (1, 3, 'unused', NULL);

-- Usage records
INSERT INTO "usage_records" ("customer_id", "voucher_id", "order_id", "status", "created_at", "updated_at")
VALUES
  (1, 1, 1001, 'used', NOW(), NOW()),
  (2, 2, 1002, 'pending', NOW(), NOW()),
  (3, 3, 1003, 'used', NOW(), NOW()),
  (1, 3, 1004, 'pending', NOW(), NOW());