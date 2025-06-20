package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"svc-transaction/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CompensationAuditParams defines parameters for compensation audit operations
type CompensationAuditParams struct {
	WorkflowID                string
	RunID                     string
	TransferID                *string
	OriginalTransactionID     *uuid.UUID
	CompensationTransactionID *uuid.UUID
	CompensationReason        string
	CompensationType          string
	CompensationStatus        string
	CompensationAttempts      *int32
	CompletedAt               *time.Time
	FailureReason             *string
	TimeoutDurationMs         *int32
	Metadata                  map[string]any
}

// CreateCompensationAudit creates a new compensation audit record
func (service *Service) CreateCompensationAudit(
	ctx context.Context,
	params CompensationAuditParams,
) (*sqlc.CoreCompensationAuditTrail, error) {
	const op = "service.Service.CreateCompensationAudit"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":              op,
		"workflow_id":       params.WorkflowID,
		"transfer_id":       params.TransferID,
		"compensation_type": params.CompensationType,
	})

	logger.Info("Creating compensation audit record")

	// Serialize metadata to JSON
	var metadataJson []byte
	if params.Metadata != nil {
		var err error
		metadataJson, err = json.Marshal(params.Metadata)
		if err != nil {
			err = fmt.Errorf("failed to marshal metadata: %w", err)
			logger.WithError(err).Error()
			return nil, err
		}
	}

	// Convert parameters to pgtype values
	var transferID pgtype.Text
	if params.TransferID != nil {
		transferID.String = *params.TransferID
		transferID.Valid = true
	}

	var originalTransactionID pgtype.UUID
	if params.OriginalTransactionID != nil {
		originalTransactionID.Bytes = *params.OriginalTransactionID
		originalTransactionID.Valid = true
	}

	// Create audit record
	auditRecord, err := service.store.CreateCompensationAudit(ctx, sqlc.CreateCompensationAuditParams{
		WorkflowID:            params.WorkflowID,
		RunID:                 params.RunID,
		TransferID:            transferID,
		OriginalTransactionID: originalTransactionID,
		CompensationReason:    params.CompensationReason,
		CompensationType:      sqlc.CoreCompensationType(params.CompensationType),
		CompensationStatus:    sqlc.CoreCompensationStatus(params.CompensationStatus),
		Metadata:              metadataJson,
	})

	if err != nil {
		err = fmt.Errorf("failed to create compensation audit: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"audit_id": auditRecord.ID,
		"status":   auditRecord.CompensationStatus,
	}).Info("Compensation audit record created successfully")

	return &auditRecord, nil
}

// UpdateCompensationAudit updates an existing compensation audit record
func (service *Service) UpdateCompensationAudit(
	ctx context.Context,
	workflowID string,
	params CompensationAuditParams,
) (*sqlc.CoreCompensationAuditTrail, error) {
	const op = "service.Service.UpdateCompensationAudit"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":        op,
		"workflow_id": workflowID,
		"status":      params.CompensationStatus,
	})

	logger.Info("Updating compensation audit record")

	// Convert parameters to pgtype values
	var compensationTransactionID pgtype.UUID
	if params.CompensationTransactionID != nil {
		compensationTransactionID.Bytes = *params.CompensationTransactionID
		compensationTransactionID.Valid = true
	}

	var failureReason pgtype.Text
	if params.FailureReason != nil {
		failureReason.String = *params.FailureReason
		failureReason.Valid = true
	}

	var timeoutDurationMs pgtype.Int4
	if params.TimeoutDurationMs != nil {
		timeoutDurationMs.Int32 = *params.TimeoutDurationMs
		timeoutDurationMs.Valid = true
	}

	auditRecord, err := service.store.UpdateCompensationAudit(ctx, sqlc.UpdateCompensationAuditParams{
		WorkflowID:                workflowID,
		CompensationStatus:        sqlc.CoreCompensationStatus(params.CompensationStatus),
		CompensationTransactionID: compensationTransactionID,
		FailureReason:             failureReason,
		TimeoutDurationMs:         timeoutDurationMs,
	})

	if err != nil {
		err = fmt.Errorf("failed to update compensation audit: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"audit_id": auditRecord.ID,
		"attempts": auditRecord.CompensationAttempts,
	}).Info("Compensation audit record updated successfully")

	return &auditRecord, nil
}

// GetCompensationAuditByWorkflowID retrieves compensation audit records for a workflow
func (service *Service) GetCompensationAuditByWorkflowID(
	ctx context.Context,
	workflowID string,
) ([]sqlc.CoreCompensationAuditTrail, error) {
	const op = "service.Service.GetCompensationAuditByWorkflowID"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":        op,
		"workflow_id": workflowID,
	})

	logger.Debug("Getting compensation audit records by workflow ID")

	records, err := service.store.GetCompensationAuditByWorkflowID(ctx, workflowID)
	if err != nil {
		err = fmt.Errorf("failed to get compensation audit records: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithField("record_count", len(records)).Debug("Retrieved compensation audit records")

	return records, nil
}

// GetCompensationStats returns compensation statistics for monitoring
func (service *Service) GetCompensationStats(ctx context.Context) (*sqlc.GetCompensationStatsRow, error) {
	const op = "service.Service.GetCompensationStats"

	logger := service.logger.WithField("[op]", op)
	logger.Debug("Getting compensation statistics")

	stats, err := service.store.GetCompensationStats(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get compensation stats: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"total_compensations":     stats.TotalCompensations,
		"completed_compensations": stats.CompletedCompensations,
		"failed_compensations":    stats.FailedCompensations,
	}).Debug("Retrieved compensation statistics")

	return &stats, nil
}

