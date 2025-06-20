-- Integration Test Data
-- This file contains test-specific accounts and transactions for integration testing

-- Test accounts with specific balances for different test scenarios
INSERT INTO core.accounts (id, account_number, account_holder_name, account_type, currency, balance, status) VALUES
    -- Test accounts for successful transfers
    ('111e8400-e29b-41d4-a716-446655440001', '1000000001', 'Integration Test Alice', 'checking', 'USD', 1000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440002', '1000000002', 'Integration Test Bob', 'checking', 'USD', 500.0000, 'active'),
    
    -- Test accounts for insufficient funds scenarios
    ('111e8400-e29b-41d4-a716-446655440003', '1000000003', 'Low Balance Charlie', 'checking', 'USD', 10.0000, 'active'),
    
    -- Test accounts for account validation failures
    ('111e8400-e29b-41d4-a716-446655440004', '1000000004', 'Frozen Account Dana', 'checking', 'USD', 1000.0000, 'frozen'),
    ('111e8400-e29b-41d4-a716-446655440005', '1000000005', 'Closed Account Eve', 'checking', 'USD', 0.0000, 'closed'),
    
    -- Test accounts for failure simulation (matches failure simulator account IDs)
    ('111e8400-e29b-41d4-a716-446655440006', '111111111111', 'Failure Test Account', 'checking', 'USD', 1000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440007', '222222222222', 'Compensation Test Account', 'checking', 'USD', 1000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440008', '123456789012', 'Timeout Test Account', 'checking', 'USD', 1000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440009', '999999999999', 'Panic Test Account', 'checking', 'USD', 1000.0000, 'active'),
    
    -- Test accounts for currency testing
    ('111e8400-e29b-41d4-a716-446655440010', '1000000010', 'EUR Test Account', 'checking', 'EUR', 1000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440011', '1000000011', 'GBP Test Account', 'checking', 'GBP', 1000.0000, 'active'),
    
    -- High balance accounts for large transfer testing
    ('111e8400-e29b-41d4-a716-446655440012', '1000000012', 'High Balance Frank', 'checking', 'USD', 100000.0000, 'active'),
    ('111e8400-e29b-41d4-a716-446655440013', '1000000013', 'High Balance Grace', 'checking', 'USD', 50000.0000, 'active');

-- Initial transaction history for test accounts
INSERT INTO core.transactions (id, account_id, transaction_type, amount, currency, description, status, completed_at) VALUES
    -- Setup transactions for test accounts
    ('771e8400-e29b-41d4-a716-446655440001', '111e8400-e29b-41d4-a716-446655440001', 'credit', 1000.0000, 'USD', 'Test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440002', '111e8400-e29b-41d4-a716-446655440002', 'credit', 500.0000, 'USD', 'Test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440003', '111e8400-e29b-41d4-a716-446655440003', 'credit', 10.0000, 'USD', 'Low balance test setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440004', '111e8400-e29b-41d4-a716-446655440004', 'credit', 1000.0000, 'USD', 'Frozen account setup', 'completed', NOW() - INTERVAL '2 days'),
    ('771e8400-e29b-41d4-a716-446655440006', '111e8400-e29b-41d4-a716-446655440006', 'credit', 1000.0000, 'USD', 'Failure test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440007', '111e8400-e29b-41d4-a716-446655440007', 'credit', 1000.0000, 'USD', 'Compensation test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440008', '111e8400-e29b-41d4-a716-446655440008', 'credit', 1000.0000, 'USD', 'Timeout test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440009', '111e8400-e29b-41d4-a716-446655440009', 'credit', 1000.0000, 'USD', 'Panic test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440010', '111e8400-e29b-41d4-a716-446655440010', 'credit', 1000.0000, 'EUR', 'EUR test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440011', '111e8400-e29b-41d4-a716-446655440011', 'credit', 1000.0000, 'GBP', 'GBP test account setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440012', '111e8400-e29b-41d4-a716-446655440012', 'credit', 100000.0000, 'USD', 'High balance test setup', 'completed', NOW() - INTERVAL '1 day'),
    ('771e8400-e29b-41d4-a716-446655440013', '111e8400-e29b-41d4-a716-446655440013', 'credit', 50000.0000, 'USD', 'High balance test setup', 'completed', NOW() - INTERVAL '1 day');

-- Balance history for test accounts
INSERT INTO core.account_balance_history (account_id, transaction_id, old_balance, new_balance, balance_change, operation, created_by) VALUES
    ('111e8400-e29b-41d4-a716-446655440001', '771e8400-e29b-41d4-a716-446655440001', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440002', '771e8400-e29b-41d4-a716-446655440002', 0.0000, 500.0000, 500.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440003', '771e8400-e29b-41d4-a716-446655440003', 0.0000, 10.0000, 10.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440004', '771e8400-e29b-41d4-a716-446655440004', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440006', '771e8400-e29b-41d4-a716-446655440006', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440007', '771e8400-e29b-41d4-a716-446655440007', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440008', '771e8400-e29b-41d4-a716-446655440008', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440009', '771e8400-e29b-41d4-a716-446655440009', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440010', '771e8400-e29b-41d4-a716-446655440010', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440011', '771e8400-e29b-41d4-a716-446655440011', 0.0000, 1000.0000, 1000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440012', '771e8400-e29b-41d4-a716-446655440012', 0.0000, 100000.0000, 100000.0000, 'credit', 'integration_test_setup'),
    ('111e8400-e29b-41d4-a716-446655440013', '771e8400-e29b-41d4-a716-446655440013', 0.0000, 50000.0000, 50000.0000, 'credit', 'integration_test_setup'); 