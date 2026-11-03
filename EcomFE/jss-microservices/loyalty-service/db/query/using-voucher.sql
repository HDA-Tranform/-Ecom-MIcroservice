-- name: UpsertUsageRecord :one
INSERT INTO usage_records (
    customer_id,
    voucher_id,
    order_id,
    status,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (customer_id, voucher_id, order_id)
DO UPDATE SET
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at
RETURNING *;
