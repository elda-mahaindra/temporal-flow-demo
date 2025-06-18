-- name: CreateTransaction :one
INSERT INTO core.transactions (
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    idempotency_key,
    metadata,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, 'pending'
) RETURNING id, created_at;

-- name: GetTransactionByID :one
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    idempotency_key,
    status,
    created_at,
    updated_at,
    completed_at,
    metadata
FROM core.transactions
WHERE id = $1;

-- name: GetTransactionByIdempotencyKey :one
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    idempotency_key,
    status,
    created_at,
    updated_at,
    completed_at,
    metadata
FROM core.transactions
WHERE idempotency_key = $1;

-- name: UpdateTransactionStatus :one
UPDATE core.transactions
SET 
    status = $2,
    updated_at = NOW(),
    completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING id, status, updated_at, completed_at;

-- name: CompleteTransaction :one
UPDATE core.transactions
SET 
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING id, status, completed_at;

-- name: FailTransaction :one
UPDATE core.transactions
SET 
    status = 'failed',
    updated_at = NOW(),
    metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('failure_reason', $2)
WHERE id = $1 AND status = 'pending'
RETURNING id, status, updated_at;

-- name: CancelTransaction :one
UPDATE core.transactions
SET 
    status = 'cancelled',
    updated_at = NOW(),
    metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('cancellation_reason', $2)
WHERE id = $1 AND status = 'pending'
RETURNING id, status, updated_at;

-- name: GetTransactionsByAccount :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionsByReference :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE reference_id = $1
ORDER BY created_at ASC;

-- name: GetTransactionsByStatus :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetPendingTransactions :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    idempotency_key,
    status,
    created_at,
    metadata
FROM core.transactions
WHERE status = 'pending'
    AND created_at < NOW() - INTERVAL '5 minutes'
ORDER BY created_at ASC
LIMIT $1;

-- name: GetTransactionsByAccountAndType :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE account_id = $1 AND transaction_type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTransactionsByDateRange :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTransactionSummaryByAccount :one
SELECT 
    account_id,
    COUNT(*) AS total_transactions,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_transactions,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_transactions,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed_transactions,
    COALESCE(SUM(CASE WHEN transaction_type = 'debit' AND status = 'completed' THEN amount ELSE 0 END), 0) AS total_debits,
    COALESCE(SUM(CASE WHEN transaction_type = 'credit' AND status = 'completed' THEN amount ELSE 0 END), 0) AS total_credits
FROM core.transactions
WHERE account_id = $1;

-- name: GetRecentTransactionsByAccount :many
SELECT 
    id,
    account_id,
    transaction_type,
    amount,
    currency,
    description,
    reference_id,
    status,
    created_at,
    completed_at
FROM core.transactions
WHERE account_id = $1
    AND created_at >= NOW() - INTERVAL '30 days'
ORDER BY created_at DESC
LIMIT $2;

-- name: UpdateTransactionMetadata :one
UPDATE core.transactions
SET 
    metadata = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, metadata, updated_at;

-- Account-related queries for transaction service

-- name: GetAccountByID :one
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    created_at,
    updated_at,
    version
FROM core.accounts
WHERE id = $1;

-- name: GetAccountByAccountNumber :one
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    created_at,
    updated_at,
    version
FROM core.accounts
WHERE account_number = $1;

-- name: CheckAccountBalance :one
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    CASE 
        WHEN $2::DECIMAL IS NULL THEN TRUE
        WHEN balance >= $2::DECIMAL THEN TRUE
        ELSE FALSE
    END AS sufficient_funds
FROM core.accounts
WHERE id = $1;

-- name: CreateBalanceHistoryRecord :one
INSERT INTO core.account_balance_history (
    account_id,
    transaction_id,
    old_balance,
    new_balance,
    balance_change,
    operation,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, created_at;

-- name: GetAccountBalanceHistory :many
SELECT 
    id,
    account_id,
    transaction_id,
    old_balance,
    new_balance,
    balance_change,
    operation,
    created_at,
    created_by
FROM core.account_balance_history
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBalanceHistoryByTransaction :many
SELECT 
    id,
    account_id,
    transaction_id,
    old_balance,
    new_balance,
    balance_change,
    operation,
    created_at,
    created_by
FROM core.account_balance_history
WHERE transaction_id = $1
ORDER BY created_at ASC;

-- name: GetBalanceHistoryByDateRange :many
SELECT 
    id,
    account_id,
    transaction_id,
    old_balance,
    new_balance,
    balance_change,
    operation,
    created_at,
    created_by
FROM core.account_balance_history
WHERE account_id = $1
    AND created_at BETWEEN $2 AND $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;
