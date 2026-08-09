package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteGatewayRequestQuotaRepository struct {
	database *sql.DB
}

type sqliteGatewayRequestQuotaQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func newSQLiteGatewayRequestQuotaRepository(database *sql.DB) *sqliteGatewayRequestQuotaRepository {
	return &sqliteGatewayRequestQuotaRepository{database: database}
}

func (repository *sqliteGatewayRequestQuotaRepository) ReadPolicy(
	quotaContext GatewayRequestQuotaContext,
) (GatewayRequestQuotaPolicy, bool, error) {
	if repository.database == nil || !validGatewayRequestQuotaContext(quotaContext) {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaContract
	}
	return readSQLiteGatewayRequestQuotaPolicy(repository.database, quotaContext)
}

func (repository *sqliteGatewayRequestQuotaRepository) PutPolicy(
	quotaContext GatewayRequestQuotaContext,
	expectedVersion int64,
	requestLimit int64,
	now time.Time,
) (GatewayRequestQuotaPolicy, error) {
	if repository.database == nil || !validGatewayRequestQuotaContext(quotaContext) || expectedVersion < 0 ||
		requestLimit < minimumGatewayRequestQuotaLimit || requestLimit > maximumGatewayRequestQuotaLimit || now.IsZero() {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaContract
	}
	connection, err := beginImmediateSQLiteGatewayRequestQuotaTransaction(repository.database, quotaContext.RequestContext)
	if err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	defer connection.Close()
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	current, found, err := readSQLiteGatewayRequestQuotaPolicy(connection, quotaContext)
	if err != nil {
		return GatewayRequestQuotaPolicy{}, err
	}
	if (!found && expectedVersion != 0) || (found && current.RecordVersion != expectedVersion) {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaPolicyVersionConflict
	}
	updated := buildGatewayRequestQuotaPolicy(quotaContext, current, found, expectedVersion, requestLimit, now)
	payload, err := json.Marshal(updated)
	if err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaContract
	}
	var result sql.Result
	if !found {
		result, err = connection.ExecContext(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_policies
            (tenant_ref,workspace_id,environment,application_id,policy_id,record_version,request_limit,sanitized_policy_record,updated_at_unix_nano)
            VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
			quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
			updated.PolicyID, updated.RecordVersion, updated.RequestLimit, string(payload), updated.UpdatedAt.UnixNano())
	} else {
		result, err = connection.ExecContext(quotaContext.RequestContext, `UPDATE gateway_request_quota_policies SET
            record_version=?,request_limit=?,sanitized_policy_record=?,updated_at_unix_nano=?
            WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND record_version=?`,
			updated.RecordVersion, updated.RequestLimit, string(payload), updated.UpdatedAt.UnixNano(),
			quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID, expectedVersion)
	}
	if err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil || affected != 1 {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaPolicyVersionConflict
	}
	if _, err = connection.ExecContext(quotaContext.RequestContext, "COMMIT"); err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	committed = true
	return updated, nil
}

func (repository *sqliteGatewayRequestQuotaRepository) ReadUsage(
	quotaContext GatewayRequestQuotaContext,
	periodStart string,
) (GatewayRequestQuotaUsage, bool, error) {
	if repository.database == nil || !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaPeriodStart(periodStart) {
		return GatewayRequestQuotaUsage{}, false, errGatewayRequestQuotaContract
	}
	return readSQLiteGatewayRequestQuotaUsage(repository.database, quotaContext, periodStart)
}

