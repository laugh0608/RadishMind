package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteGatewayModelPricingRepository struct {
	database *sql.DB
}

type sqliteGatewayModelPricingQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func newSQLiteGatewayModelPricingRepository(database *sql.DB) *sqliteGatewayModelPricingRepository {
	return &sqliteGatewayModelPricingRepository{database: database}
}

func (repository *sqliteGatewayModelPricingRepository) ReadCurrent(
	pricingContext GatewayModelPricingContext,
) (GatewayModelPricingPolicy, bool, error) {
	if repository.database == nil || !validGatewayModelPricingContext(pricingContext) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	return readSQLiteGatewayModelPricingCurrent(repository.database, pricingContext)
}

func (repository *sqliteGatewayModelPricingRepository) ReadRevision(
	pricingContext GatewayModelPricingContext,
	recordVersion int64,
) (GatewayModelPricingPolicy, bool, error) {
	if repository.database == nil || !validGatewayModelPricingContext(pricingContext) || recordVersion < 1 {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	return readSQLiteGatewayModelPricingRevision(repository.database, pricingContext, recordVersion)
}

func (repository *sqliteGatewayModelPricingRepository) PutRevision(
	pricingContext GatewayModelPricingContext,
	input GatewayModelPricingPolicyInput,
	now time.Time,
) (GatewayModelPricingPolicy, error) {
	input = normalizeGatewayModelPricingPolicyInput(input)
	if repository.database == nil || !validGatewayModelPricingContext(pricingContext) ||
		!validGatewayModelPricingPolicyInput(input) || now.IsZero() {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	connection, err := beginImmediateSQLiteGatewayModelPricingTransaction(repository.database, pricingContext.RequestContext)
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	defer connection.Close()
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	current, found, err := readSQLiteGatewayModelPricingCurrent(connection, pricingContext)
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
	result, err := connection.ExecContext(pricingContext.RequestContext, `INSERT INTO gateway_model_pricing_revisions
        (tenant_ref,workspace_id,environment,provider_id,profile_id,model_id,record_version,policy_id,
         policy_digest,sanitized_policy_record,updated_at_unix_nano)
        VALUES (?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
		pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
		pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
		policy.RecordVersion, policy.PolicyID, policy.PolicyDigest, string(payload), policy.UpdatedAt.UnixNano())
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	} else if affected != 1 {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	if !found {
		result, err = connection.ExecContext(pricingContext.RequestContext, `INSERT INTO gateway_model_pricing_current
            (tenant_ref,workspace_id,environment,provider_id,profile_id,model_id,policy_id,current_version,updated_at_unix_nano)
            VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`,
			pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
			pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
			policy.PolicyID, policy.RecordVersion, policy.UpdatedAt.UnixNano())
	} else {
		result, err = connection.ExecContext(pricingContext.RequestContext, `UPDATE gateway_model_pricing_current SET
            current_version=?,updated_at_unix_nano=?
            WHERE tenant_ref=? AND workspace_id=? AND environment=? AND provider_id=? AND profile_id=? AND model_id=?
              AND policy_id=? AND current_version=?`,
			policy.RecordVersion, policy.UpdatedAt.UnixNano(),
			pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
			pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
			policy.PolicyID, currentVersion)
	}
	if err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	if affected, resultErr := result.RowsAffected(); resultErr != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	} else if affected != 1 {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	if _, err = connection.ExecContext(pricingContext.RequestContext, "COMMIT"); err != nil {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingStoreUnavailable
	}
	committed = true
	return policy, nil
}

func beginImmediateSQLiteGatewayModelPricingTransaction(database *sql.DB, ctx context.Context) (*sql.Conn, error) {
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

func readSQLiteGatewayModelPricingCurrent(
	query sqliteGatewayModelPricingQueryer,
	pricingContext GatewayModelPricingContext,
) (GatewayModelPricingPolicy, bool, error) {
	return readSQLiteGatewayModelPricingPolicy(query, pricingContext, `SELECT r.sanitized_policy_record
        FROM gateway_model_pricing_current c
        JOIN gateway_model_pricing_revisions r ON
          r.tenant_ref=c.tenant_ref AND r.workspace_id=c.workspace_id AND r.environment=c.environment AND
          r.provider_id=c.provider_id AND r.profile_id=c.profile_id AND r.model_id=c.model_id AND
          r.record_version=c.current_version
        WHERE c.tenant_ref=? AND c.workspace_id=? AND c.environment=? AND
          c.provider_id=? AND c.profile_id=? AND c.model_id=?`, nil)
}

func readSQLiteGatewayModelPricingRevision(
	query sqliteGatewayModelPricingQueryer,
	pricingContext GatewayModelPricingContext,
	recordVersion int64,
) (GatewayModelPricingPolicy, bool, error) {
	return readSQLiteGatewayModelPricingPolicy(query, pricingContext, `SELECT sanitized_policy_record
        FROM gateway_model_pricing_revisions
        WHERE tenant_ref=? AND workspace_id=? AND environment=? AND provider_id=? AND profile_id=? AND model_id=?
          AND record_version=?`, &recordVersion)
}

func readSQLiteGatewayModelPricingPolicy(
	query sqliteGatewayModelPricingQueryer,
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
	var payload string
	err := query.QueryRowContext(pricingContext.RequestContext, statement, arguments...).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return GatewayModelPricingPolicy{}, false, nil
	}
	if err != nil {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	var policy GatewayModelPricingPolicy
	if json.Unmarshal([]byte(payload), &policy) != nil || !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	if recordVersion != nil && policy.RecordVersion != *recordVersion {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	return policy, true, nil
}

var _ GatewayModelPricingRepository = (*sqliteGatewayModelPricingRepository)(nil)
