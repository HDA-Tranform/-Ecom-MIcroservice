-- name: CreateLoyaltyPoint :one
INSERT INTO loyalty_points (customer_id, points, source, reference_id, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLoyaltyPoint :one
SELECT * FROM loyalty_points
WHERE id = $1 LIMIT 1;

-- name: GetLoyaltyPointsByCustomer :many
SELECT * FROM loyalty_points
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLoyaltyPointsBySource :many
SELECT * FROM loyalty_points
WHERE source = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAllLoyaltyPoints :many
SELECT * FROM loyalty_points
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateLoyaltyPoints :one
UPDATE loyalty_points
SET points = $2, source = $3, reference_id = $4
WHERE id = $1
RETURNING *;

-- name: DeleteLoyaltyPoint :exec
DELETE FROM loyalty_points
WHERE id = $1;

-- name: GetCustomerTotalPoints :one
SELECT COALESCE(SUM(points), 0) as total_points
FROM loyalty_points
WHERE customer_id = $1;

-- name: CreateVoucher :one
INSERT INTO vouchers (code, description, discount_type, discount_value, start_date, end_date, usage_limit, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetVoucher :one
SELECT * FROM vouchers
WHERE id = $1 LIMIT 1;

-- name: GetVoucherByCode :one
SELECT * FROM vouchers
WHERE code = $1 LIMIT 1;

-- name: GetActiveVouchers :many
SELECT * FROM vouchers
WHERE start_date <= CURRENT_DATE AND end_date >= CURRENT_DATE
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAllVouchers :many
SELECT * FROM vouchers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateVoucher :one
UPDATE vouchers
SET code = $2, description = $3, discount_type = $4, discount_value = $5, 
    start_date = $6, end_date = $7, usage_limit = $8
WHERE id = $1
RETURNING *;

-- name: DeleteVoucher :exec
DELETE FROM vouchers
WHERE id = $1;

-- name: CreateCustomerVoucher :one
INSERT INTO customer_vouchers (customer_id, voucher_id, status, used_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomerVoucher :one
SELECT * FROM customer_vouchers
WHERE id = $1 LIMIT 1;

-- name: GetCustomerVouchers :many
SELECT cv.*, v.code, v.description, v.discount_type, v.discount_value, v.start_date, v.end_date
FROM customer_vouchers cv
JOIN vouchers v ON cv.voucher_id = v.id
WHERE cv.customer_id = $1
ORDER BY cv.id DESC
LIMIT $2 OFFSET $3;

-- name: GetCustomerVouchersByStatus :many
SELECT cv.*, v.code, v.description, v.discount_type, v.discount_value, v.start_date, v.end_date
FROM customer_vouchers cv
JOIN vouchers v ON cv.voucher_id = v.id
WHERE cv.customer_id = $1 AND cv.status = $2
ORDER BY cv.id DESC
LIMIT $3 OFFSET $4;

-- name: GetAllCustomerVouchers :many
SELECT cv.*, v.code, v.description, v.discount_type, v.discount_value, v.start_date, v.end_date
FROM customer_vouchers cv
JOIN vouchers v ON cv.voucher_id = v.id
ORDER BY cv.id DESC
LIMIT $1 OFFSET $2;

-- name: UseCustomerVoucher :one
UPDATE customer_vouchers
SET status = 'used', used_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCustomerVoucherStatus :one
UPDATE customer_vouchers
SET status = $2, used_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteCustomerVoucher :exec
DELETE FROM customer_vouchers
WHERE id = $1;

-- name: GetAvailableVouchersForCustomer :many
SELECT v.*
FROM vouchers v
LEFT JOIN customer_vouchers cv ON v.id = cv.voucher_id AND cv.customer_id = $1
WHERE v.start_date <= CURRENT_DATE 
  AND v.end_date >= CURRENT_DATE
  AND (cv.id IS NULL OR cv.status = 'unused')
  AND (v.usage_limit IS NULL OR 
       (SELECT COUNT(*) FROM customer_vouchers WHERE voucher_id = v.id AND status = 'used') < v.usage_limit)
ORDER BY v.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUsageRecordsByVoucherId :many
SELECT u.*
FROM usage_records U
WHERE u.voucher_id = $1
LIMIT $2 OFFSET $3;

-- name: CountUsageRecordsByVoucherId :one
SELECT COUNT(*)
FROM usage_records
WHERE voucher_id = $1;
