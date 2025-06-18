package service

import (
	"context"
	"fmt"
	"svc-transaction/store/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// AuditBalanceHistoryParams defines the parameters for creating a balance history audit record
type AuditBalanceHistoryParams struct {
	AccountID     string
	TransactionID uuid.UUID
	OldBalance    decimal.Decimal
	NewBalance    decimal.Decimal
	Operation     string
	CreatedBy     string
}

// AuditBalanceHistoryResults defines the results from creating a balance history audit record
type AuditBalanceHistoryResults struct {
	HistoryID     uuid.UUID
	AccountID     string
	TransactionID uuid.UUID
	OldBalance    decimal.Decimal
	NewBalance    decimal.Decimal
	BalanceChange decimal.Decimal
	Operation     string
	CreatedBy     string
	CreatedAt     pgtype.Timestamptz
}

// AuditBalanceHistory creates an audit record for balance changes
func (service *Service) AuditBalanceHistory(ctx context.Context, params AuditBalanceHistoryParams) (*AuditBalanceHistoryResults, error) {
	op := "audit_balance_history"
	service.logger.WithField("op", op).Infof("Starting balance history audit for account %s", params.AccountID)

	// Validate input parameters
	if err := service.validateAuditBalanceHistoryParams(params); err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Invalid audit parameters")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate balance change
	balanceChange := params.NewBalance.Sub(params.OldBalance)

	// Convert parameters to store format
	oldBalanceNumeric, err := service.decimalToPgNumeric(params.OldBalance)
	if err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Failed to convert old balance to pgtype.Numeric")
		return nil, fmt.Errorf("failed to convert old balance: %w", err)
	}

	newBalanceNumeric, err := service.decimalToPgNumeric(params.NewBalance)
	if err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Failed to convert new balance to pgtype.Numeric")
		return nil, fmt.Errorf("failed to convert new balance: %w", err)
	}

	balanceChangeNumeric, err := service.decimalToPgNumeric(balanceChange)
	if err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Failed to convert balance change to pgtype.Numeric")
		return nil, fmt.Errorf("failed to convert balance change: %w", err)
	}

	// Convert account ID to UUID
	accountUUID, err := uuid.Parse(params.AccountID)
	if err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Failed to parse account ID as UUID")
		return nil, fmt.Errorf("invalid account ID format: %w", err)
	}

	// Create the audit record
	auditRecord, err := service.store.CreateBalanceHistoryRecord(ctx, sqlc.CreateBalanceHistoryRecordParams{
		AccountID:     pgtype.UUID{Bytes: accountUUID, Valid: true},
		TransactionID: pgtype.UUID{Bytes: params.TransactionID, Valid: true},
		OldBalance:    oldBalanceNumeric,
		NewBalance:    newBalanceNumeric,
		BalanceChange: balanceChangeNumeric,
		Operation:     params.Operation,
		CreatedBy:     pgtype.Text{String: params.CreatedBy, Valid: true},
	})

	if err != nil {
		service.logger.WithField("op", op).WithError(err).Error("Failed to create balance history record")
		return nil, fmt.Errorf("failed to create audit record: %w", err)
	}

	// Convert response
	auditResults := &AuditBalanceHistoryResults{
		HistoryID:     uuid.UUID(auditRecord.ID.Bytes),
		AccountID:     params.AccountID,
		TransactionID: params.TransactionID,
		OldBalance:    params.OldBalance,
		NewBalance:    params.NewBalance,
		BalanceChange: balanceChange,
		Operation:     params.Operation,
		CreatedBy:     params.CreatedBy,
		CreatedAt:     auditRecord.CreatedAt,
	}

	service.logger.WithField("op", op).Infof("Successfully created balance history audit record for account %s", params.AccountID)
	return auditResults, nil
}

// validateAuditBalanceHistoryParams validates the audit parameters
func (service *Service) validateAuditBalanceHistoryParams(params AuditBalanceHistoryParams) error {
	if params.AccountID == "" {
		return fmt.Errorf("account_id is required")
	}

	// Validate account ID is a valid UUID
	if _, err := uuid.Parse(params.AccountID); err != nil {
		return fmt.Errorf("invalid account ID format: %w", err)
	}

	if params.TransactionID == uuid.Nil {
		return fmt.Errorf("transaction_id is required")
	}

	if params.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	if params.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate operation type
	validOperations := map[string]bool{
		"debit":        true,
		"credit":       true,
		"compensate":   true,
		"freeze":       true,
		"unfreeze":     true,
		"adjustment":   true,
		"transfer_in":  true,
		"transfer_out": true,
	}

	if !validOperations[params.Operation] {
		return fmt.Errorf("invalid operation type: %s", params.Operation)
	}

	return nil
}
