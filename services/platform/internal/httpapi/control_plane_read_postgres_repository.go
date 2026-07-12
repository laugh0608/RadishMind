package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	controlplanereadmigrations "radishmind.local/services/platform/migrations/control_plane_admin_read"
)

type postgresControlPlaneAdminReadRepository struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

type routedControlPlaneReadRepository struct {
	admin     *postgresControlPlaneAdminReadRepository
	workspace ControlPlaneReadRepository
}

type controlPlaneAuditCursor struct {
	Version    string `json:"version"`
	RecordedAt string `json:"recorded_at"`
	AuditRef   string `json:"audit_ref"`
}

const controlPlaneAuditCursorVersion = "control_plane_audit_cursor.v1"

func newRoutedControlPlaneReadRepository(pool *pgxpool.Pool, timeout time.Duration, workspace ControlPlaneReadRepository) ControlPlaneReadRepository {
	return &routedControlPlaneReadRepository{
		admin:     &postgresControlPlaneAdminReadRepository{pool: pool, timeout: timeout},
		workspace: workspace,
	}
}

func (repository *postgresControlPlaneAdminReadRepository) ReadTenantSummary(repositoryContext ReadRepositoryContext, _ ReadTenantSummaryRequest) ReadTenantSummaryResult {
	result := ReadTenantSummaryResult{TenantRef: repositoryContext.TenantRef, Items: []TenantSummary{}, AuditRef: repositoryContext.AuditRef}
	if strings.TrimSpace(repositoryContext.TenantRef) == "" {
		result.FailureCode = ReadRepositoryFailureTenantBindingMissing
		return result
	}
	if repository == nil || repository.pool == nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	ctx, cancel := controlPlaneReadDatabaseContext(repositoryContext, repository.timeout)
	defer cancel()
	var item TenantSummary
	var schemaVersion string
	err := repository.pool.QueryRow(ctx, `SELECT tenant_ref, schema_version, tenant_display_name, tenant_state,
        plan_ref, quota_summary_ref, deployment_status_ref, audit_summary_ref
        FROM control_plane_tenant_summary_projections WHERE tenant_ref=$1`, repositoryContext.TenantRef).Scan(
		&item.TenantRef, &schemaVersion, &item.TenantDisplayName, &item.TenantState,
		&item.PlanRef, &item.QuotaSummaryRef, &item.DeploymentStatusRef, &item.AuditSummaryRef,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return result
	}
	if err != nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	if schemaVersion != controlplanereadmigrations.TenantProjectionSchemaVersion || !validTenantSummaryProjection(item, repositoryContext.TenantRef) {
		result.FailureCode = ReadRepositoryFailureContractMismatch
		return result
	}
	result.Items = []TenantSummary{item}
	return result
}

