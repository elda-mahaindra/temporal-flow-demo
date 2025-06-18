-- Trigger definitions

-- Generic function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION core.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic updated_at timestamp updates
CREATE TRIGGER trigger_accounts_updated_at
    BEFORE UPDATE ON core.accounts
    FOR EACH ROW
    EXECUTE FUNCTION core.update_updated_at_column();

CREATE TRIGGER trigger_transactions_updated_at
    BEFORE UPDATE ON core.transactions
    FOR EACH ROW
    EXECUTE FUNCTION core.update_updated_at_column();

CREATE TRIGGER trigger_transfers_updated_at
    BEFORE UPDATE ON core.transfers
    FOR EACH ROW
    EXECUTE FUNCTION core.update_updated_at_column();

-- Trigger to automatically set completed_at when transfer status changes to completed
CREATE OR REPLACE FUNCTION core.set_transfer_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    -- Set completed_at when status changes to completed
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        NEW.completed_at = NOW();
    END IF;
    
    -- Set failed_at when status changes to failed
    IF NEW.status = 'failed' AND OLD.status != 'failed' THEN
        NEW.failed_at = NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_transfers_completion
    BEFORE UPDATE ON core.transfers
    FOR EACH ROW
    EXECUTE FUNCTION core.set_transfer_completed_at();

-- Trigger to automatically set completed_at when transaction status changes to completed
CREATE OR REPLACE FUNCTION core.set_transaction_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    -- Set completed_at when status changes to completed
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        NEW.completed_at = NOW();
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_transactions_completion
    BEFORE UPDATE ON core.transactions
    FOR EACH ROW
    EXECUTE FUNCTION core.set_transaction_completed_at();

-- Trigger to validate account status before operations
CREATE OR REPLACE FUNCTION core.validate_account_operations()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if account is active for balance updates
    IF TG_OP = 'UPDATE' AND OLD.balance != NEW.balance THEN
        IF NEW.status NOT IN ('active') THEN
            RAISE EXCEPTION 'Cannot modify balance for account with status: %', NEW.status;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_accounts_validation
    BEFORE UPDATE ON core.accounts
    FOR EACH ROW
    EXECUTE FUNCTION core.validate_account_operations();

-- Trigger to validate transaction creation
CREATE OR REPLACE FUNCTION core.validate_transaction_creation()
RETURNS TRIGGER AS $$
DECLARE
    v_account_status core.account_status;
BEGIN
    -- Check if the account exists and is active
    SELECT status INTO v_account_status
    FROM core.accounts
    WHERE id = NEW.account_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Account not found: %', NEW.account_id;
    END IF;
    
    IF v_account_status NOT IN ('active') THEN
        RAISE EXCEPTION 'Cannot create transaction for account with status: %', v_account_status;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_transactions_validation
    BEFORE INSERT ON core.transactions
    FOR EACH ROW
    EXECUTE FUNCTION core.validate_transaction_creation();

