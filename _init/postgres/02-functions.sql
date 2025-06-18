-- Function definitions

-- Function to update account balance with audit trail
CREATE OR REPLACE FUNCTION core.update_account_balance(
    p_account_id UUID,
    p_amount DECIMAL(19,4),
    p_operation VARCHAR(50),
    p_transaction_id UUID DEFAULT NULL,
    p_created_by VARCHAR(255) DEFAULT 'system'
) RETURNS DECIMAL(19,4)
LANGUAGE plpgsql
AS $$
DECLARE
    v_old_balance DECIMAL(19,4);
    v_new_balance DECIMAL(19,4);
    v_account_version INTEGER;
BEGIN
    -- Lock the account row for update
    SELECT balance, version INTO v_old_balance, v_account_version
    FROM core.accounts 
    WHERE id = p_account_id 
    FOR UPDATE;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Account not found: %', p_account_id;
    END IF;
    
    -- Calculate new balance
    v_new_balance := v_old_balance + p_amount;
    
    -- Check for negative balance
    IF v_new_balance < 0 THEN
        RAISE EXCEPTION 'Insufficient funds. Current balance: %, Requested amount: %', v_old_balance, p_amount;
    END IF;
    
    -- Update account balance and version
    UPDATE core.accounts 
    SET balance = v_new_balance,
        version = version + 1,
        updated_at = NOW()
    WHERE id = p_account_id AND version = v_account_version;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Account was modified by another transaction. Please retry.';
    END IF;
    
    -- Insert audit record
    INSERT INTO core.account_balance_history (
        account_id,
        transaction_id,
        old_balance,
        new_balance,
        balance_change,
        operation,
        created_by
    ) VALUES (
        p_account_id,
        p_transaction_id,
        v_old_balance,
        v_new_balance,
        p_amount,
        p_operation,
        p_created_by
    );
    
    RETURN v_new_balance;
END;
$$;

-- Function to check account balance and status
CREATE OR REPLACE FUNCTION core.check_account_balance(
    p_account_id UUID,
    p_required_amount DECIMAL(19,4) DEFAULT NULL
) RETURNS TABLE(
    account_id UUID,
    account_number VARCHAR(20),
    account_name VARCHAR(255),
    balance DECIMAL(19,4),
    currency core.currency_code,
    status core.account_status,
    sufficient_funds BOOLEAN
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.id,
        a.account_number,
        a.account_name,
        a.balance,
        a.currency,
        a.status,
        CASE 
            WHEN p_required_amount IS NULL THEN TRUE
            WHEN a.balance >= p_required_amount THEN TRUE
            ELSE FALSE
        END AS sufficient_funds
    FROM core.accounts a
    WHERE a.id = p_account_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Account not found: %', p_account_id;
    END IF;
END;
$$;

-- Function to get account by account number
CREATE OR REPLACE FUNCTION core.get_account_by_number(
    p_account_number VARCHAR(20)
) RETURNS TABLE(
    account_id UUID,
    account_number VARCHAR(20),
    account_name VARCHAR(255),
    balance DECIMAL(19,4),
    currency core.currency_code,
    status core.account_status
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.id,
        a.account_number,
        a.account_name,
        a.balance,
        a.currency,
        a.status
    FROM core.accounts a
    WHERE a.account_number = p_account_number;
END;
$$;

-- Function to create a transaction record
CREATE OR REPLACE FUNCTION core.create_transaction(
    p_account_id UUID,
    p_transaction_type core.transaction_type,
    p_amount DECIMAL(19,4),
    p_currency core.currency_code,
    p_description TEXT DEFAULT NULL,
    p_reference_id VARCHAR(255) DEFAULT NULL,
    p_idempotency_key VARCHAR(255) DEFAULT NULL,
    p_metadata JSONB DEFAULT NULL
) RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_transaction_id UUID;
BEGIN
    -- Check for existing transaction with same idempotency key
    IF p_idempotency_key IS NOT NULL THEN
        SELECT id INTO v_transaction_id
        FROM core.transactions
        WHERE idempotency_key = p_idempotency_key;
        
        IF FOUND THEN
            RETURN v_transaction_id;
        END IF;
    END IF;
    
    -- Create new transaction
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
        p_account_id,
        p_transaction_type,
        p_amount,
        p_currency,
        p_description,
        p_reference_id,
        p_idempotency_key,
        p_metadata,
        'pending'
    ) RETURNING id INTO v_transaction_id;
    
    RETURN v_transaction_id;
END;
$$;

-- Function to complete a transaction and update balance
CREATE OR REPLACE FUNCTION core.complete_transaction(
    p_transaction_id UUID
) RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
DECLARE
    v_transaction RECORD;
    v_balance_change DECIMAL(19,4);
    v_new_balance DECIMAL(19,4);
BEGIN
    -- Get transaction details
    SELECT * INTO v_transaction
    FROM core.transactions
    WHERE id = p_transaction_id AND status = 'pending'
    FOR UPDATE;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Transaction not found or already processed: %', p_transaction_id;
    END IF;
    
    -- Calculate balance change (negative for debit, positive for credit)
    v_balance_change := CASE 
        WHEN v_transaction.transaction_type = 'debit' THEN -v_transaction.amount
        WHEN v_transaction.transaction_type = 'credit' THEN v_transaction.amount
    END;
    
    -- Update account balance
    v_new_balance := core.update_account_balance(
        v_transaction.account_id,
        v_balance_change,
        v_transaction.transaction_type::VARCHAR,
        p_transaction_id,
        'transaction_service'
    );
    
    -- Mark transaction as completed
    UPDATE core.transactions
    SET status = 'completed',
        completed_at = NOW(),
        updated_at = NOW()
    WHERE id = p_transaction_id;
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        -- Mark transaction as failed
        UPDATE core.transactions
        SET status = 'failed',
            updated_at = NOW()
        WHERE id = p_transaction_id;
        
        RAISE;
END;
$$;

-- Function to cancel a pending transaction
CREATE OR REPLACE FUNCTION core.cancel_transaction(
    p_transaction_id UUID,
    p_reason TEXT DEFAULT 'Cancelled by user'
) RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE core.transactions
    SET status = 'cancelled',
        updated_at = NOW(),
        metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('cancellation_reason', p_reason)
    WHERE id = p_transaction_id AND status = 'pending';
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Transaction not found or cannot be cancelled: %', p_transaction_id;
    END IF;
    
    RETURN TRUE;
END;
$$;

-- Function to get transaction history for an account
CREATE OR REPLACE FUNCTION core.get_account_transactions(
    p_account_id UUID,
    p_limit INTEGER DEFAULT 50,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE(
    transaction_id UUID,
    transaction_type core.transaction_type,
    amount DECIMAL(19,4),
    currency core.currency_code,
    description TEXT,
    reference_id VARCHAR(255),
    status core.transaction_status,
    created_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.id,
        t.transaction_type,
        t.amount,
        t.currency,
        t.description,
        t.reference_id,
        t.status,
        t.created_at,
        t.completed_at
    FROM core.transactions t
    WHERE t.account_id = p_account_id
    ORDER BY t.created_at DESC
    LIMIT p_limit OFFSET p_offset;
END;
$$;

