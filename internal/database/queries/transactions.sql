-- internal/database/queries/transactions.sql

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, amount, description, category, date)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY date DESC
LIMIT 100;

-- name: GetTotalSpending :one
SELECT COALESCE(SUM(amount), 0)::numeric
FROM transactions
WHERE user_id = $1;
