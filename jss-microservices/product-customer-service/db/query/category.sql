-- name: CreateProductCategory :one
INSERT INTO product_categories (name)
VALUES ($1)
RETURNING *;

-- name: GetProductCategoryByID :one
SELECT * FROM product_categories
WHERE id = $1
LIMIT 1;

-- name: GetProductCategoryByName :one
SELECT * FROM product_categories
WHERE name = $1
LIMIT 1;

-- name: ListProductCategories :many
SELECT * FROM product_categories
ORDER BY id;

-- name: UpdateProductCategory :one
UPDATE product_categories
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteProductCategory :exec
DELETE FROM product_categories
WHERE id = $1;
