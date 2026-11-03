-- name: ListProducts :many
SELECT * FROM products ORDER BY id LIMIT $1 OFFSET $2;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (
  name, code, category_id, weight, gold_price_at_time, labor_cost, stone_cost, stock,
  markup_rate, selling_price, warranty_period, image, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
)
RETURNING *;

-- name: UpdateProductByCode :one
UPDATE products
SET
  name              = COALESCE($1, name),
  category_id       = COALESCE($2, category_id),
  weight            = COALESCE($3, weight),
  gold_price_at_time= COALESCE($4, gold_price_at_time),
  labor_cost        = COALESCE($5, labor_cost),
  stone_cost        = COALESCE($6, stone_cost),
  buy_turn          = COALESCE($7, buy_turn),
  markup_rate       = COALESCE($8, markup_rate),
  selling_price     = COALESCE($9, selling_price),
  warranty_period   = COALESCE($10, warranty_period),
  image             = COALESCE($11, image),
  stock             = COALESCE($12, stock),
  updated_at        = NOW()
WHERE code = $13
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: GetProductsById :many 
SELECT * FROM products WHERE id = ANY($1::int[]);