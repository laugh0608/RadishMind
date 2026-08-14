package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGatewayRequestQuotaRepository struct {
	pool *pgxpool.Pool
}

type postgresGatewayRequestQuotaQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func newPostgresGatewayRequestQuotaRepository(pool *pgxpool.Pool) *postgresGatewayRequestQuotaRepository {
	return &postgresGatewayRequestQuotaRepository{pool: pool}
}

func (repository *postgresGatewayRequestQuotaRepository) ReadPolicy(
	quotaContext GatewayRequestQuotaContext,
) (GatewayRequestQuotaPolicy, bool, error) {
	if repository.pool == nil || !validGatewayRequestQuotaContext(quotaContext) {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaContract
	}
	return readPostgresGatewayRequestQuotaPolicy(repository.pool, quotaContext, false)
}

func (repository *postgresGatewayRequestQuotaRepository) PutPolicy(
	quotaContext GatewayRequestQuotaContext,
	expectedVersion int64,
	requestLimit int64,
	now time.Time,
) (GatewayRequestQuotaPolicy, error) {
	if repository.pool == nil || !validGatewayRequestQuotaContext(quotaContext) || expectedVersion < 0 ||
		requestLimit < minimumGatewayRequestQuotaLimit || requestLimit > maximumGatewayRequestQuotaLimit || now.IsZero() {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaContract
	}
	transaction, err := repository.pool.BeginTx(quotaContext.RequestContext, pgx.TxOptions{})
	if err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if err = lockPostgresGatewayRequestQuotaScope(transaction, quotaContext); err != nil {
		return GatewayRequestQuotaPolicy{}, err
	}
	current, found, err := readPostgresGatewayRequestQuotaPolicy(transaction, quotaContext, true)
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
	var command pgconn.CommandTag
	if !found {
		command, err = transaction.Exec(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_policies
            (tenant_ref,workspace_id,environment,application_id,policy_id,record_version,request_limit,sanitized_policy_record,updated_at)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`,
			quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
			updated.PolicyID, updated.RecordVersion, updated.RequestLimit, payload, updated.UpdatedAt)
	} else {
		command, err = transaction.Exec(quotaContext.RequestContext, `UPDATE gateway_request_quota_policies SET
            record_version=$1,request_limit=$2,sanitized_policy_record=$3,updated_at=$4
            WHERE tenant_ref=$5 AND workspace_id=$6 AND environment=$7 AND application_id=$8 AND record_version=$9`,
			updated.RecordVersion, updated.RequestLimit, payload, updated.UpdatedAt,
			quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID, expectedVersion)
	}
	if err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaPolicyVersionConflict
	}
	if err = transaction.Commit(quotaContext.RequestContext); err != nil {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaStoreUnavailable
	}
	return updated, nil
}

func (repository *postgresGatewayRequestQuotaRepository) ReadUsage(
	quotaContext GatewayRequestQuotaContext,
	periodStart string,
) (GatewayRequestQuotaUsage, bool, error) {
	if repository.pool == nil || !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaPeriodStart(periodStart) {
		return GatewayRequestQuotaUsage{}, false, errGatewayRequestQuotaContract
	}
	return readPostgresGatewayRequestQuotaUsage(repository.pool, quotaContext, periodStart)
}

func (repository *postgresGatewayRequestQuotaRepository) AdmitProviderAttempt(
	quotaContext GatewayRequestQuotaContext,
	input GatewayRequestQuotaAdmissionInput,
) (GatewayRequestQuotaAdmissionDecision, error) {
	if repository.pool == nil || !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaAdmissionInput(input) {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	transaction, err := repository.pool.BeginTx(quotaContext.RequestContext, pgx.TxOptions{})
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if err = lockPostgresGatewayRequestQuotaScope(transaction, quotaContext); err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, err
	}
	policy, found, err := readPostgresGatewayRequestQuotaPolicy(transaction, quotaContext, true)
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
	command, err := transaction.Exec(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_admissions
        (tenant_ref,workspace_id,environment,application_id,request_id,admission_id,api_key_id,request_route,period_start,
         policy_id,policy_version,admitted_at,sanitized_admission_record)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT DO NOTHING`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
		input.RequestID, decision.AdmissionID, input.APIKeyID, input.Route, periodStart,
		policy.PolicyID, policy.RecordVersion, input.AdmittedAt.UTC(), decisionPayload)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaAttemptConflict
	}
	command, err = transaction.Exec(quotaContext.RequestContext, `INSERT INTO gateway_request_quota_usage
        (tenant_ref,workspace_id,environment,application_id,period_start,admitted_request_count,updated_at)
        VALUES ($1,$2,$3,$4,$5,1,$6)
        ON CONFLICT(tenant_ref,workspace_id,environment,application_id,period_start) DO UPDATE SET
          admitted_request_count=gateway_request_quota_usage.admitted_request_count+1,
          updated_at=excluded.updated_at
        WHERE gateway_request_quota_usage.admitted_request_count < $7`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID,
		periodStart, input.AdmittedAt.UTC(), policy.RequestLimit)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaExceeded
	}
	usage, found, err := readPostgresGatewayRequestQuotaUsage(transaction, quotaContext, periodStart)
	if err != nil || !found {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	decision.Usage = usage
	decisionPayload, err = json.Marshal(decision)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	command, err = transaction.Exec(quotaContext.RequestContext, `UPDATE gateway_request_quota_admissions
        SET sanitized_admission_record=$1
        WHERE tenant_ref=$2 AND workspace_id=$3 AND environment=$4 AND application_id=$5 AND request_id=$6`,
		decisionPayload, quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment,
		quotaContext.ApplicationID, input.RequestID)
	if err != nil || command.RowsAffected() != 1 {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	if err = transaction.Commit(quotaContext.RequestContext); err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaStoreUnavailable
	}
	return decision, nil
}

