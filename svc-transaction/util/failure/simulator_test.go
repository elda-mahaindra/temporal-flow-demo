package failure

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSimulator_TransactionFailures(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	simulator := NewSimulator(logger)

	tests := []struct {
		name        string
		rules       []Rule
		operation   string
		accountID   string
		expectError bool
	}{
		{
			name: "should_fail_debit_operation",
			rules: []Rule{
				{
					Name:        "debit_failure",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0, // Always fail
					Operations:  []string{"DebitAccount"},
					Accounts:    []string{"*"},
					Message:     "Test debit failure",
				},
			},
			operation:   "DebitAccount",
			accountID:   "123456789012",
			expectError: true,
		},
		{
			name: "should_fail_credit_operation",
			rules: []Rule{
				{
					Name:        "credit_failure",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0, // Always fail
					Operations:  []string{"CreditAccount"},
					Accounts:    []string{"*"},
					Message:     "Test credit failure",
				},
			},
			operation:   "CreditAccount",
			accountID:   "123456789012",
			expectError: true,
		},
		{
			name: "should_fail_compensation_operation",
			rules: []Rule{
				{
					Name:        "compensation_failure",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0, // Always fail
					Operations:  []string{"CompensateDebit"},
					Accounts:    []string{"*"},
					Message:     "Test compensation failure",
				},
			},
			operation:   "CompensateDebit",
			accountID:   "123456789012",
			expectError: true,
		},
		{
			name: "should_not_fail_when_operation_does_not_match",
			rules: []Rule{
				{
					Name:        "different_operation",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0,
					Operations:  []string{"DebitAccount"},
					Accounts:    []string{"*"},
				},
			},
			operation:   "CreditAccount", // Different operation
			accountID:   "123456789012",
			expectError: false,
		},
		{
			name: "should_handle_timeout_simulation",
			rules: []Rule{
				{
					Name:        "transaction_timeout",
					Enabled:     true,
					Type:        "timeout",
					Probability: 1.0,
					Operations:  []string{"DebitAccount"},
					Accounts:    []string{"*"},
					TimeoutMs:   50, // Very short timeout for test
				},
			},
			operation:   "DebitAccount",
			accountID:   "123456789012",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset simulator for each test
			simulator.Reset()

			ctx := context.Background()
			err := simulator.SimulateFailure(ctx, tt.operation, tt.accountID, tt.rules)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "DEMO_TRANSACTION")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSimulator_TransactionSlowOperation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	simulator := NewSimulator(logger)

	rules := []Rule{
		{
			Name:        "slow_credit",
			Enabled:     true,
			Type:        "slow",
			Probability: 1.0,
			Operations:  []string{"CreditAccount"},
			Accounts:    []string{"*"},
			DelayMs:     100, // 100ms delay
		},
	}

	ctx := context.Background()
	start := time.Now()

	err := simulator.SimulateFailure(ctx, "CreditAccount", "123456789012", rules)

	duration := time.Since(start)

	// Should not return an error for slow type
	assert.NoError(t, err)

	// Should take at least 100ms
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
}

func TestSimulator_TransactionStatsTracking(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	simulator := NewSimulator(logger)

	// Initial stats
	stats := simulator.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "occurrences")
	assert.Contains(t, stats, "uptime_ms")

	// Trigger multiple transaction failures
	rules := []Rule{
		{
			Name:        "debit_test_rule",
			Enabled:     true,
			Type:        "error",
			Probability: 1.0,
			Operations:  []string{"DebitAccount"},
			Accounts:    []string{"*"},
		},
		{
			Name:        "credit_test_rule",
			Enabled:     true,
			Type:        "error",
			Probability: 1.0,
			Operations:  []string{"CreditAccount"},
			Accounts:    []string{"*"},
		},
	}

	ctx := context.Background()

	// Trigger debit failure
	simulator.SimulateFailure(ctx, "DebitAccount", "123456789012", rules)

	// Trigger credit failure
	simulator.SimulateFailure(ctx, "CreditAccount", "987654321098", rules)

	// Check stats updated
	stats = simulator.GetStats()
	occurrences := stats["occurrences"].(map[string]int)
	assert.Equal(t, 1, occurrences["debit_test_rule"])
	assert.Equal(t, 1, occurrences["credit_test_rule"])
}
