-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET email = $2, password_hash = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetUserWithWallet :one
SELECT u.id, u.email, w.id, w.address
FROM users u
LEFT JOIN wallets w ON u.id = w.user_id
WHERE u.id = $1 LIMIT 1;

-- name: GetUserByAddress :one
SELECT u.id, u.email, u.name, w.id
FROM users u
LEFT JOIN wallets w ON u.id = w.user_id
WHERE w.address = $1 LIMIT 1;
