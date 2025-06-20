package failure

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSimulator_SimulateFailure(t *testing.T) {
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
			name:        "should_not_fail_when_no_rules",
			rules:       []Rule{},
			operation:   "CheckBalance",
			accountID:   "123456789012",
			expectError: false,
		},
		{
			name: "should_not_fail_when_rule_disabled",
			rules: []Rule{
				{
					Name:        "disabled_rule",
					Enabled:     false,
					Type:        "error",
					Probability: 1.0,
					Operations:  []string{"CheckBalance"},
					Accounts:    []string{"*"},
				},
			},
			operation:   "CheckBalance",
			accountID:   "123456789012",
			expectError: false,
		},
		{
			name: "should_fail_when_rule_matches",
			rules: []Rule{
				{
					Name:        "test_failure",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0, // Always fail
					Operations:  []string{"CheckBalance"},
					Accounts:    []string{"*"},
					Message:     "Test failure",
				},
			},
			operation:   "CheckBalance",
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
					Operations:  []string{"ValidateAccount"},
					Accounts:    []string{"*"},
				},
			},
			operation:   "CheckBalance",
			accountID:   "123456789012",
			expectError: false,
		},
		{
			name: "should_not_fail_when_account_does_not_match",
			rules: []Rule{
				{
					Name:        "specific_account",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0,
					Operations:  []string{"CheckBalance"},
					Accounts:    []string{"987654321098"},
				},
			},
			operation:   "CheckBalance",
			accountID:   "123456789012",
			expectError: false,
		},
		{
			name: "should_respect_max_count",
			rules: []Rule{
				{
					Name:        "limited_failures",
					Enabled:     true,
					Type:        "error",
					Probability: 1.0,
					Operations:  []string{"CheckBalance"},
					Accounts:    []string{"*"},
					MaxCount:    1,
				},
			},
			operation:   "CheckBalance",
			accountID:   "123456789012",
			expectError: true, // First call should fail
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSimulator_SlowFailure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	simulator := NewSimulator(logger)

	rules := []Rule{
		{
			Name:        "slow_operation",
			Enabled:     true,
			Type:        "slow",
			Probability: 1.0,
			Operations:  []string{"CheckBalance"},
			Accounts:    []string{"*"},
			DelayMs:     100, // 100ms delay
		},
	}

	ctx := context.Background()
	start := time.Now()

	err := simulator.SimulateFailure(ctx, "CheckBalance", "123456789012", rules)

	duration := time.Since(start)

	// Should not return an error for slow type
	assert.NoError(t, err)

	// Should take at least 100ms
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
}

func TestSimulator_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	simulator := NewSimulator(logger)

	// Initial stats
	stats := simulator.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "occurrences")
	assert.Contains(t, stats, "uptime_ms")

	// Trigger a failure
	rules := []Rule{
		{
			Name:        "test_rule",
			Enabled:     true,
			Type:        "error",
			Probability: 1.0,
			Operations:  []string{"CheckBalance"},
			Accounts:    []string{"*"},
		},
	}

	ctx := context.Background()
	simulator.SimulateFailure(ctx, "CheckBalance", "123456789012", rules)

	// Check stats updated
	stats = simulator.GetStats()
	occurrences := stats["occurrences"].(map[string]int)
	assert.Equal(t, 1, occurrences["test_rule"])
}

func TestSimulator_Reset(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	simulator := NewSimulator(logger)

	// Trigger a failure
	rules := []Rule{
		{
			Name:        "test_rule",
			Enabled:     true,
			Type:        "error",
			Probability: 1.0,
			Operations:  []string{"CheckBalance"},
			Accounts:    []string{"*"},
		},
	}

	ctx := context.Background()
	simulator.SimulateFailure(ctx, "CheckBalance", "123456789012", rules)

	// Verify failure was recorded
	stats := simulator.GetStats()
	occurrences := stats["occurrences"].(map[string]int)
	assert.Equal(t, 1, occurrences["test_rule"])

	// Reset
	simulator.Reset()

	// Verify stats were reset
	stats = simulator.GetStats()
	occurrences = stats["occurrences"].(map[string]int)
	assert.Equal(t, 0, occurrences["test_rule"])
}
