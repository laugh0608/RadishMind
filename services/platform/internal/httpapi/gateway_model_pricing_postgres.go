package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGatewayModelPricingRepository struct {
	pool *pgxpool.Pool
}

type postgresGatewayModelPricingQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func newPostgresGatewayModelPricingRepository(pool *pgxpool.Pool) *postgresGatewayModelPricingRepository {
	return &postgresGatewayModelPricingRepository{pool: pool}
}

func (repository *postgresGatewayModelPricingRepository) ReadCurrent(
	pricingContext GatewayModelPricingContext,
) (GatewayModelPricingPolicy, bool, error) {
	if repository.pool == nil || !validGatewayModelPricingContext(pricingContext) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	return readPostgresGatewayModelPricingCurrent(repository.pool, pricingContext, false)
}

func (repository *postgresGatewayModelPricingRepository) ReadRevision(
	pricingContext GatewayModelPricingContext,
	recordVersion int64,
) (GatewayModelPricingPolicy, bool, error) {
	if repository.pool == nil || !validGatewayModelPricingContext(pricingContext) || recordVersion < 1 {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	return readPostgresGatewayModelPricingRevision(repository.pool, pricingContext, recordVersion)
}

func (repository *postgresGatewayModelPricingRepository) PutRevision(
	pricingContext GatewayModelPricingContext,
	input GatewayModelPricingPolicyInput,
	now time.Time,
) (GatewayModelPricingPolicy, error) {
	input = normalizeGatewayModelPricingPolicyInput(input)
	if repository.pool == nil || !validGatewayModelPricingContext(pricingContext) ||
		!validGatewayModelPricingPolicyInput(input) || now.IsZero() {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	transaction, err := repository.pool.BeginTx(pricingContext.RequestContext, pgx.TxOptions{})
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	defer func() { _ = transaction.Rollback(context.Background()) }()
	if err = lockPostgresGatewayModelPricingScope(transaction, pricingContext); err != nil {
		return GatewayModelPricingPolicy{}, err
	}
	current, found, err := readPostgresGatewayModelPricingCurrent(transaction, pricingContext, true)
	if err != nil {
		return GatewayModelPricingPolicy{}, err
	}
	currentVersion := int64(0)
	if found {
		currentVersion = current.RecordVersion
	}
	if input.ExpectedVersion != currentVersion {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	policy := buildGatewayModelPricingPolicy(pricingContext, input, currentVersion+1, now)
	if !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	payload, err := json.Marshal(policy)
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	command, err := transaction.Exec(pricingContext.RequestContext, `INSERT INTO gateway_model_pricing_revisions
        (tenant_ref,workspace_id,environment,provider_id,profile_id,model_id,record_version,policy_id,
         policy_digest,sanitized_policy_record,updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`,
		pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
		pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
		policy.RecordVersion, policy.PolicyID, policy.PolicyDigest, payload, policy.UpdatedAt)
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	if !found {
		command, err = transaction.Exec(pricingContext.RequestContext, `INSERT INTO gateway_model_pricing_current
            (tenant_ref,workspace_id,environment,provider_id,profile_id,model_id,policy_id,current_version,updated_at)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`,
			pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
			pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
			policy.PolicyID, policy.RecordVersion, policy.UpdatedAt)
	} else {
		command, err = transaction.Exec(pricingContext.RequestContext, `UPDATE gateway_model_pricing_current SET
            current_version=$1,updated_at=$2
            WHERE tenant_ref=$3 AND workspace_id=$4 AND environment=$5 AND provider_id=$6 AND profile_id=$7 AND model_id=$8
              AND policy_id=$9 AND current_version=$10`,
			policy.RecordVersion, policy.UpdatedAt,
			pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
			pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
			policy.PolicyID, currentVersion)
	}
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	if err = transaction.Commit(pricingContext.RequestContext); err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	return policy, nil
}

func readPostgresGatewayModelPricingCurrent(
	query postgresGatewayModelPricingQueryer,
	pricingContext GatewayModelPricingContext,
	forUpdate bool,
) (GatewayModelPricingPolicy, bool, error) {
	statement := `SELECT r.sanitized_policy_record
        FROM gateway_model_pricing_current c
        JOIN gateway_model_pricing_revisions r ON
          r.tenant_ref=c.tenant_ref AND r.workspace_id=c.workspace_id AND r.environment=c.environment AND
          r.provider_id=c.provider_id AND r.profile_id=c.profile_id AND r.model_id=c.model_id AND
          r.record_version=c.current_version
        WHERE c.tenant_ref=$1 AND c.workspace_id=$2 AND c.environment=$3 AND
          c.provider_id=$4 AND c.profile_id=$5 AND c.model_id=$6`
	if forUpdate {
		statement += " FOR UPDATE OF c"
	}
	return readPostgresGatewayModelPricingPolicy(query, pricingContext, statement, nil)
}

func readPostgresGatewayModelPricingRevision(
	query postgresGatewayModelPricingQueryer,
	pricingContext GatewayModelPricingContext,
	recordVersion int64,
) (GatewayModelPricingPolicy, bool, error) {
	return readPostgresGatewayModelPricingPolicy(query, pricingContext, `SELECT sanitized_policy_record
        FROM gateway_model_pricing_revisions
        WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND provider_id=$4 AND profile_id=$5 AND model_id=$6
          AND record_version=$7`, &recordVersion)
}

func readPostgresGatewayModelPricingPolicy(
	query postgresGatewayModelPricingQueryer,
	pricingContext GatewayModelPricingContext,
	statement string,
	recordVersion *int64,
) (GatewayModelPricingPolicy, bool, error) {
	arguments := []any{
		pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
		pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
	}
	if recordVersion != nil {
		arguments = append(arguments, *recordVersion)
	}
	var payload []byte
	err := query.QueryRow(pricingContext.RequestContext, statement, arguments...).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return GatewayModelPricingPolicy{}, false, nil
	}
	if err != nil {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	var policy GatewayModelPricingPolicy
	if json.Unmarshal(payload, &policy) != nil || !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	if recordVersion != nil && policy.RecordVersion != *recordVersion {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	return policy, true, nil
}

func lockPostgresGatewayModelPricingScope(
	transaction pgx.Tx,
	pricingContext GatewayModelPricingContext,
) error {
	digest := sha256.Sum256([]byte(gatewayModelPricingScopeKey(pricingContext)))
	lockKey := int64(binary.BigEndian.Uint64(digest[:8]))
	if _, err := transaction.Exec(pricingContext.RequestContext, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
		return errGatewayModelPricingStoreUnavailable
	}
	return nil
}

var _ GatewayModelPricingRepository = (*postgresGatewayModelPricingRepository)(nil)
