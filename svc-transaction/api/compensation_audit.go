package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CompensationStatsResponse represents the response for compensation statistics
type CompensationStatsResponse struct {
	TotalCompensations     int64   `json:"total_compensations"`
	CompletedCompensations int64   `json:"completed_compensations"`
	FailedCompensations    int64   `json:"failed_compensations"`
	TimeoutCompensations   int64   `json:"timeout_compensations"`
	ManualCompensations    int64   `json:"manual_compensations"`
	PendingCompensations   int64   `json:"pending_compensations"`
	AverageAttempts        float64 `json:"average_attempts"`
	Period                 string  `json:"period"`
}

// CompensationAuditRecord represents a compensation audit record
type CompensationAuditRecord struct {
	ID                        string     `json:"id"`
	WorkflowID                string     `json:"workflow_id"`
	RunID                     string     `json:"run_id"`
	TransferID                *string    `json:"transfer_id,omitempty"`
	OriginalTransactionID     *string    `json:"original_transaction_id,omitempty"`
	CompensationTransactionID *string    `json:"compensation_transaction_id,omitempty"`
	CompensationReason        string     `json:"compensation_reason"`
	CompensationType          string     `json:"compensation_type"`
	CompensationStatus        string     `json:"compensation_status"`
	CompensationAttempts      int32      `json:"compensation_attempts"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	CompletedAt               *time.Time `json:"completed_at,omitempty"`
	FailureReason             *string    `json:"failure_reason,omitempty"`
	TimeoutDurationMs         *int32     `json:"timeout_duration_ms,omitempty"`
}

// GetCompensationStats handles GET /compensation-audit/stats
func (api *Api) GetCompensationStats(ctx *fiber.Ctx) error {
	const op = "api.Api.GetCompensationStats"

	logger := api.logger.WithField("[op]", op)
	logger.Info("Getting compensation statistics")

	// Get compensation statistics
	stats, err := api.service.GetCompensationStats(ctx.Context())
	if err != nil {
		logger.WithError(err).Error("Failed to get compensation stats")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve compensation statistics")
	}

	response := CompensationStatsResponse{
		TotalCompensations:     stats.TotalCompensations,
		CompletedCompensations: stats.CompletedCompensations,
		FailedCompensations:    stats.FailedCompensations,
		TimeoutCompensations:   stats.TimeoutCompensations,
		ManualCompensations:    stats.ManualCompensations,
		PendingCompensations:   stats.PendingCompensations,
		AverageAttempts:        stats.AvgAttempts,
		Period:                 "24h",
	}

	logger.WithFields(logrus.Fields{
		"total_compensations":     response.TotalCompensations,
		"completed_compensations": response.CompletedCompensations,
		"failed_compensations":    response.FailedCompensations,
	}).Info("Retrieved compensation statistics")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Compensation statistics retrieved successfully",
		"data":    response,
	})
}

// GetCompensationAuditByWorkflow handles GET /compensation-audit/workflow/:workflow_id
func (api *Api) GetCompensationAuditByWorkflow(ctx *fiber.Ctx) error {
	const op = "api.Api.GetCompensationAuditByWorkflow"

	workflowID := ctx.Params("workflow_id")
	if workflowID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Workflow ID is required")
	}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":        op,
		"workflow_id": workflowID,
	})
	logger.Info("Getting compensation audit records by workflow ID")

	// Get compensation audit records
	records, err := api.service.GetCompensationAuditByWorkflowID(ctx.Context(), workflowID)
	if err != nil {
		logger.WithError(err).Error("Failed to get compensation audit records")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve compensation audit records")
	}

	// Convert records to response format
	response := make([]CompensationAuditRecord, len(records))
	for i, record := range records {
		response[i] = CompensationAuditRecord{
			ID:                   record.ID.String(),
			WorkflowID:           record.WorkflowID,
			RunID:                record.RunID,
			CompensationReason:   record.CompensationReason,
			CompensationType:     string(record.CompensationType),
			CompensationStatus:   string(record.CompensationStatus),
			CompensationAttempts: record.CompensationAttempts,
			CreatedAt:            record.CreatedAt.Time,
			UpdatedAt:            record.UpdatedAt.Time,
		}

		// Handle nullable fields
		if record.TransferID.Valid {
			response[i].TransferID = &record.TransferID.String
		}
		if record.OriginalTransactionID.Valid {
			uuidObj := uuid.UUID(record.OriginalTransactionID.Bytes)
			uuidStr := uuidObj.String()
			response[i].OriginalTransactionID = &uuidStr
		}
		if record.CompensationTransactionID.Valid {
			uuidObj := uuid.UUID(record.CompensationTransactionID.Bytes)
			uuidStr := uuidObj.String()
			response[i].CompensationTransactionID = &uuidStr
		}
		if record.CompletedAt.Valid {
			response[i].CompletedAt = &record.CompletedAt.Time
		}
		if record.FailureReason.Valid {
			response[i].FailureReason = &record.FailureReason.String
		}
		if record.TimeoutDurationMs.Valid {
			response[i].TimeoutDurationMs = &record.TimeoutDurationMs.Int32
		}
	}

	logger.WithField("record_count", len(response)).Info("Retrieved compensation audit records")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Compensation audit records retrieved successfully",
		"data":    response,
		"count":   len(response),
	})
}

// GetPendingCompensations handles GET /compensation-audit/pending
func (api *Api) GetPendingCompensations(ctx *fiber.Ctx) error {
	const op = "api.Api.GetPendingCompensations"

	// Parse limit parameter
	limit := int32(50) // Default limit
	if limitParam := ctx.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseInt(limitParam, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = int32(parsedLimit)
		}
	}

	logger := api.logger.WithFields(logrus.Fields{
		"[op]":  op,
		"limit": limit,
	})
	logger.Info("Getting pending compensations")

	// Get pending compensations
	records, err := api.service.GetPendingCompensations(ctx.Context(), limit)
	if err != nil {
		logger.WithError(err).Error("Failed to get pending compensations")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve pending compensations")
	}

	// Convert records to response format (same as above)
	response := make([]CompensationAuditRecord, len(records))
	for i, record := range records {
		response[i] = CompensationAuditRecord{
			ID:                   record.ID.String(),
			WorkflowID:           record.WorkflowID,
			RunID:                record.RunID,
			CompensationReason:   record.CompensationReason,
			CompensationType:     string(record.CompensationType),
			CompensationStatus:   string(record.CompensationStatus),
			CompensationAttempts: record.CompensationAttempts,
			CreatedAt:            record.CreatedAt.Time,
			UpdatedAt:            record.UpdatedAt.Time,
		}

		// Handle nullable fields (same as above)
		if record.TransferID.Valid {
			response[i].TransferID = &record.TransferID.String
		}
		if record.OriginalTransactionID.Valid {
			uuidObj := uuid.UUID(record.OriginalTransactionID.Bytes)
			uuidStr := uuidObj.String()
			response[i].OriginalTransactionID = &uuidStr
		}
		if record.CompensationTransactionID.Valid {
			uuidObj := uuid.UUID(record.CompensationTransactionID.Bytes)
			uuidStr := uuidObj.String()
			response[i].CompensationTransactionID = &uuidStr
		}
		if record.CompletedAt.Valid {
			response[i].CompletedAt = &record.CompletedAt.Time
		}
		if record.FailureReason.Valid {
			response[i].FailureReason = &record.FailureReason.String
		}
		if record.TimeoutDurationMs.Valid {
			response[i].TimeoutDurationMs = &record.TimeoutDurationMs.Int32
		}
	}

	logger.WithField("pending_count", len(response)).Info("Retrieved pending compensations")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pending compensations retrieved successfully",
		"data":    response,
		"count":   len(response),
		"note":    "These compensations may require manual intervention",
	})
}
