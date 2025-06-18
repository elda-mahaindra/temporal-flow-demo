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

-- name: GetAccountByNumber :one
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
        WHEN $2::decimal IS NULL THEN true
        WHEN balance >= $2::decimal THEN true
        ELSE false
    END AS sufficient_funds
FROM core.accounts
WHERE id = $1;

-- name: GetAccountsByStatus :many
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
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAccountsByCurrency :many
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
WHERE currency = $1
ORDER BY balance DESC
LIMIT $2 OFFSET $3;

-- name: GetAccountBalanceHistory :many
SELECT 
    h.id,
    h.account_id,
    h.transaction_id,
    h.old_balance,
    h.new_balance,
    h.balance_change,
    h.operation,
    h.created_at,
    h.created_by
FROM core.account_balance_history h
WHERE h.account_id = $1
ORDER BY h.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAccountSummary :one
SELECT 
    a.id,
    a.account_number,
    a.account_name,
    a.balance,
    a.currency,
    a.status,
    a.created_at,
    a.updated_at,
    COUNT(t.id) AS transaction_count,
    COALESCE(SUM(CASE WHEN t.transaction_type = 'debit' THEN t.amount ELSE 0 END), 0) AS total_debits,
    COALESCE(SUM(CASE WHEN t.transaction_type = 'credit' THEN t.amount ELSE 0 END), 0) AS total_credits
FROM core.accounts a
LEFT JOIN core.transactions t ON a.id = t.account_id AND t.status = 'completed'
WHERE a.id = $1
GROUP BY a.id, a.account_number, a.account_name, a.balance, a.currency, a.status, a.created_at, a.updated_at;

-- name: ValidateAccountForTransaction :one
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    CASE 
        WHEN status != 'active' THEN false
        WHEN $2::text = 'debit' AND balance < $3::decimal THEN false
        ELSE true
    END AS can_transact,
    CASE 
        WHEN status != 'active' THEN 'Account is not active'
        WHEN $2::text = 'debit' AND balance < $3::decimal THEN 'Insufficient funds'
        ELSE 'OK'
    END AS validation_message
FROM core.accounts
WHERE id = $1;

-- name: GetAccountsWithLowBalance :many
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    created_at,
    updated_at
FROM core.accounts
WHERE balance < $1 AND status = 'active'
ORDER BY balance ASC
LIMIT $2 OFFSET $3;

-- name: GetAccountsByBalanceRange :many
SELECT 
    id,
    account_number,
    account_name,
    balance,
    currency,
    status,
    created_at,
    updated_at
FROM core.accounts
WHERE balance BETWEEN $1 AND $2
ORDER BY balance DESC
LIMIT $3 OFFSET $4;
