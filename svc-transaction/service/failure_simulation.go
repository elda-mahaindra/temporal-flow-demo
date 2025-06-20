package service

import (
	"context"
	"fmt"

	"svc-transaction/util/failure"

	"github.com/sirupsen/logrus"
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
	// Combine regular transaction rules with enhanced compensation rules
	allRules := append(learningFailureRules, enhancedCompensationRules...)
	return service.failureSimulator.SimulateFailure(ctx, operation, accountID, allRules)
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

// Enhanced Compensation Learning Scenarios
var enhancedCompensationRules = []failure.Rule{
	// Nested Compensation Failures
	{
		Name:        "compensation_cascade_failure",
		Enabled:     true,
		Type:        "error",
		Probability: 0.9, // High probability to show cascading failures
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"cascade-failure-account"},
		Message:     "compensation cascade failure: original compensation failed, secondary compensation also failed",
		MaxCount:    3,
	},

	// Timeout-based Compensation Failures
	{
		Name:        "compensation_timeout_escalation",
		Enabled:     true,
		Type:        "timeout",
		Probability: 0.8,
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"timeout-escalation-account"},
		TimeoutMs:   5000, // 5 second timeout
		Message:     "compensation timeout: operation exceeded maximum retry attempts",
		MaxCount:    5,
	},

	// Partial Compensation Scenarios
	{
		Name:        "partial_compensation_recovery",
		Enabled:     true,
		Type:        "error",
		Probability: 0.6,
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"partial-compensation-account"},
		Message:     "partial compensation failure: some operations completed but compensation incomplete",
		MaxCount:    2,
	},

	// Enhanced Credit Compensation Failures
	{
		Name:        "credit_compensation_deadlock",
		Enabled:     true,
		Type:        "error",
		Probability: 0.7,
		Operations:  []string{"CreditAccount"},
		Accounts:    []string{"deadlock-compensation-account"},
		Message:     "credit compensation deadlock: unable to process due to resource contention",
		MaxCount:    4,
	},

	// Temporal Workflow State Corruption
	{
		Name:        "workflow_state_corruption_compensation",
		Enabled:     false, // Disabled by default due to panic
		Type:        "panic",
		Probability: 0.3,
		Operations:  []string{"CompensateDebit", "CreditAccount"},
		Accounts:    []string{"state-corruption-account"},
		Message:     "workflow state corruption: compensation process detected invalid workflow state",
		MaxCount:    1,
	},

	// Multi-Service Compensation Failures
	{
		Name:        "cross_service_compensation_failure",
		Enabled:     true,
		Type:        "timeout",
		Probability: 0.5,
		Operations:  []string{"DebitAccount"},
		Accounts:    []string{"cross-service-failure-account"},
		TimeoutMs:   4000, // 4 second timeout
		Message:     "cross-service compensation failure: downstream service unavailable during compensation",
		MaxCount:    3,
	},

	// Compensation Audit Trail Failures
	{
		Name:        "audit_trail_corruption_compensation",
		Enabled:     true,
		Type:        "error",
		Probability: 0.4,
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"audit-trail-corruption-account"},
		Message:     "audit trail corruption: compensation succeeded but audit record failed",
		MaxCount:    2,
	},

	// Resource Exhaustion During Compensation
	{
		Name:        "resource_exhaustion_compensation",
		Enabled:     true,
		Type:        "slow",
		Probability: 0.6,
		Operations:  []string{"DebitAccount", "CreditAccount", "CompensateDebit"},
		Accounts:    []string{"resource-exhaustion-account"},
		DelayMs:     8000, // 8 second delay
		Message:     "resource exhaustion: compensation delayed due to insufficient system resources",
		MaxCount:    4,
	},

	// Byzantine Failure During Compensation
	{
		Name:        "byzantine_compensation_failure",
		Enabled:     true,
		Type:        "error",
		Probability: 0.3,
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"byzantine-failure-account"},
		Message:     "byzantine failure: compensation appeared successful but data is corrupted",
		MaxCount:    2,
	},

	// Compensation Retry Exhaustion
	{
		Name:        "compensation_retry_exhaustion",
		Enabled:     true,
		Type:        "error",
		Probability: 1.0, // Always fail to demonstrate retry exhaustion
		Operations:  []string{"CompensateDebit"},
		Accounts:    []string{"retry-exhaustion-account"},
		Message:     "compensation retry exhaustion: all retry attempts failed, manual intervention required",
		MaxCount:    10, // High count to show multiple retry attempts
	},
}

