package service

import (
	"context"

	"svc-balance/util/failure"
)

// Learning-focused failure simulation rules
// These are hardcoded scenarios designed to demonstrate Temporal's fault tolerance capabilities
var learningFailureRules = []failure.Rule{
	// Scenario 1: Random balance check failures to demonstrate Temporal retries
	{
		Name:        "random_balance_failures",
		Enabled:     true,
		Type:        "error",
		Probability: 0.15, // 15% chance of failure
		Operations:  []string{"CheckBalance"},
		Accounts:    []string{"*"}, // All accounts
		Message:     "Random balance check failure for Temporal retry demonstration",
		MaxCount:    5, // Limit to 5 failures for learning session
	},

	// Scenario 2: Slow operations to test Temporal timeout handling
	{
		Name:        "slow_validation_demo",
		Enabled:     true,
		Type:        "slow",
		Probability: 0.2, // 20% chance of slowness
		Operations:  []string{"ValidateAccount"},
		Accounts:    []string{"*"},
		DelayMs:     3000, // 3 second delay
		MaxCount:    3,    // Limit occurrences
	},

	// Scenario 3: Specific account timeout to show targeted failure handling
	{
		Name:        "problematic_account_timeout",
		Enabled:     true,
		Type:        "timeout",
		Probability: 1.0, // Always fail for this account
		Operations:  []string{"CheckBalance"},
		Accounts:    []string{"123456789012"}, // Specific test account
		TimeoutMs:   2000,                     // 2 second timeout
		Message:     "Problematic account demonstrating Temporal timeout handling",
		MaxCount:    2, // Limit to 2 failures
	},

	// Scenario 4: Intermittent failures for general operations
	{
		Name:        "intermittent_service_issues",
		Enabled:     true,
		Type:        "error",
		Probability: 0.1,           // 10% chance
		Operations:  []string{"*"}, // All operations
		Accounts:    []string{"*"},
		Message:     "Intermittent service issues demonstrating Temporal resilience",
		MaxCount:    8, // Allow more failures for comprehensive testing
	},

	// Scenario 5: Panic simulation (disabled by default for safety)
	{
		Name:        "panic_recovery_demo",
		Enabled:     false, // Disabled by default - enable manually for testing
		Type:        "panic",
		Probability: 1.0,
		Operations:  []string{"CheckBalance"},
		Accounts:    []string{"999999999999"}, // Specific panic test account
		Message:     "Demonstrating Temporal worker recovery from panic",
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