func (repository *postgresControlPlaneAdminReadRepository) ListAuditSummaries(repositoryContext ReadRepositoryContext, request ListAuditSummariesRequest) ListAuditSummariesResult {
	result := ListAuditSummariesResult{TenantRef: repositoryContext.TenantRef, Items: []AuditSummary{}, AuditRef: repositoryContext.AuditRef}
	if strings.TrimSpace(repositoryContext.TenantRef) == "" {
		result.FailureCode = ReadRepositoryFailureTenantBindingMissing
		return result
	}
	if repository == nil || repository.pool == nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	if !validControlPlaneAuditRepositoryRequest(request.ReadRepositoryRequest) {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	cursor, err := decodeControlPlaneAuditCursor(request.Cursor)
	if err != nil {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	query, arguments := controlPlaneAuditQuery(repositoryContext.TenantRef, request, cursor)
	ctx, cancel := controlPlaneReadDatabaseContext(repositoryContext, repository.timeout)
	defer cancel()
	rows, err := repository.pool.Query(ctx, query, arguments...)
	if err != nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	defer rows.Close()
	items := make([]AuditSummary, 0, request.Limit+1)
	for rows.Next() {
		var item AuditSummary
		var schemaVersion string
		var failureCode *string
		var recordedAt time.Time
		if err := rows.Scan(&item.TenantRef, &item.AuditRef, &schemaVersion, &item.ActorSubjectRef, &item.EventKind,
			&item.ResourceRef, &item.Decision, &failureCode, &item.TraceID, &recordedAt); err != nil {
			result.FailureCode = ReadRepositoryFailureContractMismatch
			return result
		}
		item.RecordedAt = recordedAt.UTC().Format(time.RFC3339Nano)
		if failureCode != nil {
			value := ReadRepositoryFailureCode(*failureCode)
			item.FailureCode = &value
		}
		if schemaVersion != controlplanereadmigrations.AuditProjectionSchemaVersion || !validAuditSummaryProjection(item, repositoryContext.TenantRef) {
			result.FailureCode = ReadRepositoryFailureContractMismatch
			return result
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	if len(items) > request.Limit {
		items = items[:request.Limit]
		encoded, encodeErr := encodeControlPlaneAuditCursor(items[len(items)-1])
		if encodeErr != nil {
			result.FailureCode = ReadRepositoryFailureContractMismatch
			return result
		}
		result.NextCursor = &encoded
	}
	result.Items = items
	return result
}

func validControlPlaneAuditRepositoryRequest(request ReadRepositoryRequest) bool {
	if request.Limit < 1 || request.Limit > 100 || request.Sort != "recorded_at_desc" || len(request.Cursor) > 1024 {
		return false
	}
	allowed := map[string]bool{"event_kind": true, "resource_ref": true, "actor_subject_ref": true, "failure_code": true}
	for key := range request.Filters {
		if !allowed[key] {
			return false
		}
	}
	return true
}

func controlPlaneAuditQuery(tenantRef string, request ListAuditSummariesRequest, cursor *controlPlaneAuditCursor) (string, []any) {
	arguments := []any{tenantRef}
	conditions := []string{"tenant_ref=$1"}
	for _, key := range []string{"event_kind", "resource_ref", "actor_subject_ref", "failure_code"} {
		if value := strings.TrimSpace(request.Filters[key]); value != "" {
			arguments = append(arguments, value)
			conditions = append(conditions, fmt.Sprintf("%s=$%d", key, len(arguments)))
		}
	}
	if cursor != nil {
		arguments = append(arguments, cursor.RecordedAt, cursor.AuditRef)
		conditions = append(conditions, fmt.Sprintf("(recorded_at, audit_ref) < ($%d, $%d)", len(arguments)-1, len(arguments)))
	}
	arguments = append(arguments, request.Limit+1)
	query := `SELECT tenant_ref, audit_ref, schema_version, actor_subject_ref, event_kind, resource_ref,
        decision, failure_code, trace_id, recorded_at FROM control_plane_audit_summary_projections WHERE ` +
		strings.Join(conditions, " AND ") + fmt.Sprintf(" ORDER BY recorded_at DESC, audit_ref DESC LIMIT $%d", len(arguments))
	return query, arguments
}

func encodeControlPlaneAuditCursor(item AuditSummary) (string, error) {
	if _, err := time.Parse(time.RFC3339Nano, item.RecordedAt); err != nil || strings.TrimSpace(item.AuditRef) == "" {
		return "", errors.New("invalid audit cursor source")
	}
	payload, err := json.Marshal(controlPlaneAuditCursor{Version: controlPlaneAuditCursorVersion, RecordedAt: item.RecordedAt, AuditRef: item.AuditRef})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeControlPlaneAuditCursor(encoded string) (*controlPlaneAuditCursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid audit cursor encoding")
	}
	var cursor controlPlaneAuditCursor
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil || cursor.Version != controlPlaneAuditCursorVersion || strings.TrimSpace(cursor.AuditRef) == "" {
		return nil, errors.New("invalid audit cursor contract")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("invalid audit cursor trailing payload")
	}
	parsed, err := time.Parse(time.RFC3339Nano, cursor.RecordedAt)
	if err != nil {
		return nil, errors.New("invalid audit cursor time")
	}
	cursor.RecordedAt = parsed.UTC().Format(time.RFC3339Nano)
	return &cursor, nil
}

func validTenantSummaryProjection(item TenantSummary, tenantRef string) bool {
	return item.TenantRef == tenantRef && strings.TrimSpace(item.TenantDisplayName) != "" && strings.TrimSpace(item.TenantState) != "" &&
		strings.TrimSpace(item.PlanRef) != "" && strings.TrimSpace(item.QuotaSummaryRef) != "" &&
		strings.TrimSpace(item.DeploymentStatusRef) != "" && strings.TrimSpace(item.AuditSummaryRef) != ""
}

func validAuditSummaryProjection(item AuditSummary, tenantRef string) bool {
	return item.TenantRef == tenantRef && strings.TrimSpace(item.AuditRef) != "" && strings.TrimSpace(item.ActorSubjectRef) != "" &&
		strings.TrimSpace(item.EventKind) != "" && strings.TrimSpace(item.ResourceRef) != "" && strings.TrimSpace(item.Decision) != "" &&
		strings.TrimSpace(item.TraceID) != "" && strings.TrimSpace(item.RecordedAt) != ""
}

func controlPlaneReadDatabaseContext(repositoryContext ReadRepositoryContext, timeout time.Duration) (context.Context, context.CancelFunc) {
	parent := repositoryContext.RequestContext
	if parent == nil {
		parent = context.Background()
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(parent, timeout)
}

func (repository *routedControlPlaneReadRepository) ReadTenantSummary(context ReadRepositoryContext, request ReadTenantSummaryRequest) ReadTenantSummaryResult {
	return repository.admin.ReadTenantSummary(context, request)
}

func (repository *routedControlPlaneReadRepository) ListAuditSummaries(context ReadRepositoryContext, request ListAuditSummariesRequest) ListAuditSummariesResult {
	return repository.admin.ListAuditSummaries(context, request)
}

func (repository *routedControlPlaneReadRepository) ListApplicationSummaries(context ReadRepositoryContext, request ListApplicationSummariesRequest) ListApplicationSummariesResult {
	return repository.workspace.ListApplicationSummaries(context, request)
}

func (repository *routedControlPlaneReadRepository) ListAPIKeySummaries(context ReadRepositoryContext, request ListAPIKeySummariesRequest) ListAPIKeySummariesResult {
	return repository.workspace.ListAPIKeySummaries(context, request)
}

func (repository *routedControlPlaneReadRepository) ReadQuotaSummary(context ReadRepositoryContext, request ReadQuotaSummaryRequest) ReadQuotaSummaryResult {
	return repository.workspace.ReadQuotaSummary(context, request)
}

func (repository *routedControlPlaneReadRepository) ListWorkflowDefinitionSummaries(context ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	return repository.workspace.ListWorkflowDefinitionSummaries(context, request)
}

func (repository *routedControlPlaneReadRepository) ListRunRecordSummaries(context ReadRepositoryContext, request ListRunRecordSummariesRequest) ListRunRecordSummariesResult {
	return repository.workspace.ListRunRecordSummaries(context, request)
}

func (repository *routedControlPlaneReadRepository) SideEffects() controlPlaneReadSideEffects {
	return repository.workspace.SideEffects()
}