// Enhanced compensation activity params with timeout and retry configuration
type EnhancedCompensateDebitParams struct {
	// Base activity parameters
	OriginalTransactionID string          `json:"original_transaction_id"`
	AccountID             string          `json:"account_id"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	ReferenceID           string          `json:"reference_id"`
	IdempotencyKey        string          `json:"idempotency_key"`
	TransferID            string          `json:"transfer_id"`
	WorkflowID            string          `json:"workflow_id"`
	RunID                 string          `json:"run_id"`

	// Enhanced compensation parameters
	TimeoutMs          int32
	MaxRetries         int32
	BackoffMultiplier  float64
	EnableAuditTrail   bool
	CompensationReason string
	CompensationType   string
}

// CompensateDebitWithEnhancedHandling performs debit compensation with comprehensive audit trail
func (service *Service) CompensateDebitWithEnhancedHandling(
	ctx context.Context,
	params EnhancedCompensateDebitParams,
) (*CompensateDebitResults, error) {
	const op = "service.Service.CompensateDebitWithEnhancedHandling"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":                    op,
		"workflow_id":             params.WorkflowID,
		"original_transaction_id": params.OriginalTransactionID,
		"compensation_type":       params.CompensationType,
	})

	logger.Info("Starting enhanced compensation with audit trail")

	var auditRecord *sqlc.CoreCompensationAuditTrail
	var err error

	// Create audit trail if enabled
	if params.EnableAuditTrail {
		originalTxID := uuid.MustParse(params.OriginalTransactionID)
		auditRecord, err = service.CreateCompensationAudit(ctx, CompensationAuditParams{
			WorkflowID:            params.WorkflowID,
			RunID:                 params.RunID,
			TransferID:            &params.TransferID,
			OriginalTransactionID: &originalTxID,
			CompensationReason:    params.CompensationReason,
			CompensationType:      params.CompensationType,
			CompensationStatus:    "pending",
			Metadata: map[string]any{
				"timeout_ms":         params.TimeoutMs,
				"max_retries":        params.MaxRetries,
				"backoff_multiplier": params.BackoffMultiplier,
			},
		})

		if err != nil {
			logger.WithError(err).Error("Failed to create compensation audit record")
			// Continue without audit trail in case of audit failure
		}
	}

	// ENHANCED FAILURE SIMULATION: Check for sophisticated compensation failures
	if err := service.SimulateFailure(ctx, "CompensateDebit", params.AccountID); err != nil {
		logger.WithError(err).Warn("🚨 Enhanced compensation failure simulation triggered")

		// Update audit trail with failure
		if auditRecord != nil {
			errStr := err.Error()
			service.UpdateCompensationAudit(ctx, params.WorkflowID, CompensationAuditParams{
				CompensationStatus: "failed",
				FailureReason:      &errStr,
			})
		}

		return nil, err
	}

	// Execute standard compensation logic
	originalTxID := uuid.MustParse(params.OriginalTransactionID)
	accountID := uuid.MustParse(params.AccountID)
	compensateParams := CompensateDebitParams{
		OriginalTransactionID: &originalTxID,
		AccountID:             &accountID,
		Amount:                params.Amount,
		Currency:              params.Currency,
		Description:           &params.CompensationReason,
		ReferenceID:           &params.ReferenceID,
		IdempotencyKey:        &params.IdempotencyKey,
		CompensationReason:    &params.CompensationReason,
		WorkflowID:            &params.WorkflowID,
		RunID:                 &params.RunID,
		Metadata: map[string]any{
			"transfer_id":           params.TransferID,
			"enhanced_compensation": true,
			"audit_record_id": func() string {
				if auditRecord != nil {
					return auditRecord.ID.String()
				}
				return ""
			}(),
		},
	}

	result, err := service.CompensateDebit(ctx, compensateParams)

	// Update audit trail with final result
	if auditRecord != nil {
		var compensationTransactionID *uuid.UUID
		if result != nil {
			compensationTransactionID = &result.TransactionID
		}

		_, updateErr := service.UpdateCompensationAudit(ctx, params.WorkflowID, CompensationAuditParams{
			CompensationStatus: func() string {
				if err != nil {
					return "failed"
				}
				return "completed"
			}(),
			CompensationTransactionID: compensationTransactionID,
			FailureReason: func() *string {
				if err != nil {
					errStr := err.Error()
					return &errStr
				}
				return nil
			}(),
		})

		if updateErr != nil {
			logger.WithError(updateErr).Warn("Failed to update compensation audit record")
		}
	}

	if err != nil {
		err = fmt.Errorf("enhanced compensation failed: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"transaction_id": result.TransactionID,
		"new_balance":    result.NewBalance,
	}).Info("Enhanced compensation completed successfully")

	return result, nil
}

// GetPendingCompensations retrieves compensations that may need manual intervention
func (service *Service) GetPendingCompensations(
	ctx context.Context,
	limit int32,
) ([]sqlc.CoreCompensationAuditTrail, error) {
	const op = "service.Service.GetPendingCompensations"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]":  op,
		"limit": limit,
	})

	logger.Debug("Getting pending compensations for manual review")

	records, err := service.store.GetPendingCompensations(ctx, limit)
	if err != nil {
		err = fmt.Errorf("failed to get pending compensations: %w", err)
		logger.WithError(err).Error()
		return nil, err
	}

	logger.WithField("pending_count", len(records)).Debug("Retrieved pending compensations")

	return records, nil
}