func (repository *sqliteGatewayRequestQuotaRepository) AdmitProviderAttempt(
	quotaContext GatewayRequestQuotaContext,
	input GatewayRequestQuotaAdmissionInput,
) (GatewayRequestQuotaAdmissionDecision, error) {
	if repository.database == nil || !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaAdmissionInput(input) {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	connection, err := beginImmediateSQLiteGatewayRequestQuotaTransaction(repository.database, quotaContext.RequestContext)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	defer connection.Close()
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	policy, found, err := readSQLiteGatewayRequestQuotaPolicy(connection, quotaContext)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, err
	}
	if !found {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaPolicyNotFound
	}
	periodStart := gatewayRequestQuotaPeriodStart(input.AdmittedAt)
	decision := GatewayRequestQuotaAdmissionDecision{
		SchemaVersion: GatewayRequestQuotaSchemaVersion,
		AdmissionID:   gatewayRequestQuotaAdmissionID(quotaContext, input.RequestID),
		APIKeyID:      input.APIKeyID,
		RequestID:     input.RequestID,
		Route:         input.Route,
		AdmittedAt:    input.AdmittedAt.UTC(),
		PolicyID:      policy.PolicyID,
		PolicyVersion: policy.RecordVersion,
	}
	decisionPayload, err := json.Marshal(decision)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	result, err := connection.ExecContext(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_admissions
        (tenant_ref,workspace_id,environment,application_id,request_id,admission_id,api_key_id,request_route,period_start,
         policy_id,policy_version,admitted_at_unix_nano,sanitized_admission_record)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
		input.RequestID, decision.AdmissionID, input.APIKeyID, input.Route, periodStart,
		policy.PolicyID, policy.RecordVersion, input.AdmittedAt.UTC().UnixNano(), string(decisionPayload))
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	} else if affected != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaAttemptConflict
	}
	result, err = connection.ExecContext(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_usage
        (tenant_ref,workspace_id,environment,application_id,period_start,admitted_request_count,updated_at_unix_nano)
        VALUES (?,?,?,?,?,1,?)
        ON CONFLICT(tenant_ref,workspace_id,environment,application_id,period_start) DO UPDATE SET
          admitted_request_count=gateway_request_quota_usage.admitted_request_count+1,
          updated_at_unix_nano=excluded.updated_at_unix_nano
        WHERE gateway_request_quota_usage.admitted_request_count < ?`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
		periodStart, input.AdmittedAt.UTC().UnixNano(), policy.RequestLimit)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	} else if affected != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaExceeded
	}
	usage, found, err := readSQLiteGatewayRequestQuotaUsage(connection, quotaContext, periodStart)
	if err != nil || !found {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	decision.Usage = usage
	decisionPayload, err = json.Marshal(decision)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	result, err = connection.ExecContext(quotaContext.RequestContext, `UPDATE gateway_request_quota_admissions
        SET sanitized_admission_record=?
        WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND request_id=?`,
		string(decisionPayload), quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment,
		quotaContext.ApplicationID, input.RequestID)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil || affected != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if _, err = connection.ExecContext(quotaContext.RequestContext, "COMMIT"); err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	committed = true
	return decision, nil
}

func beginImmediateSQLiteGatewayRequestQuotaTransaction(database *sql.DB, ctx context.Context) (*sql.Conn, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = connection.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

func readSQLiteGatewayRequestQuotaPolicy(
	query sqliteGatewayRequestQuotaQueryer,
	quotaContext GatewayRequestQuotaContext,
) (GatewayRequestQuotaPolicy, bool, error) {
	var payload string
	err := query.QueryRowContext(quotaContext.RequestContext, `SELECT sanitized_policy_record FROM gateway_request_quota_policies
        WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=?`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return GatewayRequestQuotaPolicy{}, false, nil
	}
	if err != nil {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaStoreUnavailable
	}
	var policy GatewayRequestQuotaPolicy
	if json.Unmarshal([]byte(payload), &policy) != nil || !validStoredGatewayRequestQuotaPolicy(policy, quotaContext) {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaStoreUnavailable
	}
	return policy, true, nil
}

func readSQLiteGatewayRequestQuotaUsage(
	query sqliteGatewayRequestQuotaQueryer,
	quotaContext GatewayRequestQuotaContext,
	periodStart string,
) (GatewayRequestQuotaUsage, bool, error) {
	var admitted int64
	var updatedNano int64
	var policyID string
	var policyVersion, requestLimit int64
	err := query.QueryRowContext(quotaContext.RequestContext, `SELECT u.admitted_request_count,u.updated_at_unix_nano,
        p.policy_id,p.record_version,p.request_limit FROM gateway_request_quota_usage u
        JOIN gateway_request_quota_policies p USING (tenant_ref,workspace_id,environment,application_id)
        WHERE u.tenant_ref=? AND u.workspace_id=? AND u.environment=? AND u.application_id=? AND u.period_start=?`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID, periodStart).
		Scan(&admitted, &updatedNano, &policyID, &policyVersion, &requestLimit)
	if errors.Is(err, sql.ErrNoRows) {
		return GatewayRequestQuotaUsage{}, false, nil
	}
	if err != nil {
		return GatewayRequestQuotaUsage{}, false, errGatewayRequestQuotaStoreUnavailable
	}
	return GatewayRequestQuotaUsage{
		SchemaVersion: GatewayRequestQuotaSchemaVersion, TenantRef: quotaContext.TenantRef,
		WorkspaceID: quotaContext.WorkspaceID, Environment: quotaContext.Environment,
		ApplicationID: quotaContext.ApplicationID, Period: GatewayRequestQuotaPeriod, PeriodStart: periodStart,
		PolicyID: policyID, PolicyVersion: policyVersion, RequestLimit: requestLimit,
		AdmittedRequestCount: admitted, RemainingRequestCount: maxInt64(requestLimit-admitted, 0),
		UpdatedAt: time.Unix(0, updatedNano).UTC(),
	}, true, nil
}

func buildGatewayRequestQuotaPolicy(
	quotaContext GatewayRequestQuotaContext,
	current GatewayRequestQuotaPolicy,
	found bool,
	expectedVersion int64,
	requestLimit int64,
	now time.Time,
) GatewayRequestQuotaPolicy {
	now = now.UTC()
	if !found {
		current = GatewayRequestQuotaPolicy{
			SchemaVersion: GatewayRequestQuotaSchemaVersion, PolicyID: gatewayRequestQuotaPolicyID(quotaContext),
			TenantRef: quotaContext.TenantRef, WorkspaceID: quotaContext.WorkspaceID,
			Environment: quotaContext.Environment, ApplicationID: quotaContext.ApplicationID,
			Period: GatewayRequestQuotaPeriod, CreatedAt: now, CreatedBy: quotaContext.ActorRef,
		}
	}
	current.RequestLimit = requestLimit
	current.RecordVersion = expectedVersion + 1
	current.UpdatedAt = now
	current.UpdatedBy = quotaContext.ActorRef
	current.LastRequestID = quotaContext.RequestID
	current.LastAuditRef = quotaContext.AuditRef
	return current
}

func validStoredGatewayRequestQuotaPolicy(policy GatewayRequestQuotaPolicy, quotaContext GatewayRequestQuotaContext) bool {
	return policy.SchemaVersion == GatewayRequestQuotaSchemaVersion && policy.PolicyID == gatewayRequestQuotaPolicyID(quotaContext) &&
		policy.TenantRef == quotaContext.TenantRef && policy.WorkspaceID == quotaContext.WorkspaceID &&
		policy.Environment == quotaContext.Environment && policy.ApplicationID == quotaContext.ApplicationID &&
		policy.Period == GatewayRequestQuotaPeriod && policy.RecordVersion > 0 &&
		policy.RequestLimit >= minimumGatewayRequestQuotaLimit && policy.RequestLimit <= maximumGatewayRequestQuotaLimit &&
		!policy.CreatedAt.IsZero() && !policy.UpdatedAt.IsZero() && !policy.UpdatedAt.Before(policy.CreatedAt) &&
		gatewayRequestQuotaIdentifierPattern.MatchString(policy.CreatedBy) &&
		gatewayRequestQuotaIdentifierPattern.MatchString(policy.UpdatedBy) &&
		gatewayRequestQuotaIdentifierPattern.MatchString(policy.LastRequestID) &&
		gatewayRequestQuotaIdentifierPattern.MatchString(policy.LastAuditRef)
}
