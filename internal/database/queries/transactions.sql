-- internal/database/queries/transactions.sql

-- name: CreateTransaction :one
INSERT INTO transactions (
  user_id, amount, description, category, type, currency, status,
  account, tags, notes, has_receipt, receipt_url, date
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1 AND user_id = $2;

-- name: UpdateTransaction :one
UPDATE transactions
SET
  amount = $2,
  description = $3,
  category = $4,
  type = $5,
  currency = $6,
  status = $7,
  account = $8,
  tags = $9,
  notes = $10,
  has_receipt = $11,
  receipt_url = $12,
  updated_at = NOW()
WHERE id = $1 AND user_id = $13
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1 AND user_id = $2;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactionsWithFilters :many
SELECT * FROM transactions
WHERE user_id = $1
  AND ($2::text IS NULL OR description ILIKE '%' || $2 || '%')
  AND ($3::date IS NULL OR date >= $3::date)
  AND ($4::date IS NULL OR date <= $4::date)
  AND ($5::text[] IS NULL OR category = ANY($5::text[]))
  AND ($6::text IS NULL OR type = $6)
  AND ($7::numeric IS NULL OR amount >= $7::numeric)
  AND ($8::numeric IS NULL OR amount <= $8::numeric)
  AND ($9::text[] IS NULL OR tags && $9::text[])
ORDER BY date DESC
LIMIT $10 OFFSET $11;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1
  AND ($2::text IS NULL OR description ILIKE '%' || $2 || '%')
  AND ($3::date IS NULL OR date >= $3::date)
  AND ($4::date IS NULL OR date <= $4::date)
  AND ($5::text[] IS NULL OR category = ANY($5::text[]))
  AND ($6::text IS NULL OR type = $6)
  AND ($7::numeric IS NULL OR amount >= $7::numeric)
  AND ($8::numeric IS NULL OR amount <= $8::numeric)
  AND ($9::text[] IS NULL OR tags && $9::text[]);

-- name: GetTotalSpending :one
SELECT COALESCE(SUM(amount), 0)::numeric
FROM transactions
WHERE user_id = $1 AND type = 'expense';

-- name: GetTotalIncome :one
SELECT COALESCE(SUM(amount), 0)::numeric
FROM transactions
WHERE user_id = $1 AND type = 'income';

-- name: GetCategories :many
SELECT DISTINCT category as name, category as id
FROM transactions
WHERE user_id = $1
ORDER BY category;

-- name: GetTags :many
SELECT DISTINCT unnest(tags) as name, unnest(tags) as id
FROM transactions
WHERE user_id = $1 AND array_length(tags, 1) > 0
ORDER BY name;
