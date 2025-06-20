-- name: CreateCompensationAudit :one
INSERT INTO core.compensation_audit_trail (
    workflow_id,
    run_id,
    transfer_id,
    original_transaction_id,
    compensation_reason,
    compensation_type,
    compensation_status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateCompensationAudit :one
UPDATE core.compensation_audit_trail 
SET 
    compensation_status = $2,
    compensation_transaction_id = $3,
    compensation_attempts = compensation_attempts + 1,
    completed_at = CASE WHEN $2 IN ('completed', 'failed', 'timeout', 'manual_required') THEN NOW() ELSE completed_at END,
    failure_reason = $4,
    timeout_duration_ms = $5,
    updated_at = NOW()
WHERE workflow_id = $1
RETURNING *;

-- name: GetCompensationAuditByWorkflowID :many
SELECT * FROM core.compensation_audit_trail 
WHERE workflow_id = $1 
ORDER BY created_at DESC;

-- name: GetCompensationAuditByTransferID :many
SELECT * FROM core.compensation_audit_trail 
WHERE transfer_id = $1 
ORDER BY created_at DESC;

-- name: GetPendingCompensations :many
SELECT * FROM core.compensation_audit_trail 
WHERE compensation_status = 'pending' 
AND created_at < NOW() - INTERVAL '5 minutes'
ORDER BY created_at ASC
LIMIT $1;

-- name: GetCompensationStats :one
SELECT 
    COUNT(*) as total_compensations,
    COUNT(*) FILTER (WHERE compensation_status = 'completed') as completed_compensations,
    COUNT(*) FILTER (WHERE compensation_status = 'failed') as failed_compensations,
    COUNT(*) FILTER (WHERE compensation_status = 'timeout') as timeout_compensations,
    COUNT(*) FILTER (WHERE compensation_status = 'manual_required') as manual_compensations,
    COUNT(*) FILTER (WHERE compensation_status = 'pending') as pending_compensations,
    AVG(compensation_attempts) as avg_attempts
FROM core.compensation_audit_trail
WHERE created_at >= NOW() - INTERVAL '24 hours';

-- name: GetFailedCompensationsByTimeoutDuration :many
SELECT 
    workflow_id,
    transfer_id,
    compensation_reason,
    timeout_duration_ms,
    compensation_attempts,
    created_at,
    failure_reason
FROM core.compensation_audit_trail 
WHERE compensation_status = 'timeout' 
AND timeout_duration_ms > $1
ORDER BY timeout_duration_ms DESC, created_at DESC
LIMIT $2; 