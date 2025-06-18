-- Data initialization

-- Insert sample accounts for testing
INSERT INTO core.accounts (id, account_number, account_name, balance, currency, status) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'ACC001', 'John Doe Primary Account', 5000.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440002', 'ACC002', 'Jane Smith Savings Account', 10000.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440003', 'ACC003', 'Business Account - Tech Corp', 25000.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440004', 'ACC004', 'Alice Johnson EUR Account', 7500.0000, 'EUR', 'active'),
    ('550e8400-e29b-41d4-a716-446655440005', 'ACC005', 'Bob Wilson GBP Account', 3000.0000, 'GBP', 'active'),
    ('550e8400-e29b-41d4-a716-446655440006', 'ACC006', 'Test Account - Low Balance', 100.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440007', 'ACC007', 'Suspended Account', 1000.0000, 'USD', 'suspended'),
    ('550e8400-e29b-41d4-a716-446655440008', 'ACC008', 'Corporate Account - BigCorp', 50000.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440009', 'ACC009', 'Zero Balance Account', 0.0000, 'USD', 'active'),
    ('550e8400-e29b-41d4-a716-446655440010', 'ACC010', 'High Balance Account', 100000.0000, 'USD', 'active');

-- Insert some sample transaction history
INSERT INTO core.transactions (id, account_id, transaction_type, amount, currency, description, status, completed_at) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'credit', 5000.0000, 'USD', 'Initial deposit', 'completed', NOW() - INTERVAL '30 days'),
    ('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', 'credit', 10000.0000, 'USD', 'Initial deposit', 'completed', NOW() - INTERVAL '25 days'),
    ('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440003', 'credit', 25000.0000, 'USD', 'Business initial funding', 'completed', NOW() - INTERVAL '20 days'),
    ('660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'debit', 200.0000, 'USD', 'ATM withdrawal', 'completed', NOW() - INTERVAL '15 days'),
    ('660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', 'credit', 200.0000, 'USD', 'Salary deposit', 'completed', NOW() - INTERVAL '10 days'),
    ('660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440002', 'debit', 500.0000, 'USD', 'Online purchase', 'completed', NOW() - INTERVAL '8 days'),
    ('660e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440002', 'credit', 500.0000, 'USD', 'Refund', 'completed', NOW() - INTERVAL '5 days'),
    ('660e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440008', 'credit', 50000.0000, 'USD', 'Corporate funding', 'completed', NOW() - INTERVAL '3 days'),
    ('660e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440010', 'credit', 100000.0000, 'USD', 'Large deposit', 'completed', NOW() - INTERVAL '1 day');

-- Insert corresponding balance history records
INSERT INTO core.account_balance_history (account_id, transaction_id, old_balance, new_balance, balance_change, operation, created_by) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', 0.0000, 5000.0000, 5000.0000, 'credit', 'initial_setup'),
    ('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440002', 0.0000, 10000.0000, 10000.0000, 'credit', 'initial_setup'),
    ('550e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440003', 0.0000, 25000.0000, 25000.0000, 'credit', 'initial_setup'),
    ('550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440004', 5000.0000, 4800.0000, -200.0000, 'debit', 'transaction_service'),
    ('550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440005', 4800.0000, 5000.0000, 200.0000, 'credit', 'transaction_service'),
    ('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440006', 10000.0000, 9500.0000, -500.0000, 'debit', 'transaction_service'),
    ('550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440007', 9500.0000, 10000.0000, 500.0000, 'credit', 'transaction_service'),
    ('550e8400-e29b-41d4-a716-446655440008', '660e8400-e29b-41d4-a716-446655440008', 0.0000, 50000.0000, 50000.0000, 'credit', 'initial_setup'),
    ('550e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440009', 0.0000, 100000.0000, 100000.0000, 'credit', 'initial_setup');

-- Insert some sample transfer records for testing
INSERT INTO core.transfers (id, transfer_id, from_account_id, to_account_id, amount, currency, description, status, debit_transaction_id, credit_transaction_id, completed_at) VALUES
    ('770e8400-e29b-41d4-a716-446655440001', 'TXF-2024-001', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 500.0000, 'USD', 'Test transfer - completed', 'completed', '660e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440007', NOW() - INTERVAL '5 days'),
    ('770e8400-e29b-41d4-a716-446655440002', 'TXF-2024-002', '550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440008', 1000.0000, 'USD', 'Business to business transfer', 'pending', NULL, NULL, NULL),
    ('770e8400-e29b-41d4-a716-446655440003', 'TXF-2024-003', '550e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440001', 2000.0000, 'USD', 'Large account to personal', 'processing', NULL, NULL, NULL);

-- Create some indexes for better query performance on sample data
ANALYZE core.accounts;
ANALYZE core.transactions;
ANALYZE core.transfers;
ANALYZE core.account_balance_history;

-- Display summary of created data
DO $$
BEGIN
    RAISE NOTICE 'Database initialization completed successfully!';
    RAISE NOTICE 'Created % accounts', (SELECT COUNT(*) FROM core.accounts);
    RAISE NOTICE 'Created % transactions', (SELECT COUNT(*) FROM core.transactions);
    RAISE NOTICE 'Created % transfers', (SELECT COUNT(*) FROM core.transfers);
    RAISE NOTICE 'Created % balance history records', (SELECT COUNT(*) FROM core.account_balance_history);
    RAISE NOTICE 'Total balance across all accounts: %', (SELECT SUM(balance) FROM core.accounts);
END $$;

