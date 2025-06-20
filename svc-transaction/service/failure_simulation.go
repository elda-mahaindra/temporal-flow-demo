package service

import (
	"context"

	"svc-transaction/util/failure"
)

// Learning-focused failure simulation rules for transaction operations
// These are hardcoded scenarios designed to demonstrate Temporal's compensation and transaction handling
var learningFailureRules = []failure.Rule{
	// Scenario 1: Random debit failures to demonstrate transaction rollbacks
	{
		Name:        "random_debit_failures",
		Enabled:     true,
		Type:        "error",
		Probability: 0.12, // 12% chance of failure
		Operations:  []string{"DebitAccount"},
		Accounts:    []string{"*"}, // All accounts
		Message:     "Random debit failure for Temporal compensation demonstration",
		MaxCount:    4, // Limit to 4 failures for learning session
	},

	// Scenario 2: Credit operation timeouts to trigger compensation
	{
		Name:        "credit_timeout_demo",
		Enabled:     true,
		Type:        "timeout",
		Probability: 0.08, // 8% chance of timeout
		Operations:  []string{"CreditAccount"},
		Accounts:    []string{"*"},
		TimeoutMs:   3000, // 3 second timeout
		Message:     "Credit timeout demonstrating Temporal compensation workflows",
		MaxCount:    3, // Limit occurrences
	},

	// Scenario 3: Slow credit operations to test workflow timing
	{
		Name:        "slow_credit_processing",
		Enabled:     true,
		Type:        "slow",
		Probability: 0.15, // 15% chance of slowness
		Operations:  []string{"CreditAccount"},
		Accounts:    []string{"*"},
		DelayMs:     2500, // 2.5 second delay
		MaxCount:    5,    // Allow more for timing tests
	},

	// Scenario 4: Problematic account that always fails debits
	{
		Name:        "problematic_debit_account",
		Enabled:     true,
		Type:        "error",
		Probability: 1.0, // Always fail for this account
		Operations:  []string{"DebitAccount"},
		Accounts:    []string{"111111111111"}, // Specific test account
		Message:     "Problematic account demonstrating debit failure and workflow termination",
		MaxCount:    3, // Limit to 3 failures
	},

	// Scenario 5: Credit failures for specific account to test compensation
	{
		Name:        "compensation_trigger_account",
		Enabled:     true,
		Type:        "error",
		Probability: 1.0, // Always fail credits for this account
		Operations:  []string{"CreditAccount"},
		Accounts:    []string{"222222222222"}, // Specific compensation test account
		Message:     "Credit failure to trigger compensation workflow",
		MaxCount:    2, // Limit to test compensation
	},

	// Scenario 6: Compensation operation failures (advanced scenario)
	{
		Name:        "compensation_failure_demo",
		Enabled:     true,
		Type:        "error",
		Probability: 0.3, // 30% chance when compensation runs
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"*"},
		Message:     "Compensation failure demonstrating advanced error handling",
		MaxCount:    2, // Very limited for safety
	},

	// Scenario 7: Intermittent transaction processing issues
	{
		Name:        "intermittent_transaction_issues",
		Enabled:     true,
		Type:        "error",
		Probability: 0.05,          // 5% chance for general reliability
		Operations:  []string{"*"}, // All transaction operations
		Accounts:    []string{"*"},
		Message:     "Intermittent transaction service issues demonstrating Temporal resilience",
		MaxCount:    6, // Allow several failures for comprehensive testing
	},

	// Scenario 8: Panic simulation for transaction worker recovery (disabled by default)
	{
		Name:        "transaction_panic_recovery_demo",
		Enabled:     false, // Disabled by default - enable manually for advanced testing
		Type:        "panic",
		Probability: 1.0,
		Operations:  []string{"DebitAccount"},
		Accounts:    []string{"999999999998"}, // Specific panic test account
		Message:     "Demonstrating Temporal transaction worker recovery from panic",
		MaxCount:    1, // Only once per session
	},
}

// SimulateFailure executes failure simulation based on hardcoded learning rules
// This method is called from activities to inject controlled failures for demonstration
func (service *Service) SimulateFailure(ctx context.Context, operation string, accountID string) error {
	return service.failureSimulator.SimulateFailure(ctx, operation, accountID, learningFailureRules)
}

// GetFailureSimulationStats returns statistics about failure simulation
// This helps track how many failures have been injected during the learning session
func (service *Service) GetFailureSimulationStats() map[string]any {
	stats := service.failureSimulator.GetStats()

	// Add learning context to stats
	stats["learning_mode"] = true
	stats["service"] = "transaction"
	stats["total_rules"] = len(learningFailureRules)
	stats["enabled_rules"] = countEnabledRules()

	return stats
}

// ResetFailureSimulation resets the failure simulation state
// Useful for starting a new learning session with fresh statistics
func (service *Service) ResetFailureSimulation() {
	service.failureSimulator.Reset()
}

// countEnabledRules counts how many rules are currently enabled
func countEnabledRules() int {
	count := 0
	for _, rule := range learningFailureRules {
		if rule.Enabled {
			count++
		}
	}
	return count
}

// GetLearningScenarios returns a description of available failure scenarios
// This helps developers understand what failure scenarios are available for testing
func (service *Service) GetLearningScenarios() []map[string]any {
	scenarios := make([]map[string]any, 0, len(learningFailureRules))

	for _, rule := range learningFailureRules {
		scenario := map[string]any{
			"name":        rule.Name,
			"enabled":     rule.Enabled,
			"type":        rule.Type,
			"probability": rule.Probability,
			"operations":  rule.Operations,
			"accounts":    rule.Accounts,
			"description": rule.Message,
		}

		if rule.DelayMs > 0 {
			scenario["delay_ms"] = rule.DelayMs
		}
		if rule.TimeoutMs > 0 {
			scenario["timeout_ms"] = rule.TimeoutMs
		}
		if rule.MaxCount > 0 {
			scenario["max_count"] = rule.MaxCount
		}

		scenarios = append(scenarios, scenario)
	}

	return scenarios
}
