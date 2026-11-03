-- name: CreateCustomer :one
INSERT INTO customers (
    name, phone, email, address, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
)
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM customers
WHERE id = $1;

-- name: GetCustomerByPhone :one
SELECT * FROM customers
WHERE phone = $1;

-- name: ListCustomers :many
SELECT * FROM customers
ORDER BY id LIMIT $1 OFFSET $2;

-- name: UpdateCustomer :one
UPDATE customers
SET
    name = $2,
    phone = $3,
    email = $4,
    address = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;