func readPostgresGatewayRequestQuotaPolicy(
	query postgresGatewayRequestQuotaQueryer,
	quotaContext GatewayRequestQuotaContext,
	forUpdate bool,
) (GatewayRequestQuotaPolicy, bool, error) {
	statement := `SELECT sanitized_policy_record FROM gateway_request_quota_policies
        WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4`
	if forUpdate {
		statement += " FOR UPDATE"
	}
	var payload []byte
	err := query.QueryRow(quotaContext.RequestContext, statement,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return GatewayRequestQuotaPolicy{}, false, nil
	}
	if err != nil {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaStoreUnavailable
	}
	var policy GatewayRequestQuotaPolicy
	if json.Unmarshal(payload, &policy) != nil || !validStoredGatewayRequestQuotaPolicy(policy, quotaContext) {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaStoreUnavailable
	}
	return policy, true, nil
}

func readPostgresGatewayRequestQuotaUsage(
	query postgresGatewayRequestQuotaQueryer,
	quotaContext GatewayRequestQuotaContext,
	periodStart string,
) (GatewayRequestQuotaUsage, bool, error) {
	var admitted, policyVersion, requestLimit int64
	var updatedAt time.Time
	var policyID string
	err := query.QueryRow(quotaContext.RequestContext, `SELECT u.admitted_request_count,u.updated_at,
        p.policy_id,p.record_version,p.request_limit FROM gateway_request_quota_usage u
        JOIN gateway_request_quota_policies p USING (tenant_ref,workspace_id,environment,application_id)
        WHERE u.tenant_ref=$1 AND u.workspace_id=$2 AND u.environment=$3 AND u.application_id=$4 AND u.period_start=$5`,
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID, periodStart).
		Scan(&admitted, &updatedAt, &policyID, &policyVersion, &requestLimit)
	if errors.Is(err, pgx.ErrNoRows) {
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
		UpdatedAt: updatedAt.UTC(),
	}, true, nil
}

func lockPostgresGatewayRequestQuotaScope(transaction pgx.Tx, quotaContext GatewayRequestQuotaContext) error {
	digest := sha256.Sum256([]byte(gatewayRequestQuotaScopeKey(quotaContext)))
	lockKey := int64(binary.BigEndian.Uint64(digest[:8]))
	if _, err := transaction.Exec(quotaContext.RequestContext, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
		return errGatewayRequestQuotaStoreUnavailable
	}
	return nil
}
