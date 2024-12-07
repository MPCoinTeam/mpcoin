-- name: CreateTransaction :one
INSERT INTO transactions (id, wallet_id , chain_id, to_address, amount, token_id, gas_price, gas_limit, nonce, from_address, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: InsertSetteledTransaction :one
INSERT INTO transactions (id, wallet_id , chain_id, to_address, amount, token_id, gas_price, gas_limit, nonce, status, tx_hash, from_address)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 LIMIT 1;

-- name: GetTransactionsByWalletID :many
SELECT * FROM transactions
WHERE wallet_id = $1;

-- name: UpdateTransaction :one
UPDATE transactions 
SET (status, tx_hash, gas_price, gas_limit, nonce) = ($2, $3, $4, $5, $6)
WHERE id = $1
RETURNING *;

-- name: GetPaginatedTransactions :many
SELECT transactions.*
FROM users
JOIN wallets ON users.id = wallets.user_id
JOIN transactions ON wallets.address = transactions.from_address OR wallets.address = transactions.to_address
WHERE users.id = $1 AND transactions.token_id = $2
ORDER BY transactions.created_at DESC
LIMIT $3
OFFSET $4;

-- name: GetPaginatedAllTokenTransactions :many
SELECT transactions.*
FROM users
JOIN wallets ON users.id = wallets.user_id
JOIN transactions ON wallets.address = transactions.from_address OR wallets.address = transactions.to_address
WHERE users.id = $1
ORDER BY transactions.created_at DESC
LIMIT $2
OFFSET $3;


-- name: DeleteTransaction :one
DELETE FROM transactions
WHERE tx_hash = $1
RETURNING *;