// AddEnhancedCompensationScenarios adds sophisticated compensation failure scenarios
func (service *Service) AddEnhancedCompensationScenarios() {
	const op = "service.Service.AddEnhancedCompensationScenarios"

	logger := service.logger.WithField("[op]", op)
	logger.Info("Enhanced compensation failure scenarios are now available")

	logger.WithField("scenario_count", len(enhancedCompensationRules)).Info("Enhanced compensation scenarios loaded successfully")
}

// GetEnhancedCompensationScenarios returns the enhanced compensation learning scenarios
func (service *Service) GetEnhancedCompensationScenarios() []map[string]any {
	scenarios := make([]map[string]any, 0, len(enhancedCompensationRules))

	for _, rule := range enhancedCompensationRules {
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

		// Add learning notes for enhanced scenarios
		scenario["learning_note"] = getEnhancedLearningNote(rule.Name)

		scenarios = append(scenarios, scenario)
	}

	return scenarios
}

// getEnhancedLearningNote returns educational notes for enhanced compensation scenarios
func getEnhancedLearningNote(ruleName string) string {
	notes := map[string]string{
		"compensation_cascade_failure":           "Demonstrates complex compensation scenarios where compensations themselves need compensation",
		"compensation_timeout_escalation":        "Shows how Temporal handles compensation timeouts and escalation to manual processes",
		"partial_compensation_recovery":          "Demonstrates handling of partial failures in complex compensation workflows",
		"credit_compensation_deadlock":           "Shows complex recovery patterns for credit operations during compensation",
		"workflow_state_corruption_compensation": "Demonstrates Temporal's ability to recover from workflow state corruption during compensation",
		"cross_service_compensation_failure":     "Shows how Temporal handles compensation when external services become unavailable",
		"audit_trail_corruption_compensation":    "Demonstrates separation of business logic from audit concerns in compensation",
		"resource_exhaustion_compensation":       "Shows how Temporal handles resource constraints during compensation workflows",
		"byzantine_compensation_failure":         "Demonstrates detection and handling of Byzantine failures in compensation logic",
		"compensation_retry_exhaustion":          "Shows how Temporal escalates to manual processes when all automated recovery fails",
	}

	if note, exists := notes[ruleName]; exists {
		return note
	}
	return "Enhanced compensation scenario for advanced Temporal learning"
}

// TriggerEnhancedCompensationFailure manually triggers a specific enhanced compensation failure
func (service *Service) TriggerEnhancedCompensationFailure(
	ctx context.Context,
	scenarioName string,
	accountID string,
) error {
	const op = "service.Service.TriggerEnhancedCompensationFailure"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":          op,
		"scenario_name": scenarioName,
		"account_id":    accountID,
	})

	logger.Info("Triggering enhanced compensation failure scenario")

	for _, rule := range enhancedCompensationRules {
		if rule.Name == scenarioName {
			// Create a single-use rule to force trigger the scenario
			triggerRule := rule
			triggerRule.Probability = 1.0 // Force trigger
			triggerRule.Enabled = true
			triggerRule.Accounts = []string{accountID} // Target specific account

			err := service.failureSimulator.SimulateFailure(ctx, "CompensateDebit", accountID, []failure.Rule{triggerRule})
			if err != nil {
				logger.WithError(err).Warn("🚨 Enhanced compensation failure triggered manually")
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("enhanced compensation scenario '%s' not found", scenarioName)
}
