-- name: CreateBuybackPolicy :one
INSERT INTO buyback_policies (product_type, buyback_rate, description, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBuybackPolicy :one
SELECT * FROM buyback_policies
WHERE id = $1
LIMIT 1;

-- name: GetBuybackPolicies :many
SELECT * FROM buyback_policies
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetBuybackPoliciesByProductType :many
SELECT * FROM buyback_policies
WHERE product_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetActiveBuybackPolicies :many
SELECT * FROM buyback_policies
WHERE buyback_rate > 0
ORDER BY buyback_rate DESC
LIMIT $1 OFFSET $2;

-- name: UpdateBuybackPolicy :one
UPDATE buyback_policies
SET product_type = $2, buyback_rate = $3, description = $4, created_at = $5
WHERE id = $1
RETURNING *;

-- name: UpdateBuybackRate :one
UPDATE buyback_policies
SET buyback_rate = $2
WHERE id = $1
RETURNING *;

-- name: DeleteBuybackPolicy :exec
DELETE FROM buyback_policies
WHERE id = $1;

-- name: CountBuybackPolicies :one
SELECT COUNT(*) FROM buyback_policies;

-- name: CountBuybackPoliciesByProductType :one
SELECT COUNT(*) FROM buyback_policies
WHERE product_type = $1;

-- ===== ADDITIONAL USEFUL QUERIES =====

-- name: GetGoldPricesByDateRange :many
SELECT * FROM gold_prices
WHERE date BETWEEN $1 AND $2
ORDER BY date DESC
LIMIT $3 OFFSET $4;

-- name: GetGoldPricesWithHighestBuyPrice :many
SELECT * FROM gold_prices
ORDER BY buy_price DESC
LIMIT $1 OFFSET $2;

-- name: SearchGoldPricesByType :many
SELECT * FROM gold_prices
WHERE gold_type ILIKE '%' || $1 || '%'
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: SearchBuybackPoliciesByProductType :many
SELECT * FROM buyback_policies
WHERE product_type ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;