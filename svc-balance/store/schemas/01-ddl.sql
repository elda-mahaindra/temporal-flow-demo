-- Schema definitions
CREATE SCHEMA IF NOT EXISTS "core";

-- Type definitions
CREATE TYPE core.account_status AS ENUM ('active', 'inactive', 'suspended', 'closed');
CREATE TYPE core.currency_code AS ENUM ('USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY', 'SGD', 'HKD');
CREATE TYPE core.transaction_type AS ENUM ('debit', 'credit');
CREATE TYPE core.transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled');
CREATE TYPE core.transfer_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'cancelled');
CREATE TYPE core.compensation_type AS ENUM ('debit_reversal', 'credit_reversal', 'manual_adjustment');
CREATE TYPE core.compensation_status AS ENUM ('pending', 'completed', 'failed', 'timeout', 'manual_required');

-- Table definitions

-- Accounts table for balance service
CREATE TABLE core.accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(20) NOT NULL UNIQUE,
    account_name VARCHAR(255) NOT NULL,
    balance DECIMAL(19,4) NOT NULL DEFAULT 0.0000 CHECK (balance >= 0),
    currency core.currency_code NOT NULL DEFAULT 'USD',
    status core.account_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1 -- For optimistic locking
);

-- Transactions table for transaction service
CREATE TABLE core.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES core.accounts(id),
    transaction_type core.transaction_type NOT NULL,
    amount DECIMAL(19,4) NOT NULL CHECK (amount > 0),
    currency core.currency_code NOT NULL,
    description TEXT,
    reference_id VARCHAR(255), -- External reference (e.g., transfer ID)
    status core.transaction_status NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(255) UNIQUE, -- For idempotent operations
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB -- Additional transaction metadata
);

-- Transfers table for tracking complete transfer operations
CREATE TABLE core.transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id VARCHAR(255) NOT NULL UNIQUE, -- External transfer identifier
    from_account_id UUID NOT NULL REFERENCES core.accounts(id),
    to_account_id UUID NOT NULL REFERENCES core.accounts(id),
    amount DECIMAL(19,4) NOT NULL CHECK (amount > 0),
    currency core.currency_code NOT NULL,
    description TEXT,
    status core.transfer_status NOT NULL DEFAULT 'pending',
    debit_transaction_id UUID REFERENCES core.transactions(id),
    credit_transaction_id UUID REFERENCES core.transactions(id),
    workflow_id VARCHAR(255), -- Temporal workflow ID
    run_id VARCHAR(255), -- Temporal run ID
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    metadata JSONB -- Additional transfer metadata
);

-- Audit log table for tracking all balance changes
CREATE TABLE core.account_balance_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES core.accounts(id),
    transaction_id UUID REFERENCES core.transactions(id),
    old_balance DECIMAL(19,4) NOT NULL,
    new_balance DECIMAL(19,4) NOT NULL,
    balance_change DECIMAL(19,4) NOT NULL,
    operation VARCHAR(50) NOT NULL, -- 'debit', 'credit', 'adjustment'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) -- Service or user that made the change
);

-- Compensation audit trail for tracking compensation operations
CREATE TABLE core.compensation_audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id VARCHAR(255) NOT NULL,
    run_id VARCHAR(255) NOT NULL,
    transfer_id VARCHAR(255), -- Reference to the transfer being compensated
    original_transaction_id UUID REFERENCES core.transactions(id),
    compensation_transaction_id UUID REFERENCES core.transactions(id),
    compensation_reason TEXT NOT NULL,
    compensation_type core.compensation_type NOT NULL,
    compensation_status core.compensation_status NOT NULL DEFAULT 'pending',
    compensation_attempts INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    timeout_duration_ms INTEGER, -- Timeout that occurred (if any)
    metadata JSONB -- Additional compensation context
);

-- Index definitions

-- Accounts indexes
CREATE INDEX idx_accounts_account_number ON core.accounts(account_number);
CREATE INDEX idx_accounts_status ON core.accounts(status);
CREATE INDEX idx_accounts_currency ON core.accounts(currency);
CREATE INDEX idx_accounts_created_at ON core.accounts(created_at);

