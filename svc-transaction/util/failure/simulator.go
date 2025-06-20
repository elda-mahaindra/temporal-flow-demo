package failure

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Rule represents a failure simulation rule
type Rule struct {
	Name        string
	Enabled     bool
	Type        string   // "error", "timeout", "slow", "panic"
	Probability float64  // 0.0 to 1.0
	Operations  []string // operations to target, ["*"] for all
	Accounts    []string // account IDs to target, ["*"] for all
	Message     string   // custom error message
	DelayMs     int      // delay for "slow" type
	TimeoutMs   int      // timeout duration
	MaxCount    int      // maximum occurrences (0 = unlimited)
}

// Simulator manages failure injection for learning and testing purposes
type Simulator struct {
	logger      *logrus.Logger
	startTime   time.Time
	occurrences map[string]int // track occurrences per rule
	mutex       sync.RWMutex
}

// NewSimulator creates a new failure simulator
func NewSimulator(logger *logrus.Logger) *Simulator {
	return &Simulator{
		logger:      logger,
		startTime:   time.Now(),
		occurrences: make(map[string]int),
	}
}

// SimulateFailure checks if a failure should be injected based on the provided rules
func (s *Simulator) SimulateFailure(ctx context.Context, operation string, accountID string, rules []Rule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, rule := range rules {
		if s.shouldApplyRule(rule, operation, accountID) {
			// Track occurrence
			s.occurrences[rule.Name]++

			s.logger.WithFields(logrus.Fields{
				"rule":       rule.Name,
				"operation":  operation,
				"account_id": accountID,
				"type":       rule.Type,
				"occurrence": s.occurrences[rule.Name],
			}).Warn("🚨 Injecting simulated transaction failure for Temporal testing")

			return s.executeFailure(ctx, rule)
		}
	}

	return nil
}

// shouldApplyRule determines if a failure rule should be applied
func (s *Simulator) shouldApplyRule(rule Rule, operation string, accountID string) bool {
	// Check if rule is enabled
	if !rule.Enabled {
		return false
	}

	// Check max occurrences
	if rule.MaxCount > 0 && s.occurrences[rule.Name] >= rule.MaxCount {
		return false
	}

	// Check operation match
	if !s.matchesOperation(rule.Operations, operation) {
		return false
	}

	// Check account match
	if !s.matchesAccount(rule.Accounts, accountID) {
		return false
	}

	// Check probability
	if rule.Probability > 0 && rand.Float64() > rule.Probability {
		return false
	}

	return true
}

// matchesOperation checks if the operation matches the rule's operation filters
func (s *Simulator) matchesOperation(operations []string, operation string) bool {
	if len(operations) == 0 {
		return true // no filter means match all
	}

	for _, op := range operations {
		if op == "*" || strings.EqualFold(op, operation) {
			return true
		}
	}
	return false
}

// matchesAccount checks if the account matches the rule's account filters
func (s *Simulator) matchesAccount(accounts []string, accountID string) bool {
	if len(accounts) == 0 {
		return true // no filter means match all
	}

	for _, acc := range accounts {
		if acc == "*" || acc == accountID {
			return true
		}
	}
	return false
}

// executeFailure executes the specified failure type
func (s *Simulator) executeFailure(ctx context.Context, rule Rule) error {
	switch rule.Type {
	case "error":
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf("Learning demo transaction failure: %s", rule.Name)
		}
		return fmt.Errorf("DEMO_TRANSACTION_FAILURE: %s", message)

	case "timeout":
		timeoutDuration := time.Duration(rule.TimeoutMs) * time.Millisecond
		if timeoutDuration == 0 {
			timeoutDuration = 5 * time.Second
		}

		s.logger.WithField("timeout", timeoutDuration).Debug("⏱️ Simulating transaction timeout for Temporal testing")

		select {
		case <-time.After(timeoutDuration):
			return fmt.Errorf("DEMO_TRANSACTION_TIMEOUT: operation timed out after %v", timeoutDuration)
		case <-ctx.Done():
			return ctx.Err()
		}

	case "slow":
		delay := time.Duration(rule.DelayMs) * time.Millisecond
		if delay == 0 {
			delay = 2 * time.Second
		}

		s.logger.WithField("delay", delay).Debug("🐌 Simulating slow transaction for Temporal testing")

		select {
		case <-time.After(delay):
			return nil // Continue normally after delay
		case <-ctx.Done():
			return ctx.Err()
		}

	case "panic":
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf("Learning demo transaction panic: %s", rule.Name)
		}
		panic(fmt.Sprintf("DEMO_TRANSACTION_PANIC: %s", message))

	default:
		return fmt.Errorf("DEMO_TRANSACTION_FAILURE: unknown failure type: %s", rule.Type)
	}
}

// GetStats returns statistics about failure simulation
func (s *Simulator) GetStats() map[string]any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := map[string]any{
		"uptime_ms":   time.Since(s.startTime).Milliseconds(),
		"occurrences": make(map[string]int),
	}

	// Copy occurrences to avoid race conditions
	for rule, count := range s.occurrences {
		stats["occurrences"].(map[string]int)[rule] = count
	}

	return stats
}

// Reset resets the failure simulator state
func (s *Simulator) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.startTime = time.Now()
	s.occurrences = make(map[string]int)

	s.logger.Info("🔄 Transaction failure simulator state reset for new learning session")
}
