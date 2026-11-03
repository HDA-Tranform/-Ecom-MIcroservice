TRUNCATE TABLE product_categories CASCADE;
TRUNCATE TABLE products CASCADE;
TRUNCATE TABLE customers CASCADE;

INSERT INTO product_categories (name)
VALUES 
  ('Nhẫn'),
  ('Dây Chuyền'),
  ('Mặt Dây Chuyền'),
  ('Bông Tai'),
  ('Lắc'),
  ('Vòng'),
  ('Charm'),
  ('Dây Cổ')
ON CONFLICT (name) 
DO UPDATE SET name = EXCLUDED.name;

-- Seed data for products
INSERT INTO products 
(name, code, category_id, weight, gold_price_at_time, labor_cost, stone_cost, markup_rate, selling_price, warranty_period, image, created_at, updated_at, stock) 
VALUES
('Gold Ring Classic', 'PRD001', 1, 5.20, 5600000, 500000, 0, 10.00, 6700000, 12, 'ring1.jpg', NOW(), NOW(), 10),
('Diamond Necklace', 'PRD002', 2, 12.50, 5600000, 1200000, 3500000, 15.00, 13000000, 24, 'necklace1.jpg', NOW(), NOW(), 10),
('Gold Bracelet', 'PRD003', 1, 8.75, 5600000, 800000, 0, 8.50, 9000000, 6, 'bracelet1.jpg', NOW(), NOW(), 10),
('Platinum Earrings', 'PRD004', 3, 3.40, 7500000, 400000, 500000, 12.00, 8900000, 18, 'earring1.jpg', NOW(), NOW(), 10);

-- Seed data for customers
INSERT INTO customers 
(name, phone, email, address, created_at, updated_at) 
VALUES
('Nguyen Van A', '0901234567', 'vana@example.com', '123 Nguyen Trai, Ha Noi', NOW(), NOW()),
('Tran Thi B', '0912345678', 'thib@example.com', '45 Le Loi, Ho Chi Minh', NOW(), NOW()),
('Le Van C', '0923456789', 'vanc@example.com', '78 Tran Hung Dao, Da Nang', NOW(), NOW()),
('Pham Thi D', '0934567890', 'thid@example.com', '12 Phan Boi Chau, Hue', NOW(), NOW());