-- Transactions indexes
CREATE INDEX idx_transactions_account_id ON core.transactions(account_id);
CREATE INDEX idx_transactions_type ON core.transactions(transaction_type);
CREATE INDEX idx_transactions_status ON core.transactions(status);
CREATE INDEX idx_transactions_reference_id ON core.transactions(reference_id);
CREATE INDEX idx_transactions_idempotency_key ON core.transactions(idempotency_key);
CREATE INDEX idx_transactions_created_at ON core.transactions(created_at);
CREATE INDEX idx_transactions_account_created ON core.transactions(account_id, created_at);

-- Transfers indexes
CREATE INDEX idx_transfers_transfer_id ON core.transfers(transfer_id);
CREATE INDEX idx_transfers_from_account ON core.transfers(from_account_id);
CREATE INDEX idx_transfers_to_account ON core.transfers(to_account_id);
CREATE INDEX idx_transfers_status ON core.transfers(status);
CREATE INDEX idx_transfers_workflow_id ON core.transfers(workflow_id);
CREATE INDEX idx_transfers_created_at ON core.transfers(created_at);

-- Balance history indexes
CREATE INDEX idx_balance_history_account_id ON core.account_balance_history(account_id);
CREATE INDEX idx_balance_history_transaction_id ON core.account_balance_history(transaction_id);
CREATE INDEX idx_balance_history_created_at ON core.account_balance_history(created_at);
CREATE INDEX idx_balance_history_account_created ON core.account_balance_history(account_id, created_at);

-- Compensation audit trail indexes
CREATE INDEX idx_compensation_workflow_id ON core.compensation_audit_trail(workflow_id);
CREATE INDEX idx_compensation_run_id ON core.compensation_audit_trail(run_id);
CREATE INDEX idx_compensation_transfer_id ON core.compensation_audit_trail(transfer_id);
CREATE INDEX idx_compensation_original_tx ON core.compensation_audit_trail(original_transaction_id);
CREATE INDEX idx_compensation_status ON core.compensation_audit_trail(compensation_status);
CREATE INDEX idx_compensation_type ON core.compensation_audit_trail(compensation_type);
CREATE INDEX idx_compensation_created_at ON core.compensation_audit_trail(created_at);
CREATE INDEX idx_compensation_workflow_status ON core.compensation_audit_trail(workflow_id, compensation_status);

-- Comment definitions
COMMENT ON SCHEMA core IS 'Core banking schema for temporal-flow-demo';

COMMENT ON TABLE core.accounts IS 'Account information and balances';
COMMENT ON COLUMN core.accounts.version IS 'Version for optimistic locking';
COMMENT ON COLUMN core.accounts.balance IS 'Current account balance with 4 decimal precision';

COMMENT ON TABLE core.transactions IS 'Individual debit/credit transactions';
COMMENT ON COLUMN core.transactions.idempotency_key IS 'Ensures idempotent transaction processing';
COMMENT ON COLUMN core.transactions.metadata IS 'Additional transaction context and data';

COMMENT ON TABLE core.transfers IS 'Complete money transfer operations';
COMMENT ON COLUMN core.transfers.workflow_id IS 'Temporal workflow ID for tracking';
COMMENT ON COLUMN core.transfers.run_id IS 'Temporal run ID for tracking';

COMMENT ON TABLE core.account_balance_history IS 'Audit trail for all balance changes';
COMMENT ON TABLE core.compensation_audit_trail IS 'Audit trail for compensation operations in Temporal workflows';
COMMENT ON COLUMN core.compensation_audit_trail.workflow_id IS 'Temporal workflow ID for compensation tracking';
COMMENT ON COLUMN core.compensation_audit_trail.compensation_attempts IS 'Number of attempts made for this compensation';
COMMENT ON COLUMN core.compensation_audit_trail.timeout_duration_ms IS 'Timeout duration if compensation timed out';

-- Reference definitions
ALTER TABLE core.transfers ADD CONSTRAINT fk_transfers_from_account 
    FOREIGN KEY (from_account_id) REFERENCES core.accounts(id) ON DELETE RESTRICT;
    
ALTER TABLE core.transfers ADD CONSTRAINT fk_transfers_to_account 
    FOREIGN KEY (to_account_id) REFERENCES core.accounts(id) ON DELETE RESTRICT;

ALTER TABLE core.transfers ADD CONSTRAINT chk_transfers_different_accounts 
    CHECK (from_account_id != to_account_id);

ALTER TABLE core.transactions ADD CONSTRAINT fk_transactions_account 
    FOREIGN KEY (account_id) REFERENCES core.accounts(id) ON DELETE RESTRICT;

