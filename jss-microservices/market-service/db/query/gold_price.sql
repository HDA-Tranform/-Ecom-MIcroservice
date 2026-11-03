-- name: CreateGoldPrice :one
INSERT INTO gold_prices (date, gold_type, buy_price, sell_price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGoldPrice :one
SELECT * FROM gold_prices
WHERE id = $1
LIMIT 1;

-- name: GetGoldPrices :many
SELECT * FROM gold_prices
ORDER BY date DESC
LIMIT $1 OFFSET $2;

-- name: GetGoldPricesByType :many
SELECT * FROM gold_prices
WHERE gold_type = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: GetLatestGoldPrices :many
SELECT DISTINCT ON (gold_id) *
FROM gold_prices
ORDER BY gold_id, date DESC;

-- name: UpdateGoldPrice :one
UPDATE gold_prices
SET date = $2, gold_type = $3, buy_price = $4, sell_price = $5
WHERE id = $1
RETURNING *;

-- name: DeleteGoldPrice :exec
DELETE FROM gold_prices
WHERE id = $1;

-- name: CountGoldPrices :one
SELECT COUNT(*) FROM gold_prices;

-- name: CountGoldPricesByType :one
SELECT COUNT(*) FROM gold_prices
WHERE gold_type = $1;

-- name: UpsertPrice :one
INSERT INTO gold_prices (date, gold_id, gold_type, buy_price, sell_price)
	VALUES ($1, $2, $3, $4, $5)
RETURNING *;