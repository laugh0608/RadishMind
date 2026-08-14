package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteApplicationEvaluationRepository struct{ database *sql.DB }

func newSQLiteApplicationEvaluationRepository(database *sql.DB) *sqliteApplicationEvaluationRepository {
	return &sqliteApplicationEvaluationRepository{database: database}
}

type applicationEvaluationSQLScanner interface{ Scan(...any) error }

func (repository *sqliteApplicationEvaluationRepository) CreatePlan(ctx ApplicationEvaluationContext, plan ApplicationEvaluationPlan, version ApplicationEvaluationPlanVersion) error {
	if repository == nil || repository.database == nil || validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.PlanID != version.PlanID || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest || plan.Name != version.Name ||
		plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return errApplicationEvaluationStoreContract
	}
	planPayload, planAt, err := encodeApplicationEvaluationPlan(plan)
	if err != nil {
		return errApplicationEvaluationStoreContract
	}
	versionPayload, versionAt, err := encodeApplicationEvaluationPlanVersion(version)
	if err != nil {
		return errApplicationEvaluationStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationPlan(ctx, tx.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
	if err != nil {
		return err
	}
	if found {
		return applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_plans
(tenant_ref,workspace_id,environment,application_id,plan_id,record_version,latest_plan_version,latest_plan_digest,lifecycle_state,updated_at_unix_nano,sanitized_plan_record)
VALUES(?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID, plan.RecordVersion, plan.LatestPlanVersion, plan.LatestPlanDigest, plan.LifecycleState, planAt, planPayload); err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_plan_versions
(tenant_ref,workspace_id,environment,application_id,plan_id,plan_version,created_at_unix_nano,sanitized_plan_version_record)
VALUES(?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.PlanID, version.PlanVersion, versionAt, versionPayload); err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	if err = tx.Commit(); err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	return nil
}

func (repository *sqliteApplicationEvaluationRepository) RevisePlan(ctx ApplicationEvaluationContext, expected int, plan ApplicationEvaluationPlan, version ApplicationEvaluationPlanVersion) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.database == nil || expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.RecordVersion != expected+1 || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest || plan.PlanID != version.PlanID ||
		plan.Name != version.Name || plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationPlan(ctx, tx.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
	if err != nil || !found {
		return ApplicationEvaluationPlan{}, false, err
	}
	if current.RecordVersion != expected {
		return current, false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return current, false, errApplicationEvaluationArchived
	}
	if version.PlanVersion != current.LatestPlanVersion+1 || version.PreviousPlanVersion != current.LatestPlanVersion || plan.CreatedAt != current.CreatedAt ||
		plan.CreatedByActorRef != current.CreatedByActorRef || plan.LifecycleState != applicationEvaluationPlanStateActive {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	planPayload, planAt, err := encodeApplicationEvaluationPlan(plan)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	versionPayload, versionAt, err := encodeApplicationEvaluationPlanVersion(version)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_plan_versions
(tenant_ref,workspace_id,environment,application_id,plan_id,plan_version,created_at_unix_nano,sanitized_plan_version_record)
VALUES(?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.PlanID, version.PlanVersion, versionAt, versionPayload); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_plans SET record_version=?,latest_plan_version=?,latest_plan_digest=?,lifecycle_state=?,updated_at_unix_nano=?,sanitized_plan_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND plan_id=? AND record_version=?`, plan.RecordVersion, plan.LatestPlanVersion, plan.LatestPlanDigest, plan.LifecycleState, planAt, planPayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID, expected)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *sqliteApplicationEvaluationRepository) ArchivePlan(ctx ApplicationEvaluationContext, expected int, plan ApplicationEvaluationPlan) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.database == nil || expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || plan.RecordVersion != expected+1 || plan.LifecycleState != applicationEvaluationPlanStateArchived {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationPlan(ctx, tx.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
	if err != nil || !found {
		return ApplicationEvaluationPlan{}, false, err
	}
	if current.RecordVersion != expected {
		return current, false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return current, false, errApplicationEvaluationArchived
	}
	if plan.LatestPlanVersion != current.LatestPlanVersion || plan.LatestPlanDigest != current.LatestPlanDigest || plan.Name != current.Name ||
		plan.ExecutionProfile != current.ExecutionProfile || plan.ItemCount != current.ItemCount || plan.CreatedAt != current.CreatedAt || plan.CreatedByActorRef != current.CreatedByActorRef {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	payload, updatedAt, err := encodeApplicationEvaluationPlan(plan)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_plans SET record_version=?,lifecycle_state=?,updated_at_unix_nano=?,sanitized_plan_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND plan_id=? AND record_version=?`, plan.RecordVersion, plan.LifecycleState, updatedAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID, expected)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *sqliteApplicationEvaluationRepository) ReadPlan(ctx ApplicationEvaluationContext, planID string) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	return readSQLiteApplicationEvaluationPlan(ctx, repository.database.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID))
}

func (repository *sqliteApplicationEvaluationRepository) ListPlans(ctx ApplicationEvaluationContext, filter ApplicationEvaluationPlanListFilter) (ApplicationEvaluationPlanListPage, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreContract
	}
	before, err := optionalApplicationEvaluationUnixNano(filter.BeforeUpdatedAt)
	if err != nil {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreContract
	}
	rows, err := repository.database.QueryContext(ctx.RequestContext, `SELECT sanitized_plan_record FROM application_evaluation_plans
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND lifecycle_state=?
AND (? IS NULL OR updated_at_unix_nano<? OR (updated_at_unix_nano=? AND plan_id<?))
ORDER BY updated_at_unix_nano DESC,plan_id DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.LifecycleState,
		before, before, before, filter.BeforePlanID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationPlan, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readSQLiteApplicationEvaluationPlan(ctx, rows)
		if scanErr != nil {
			return ApplicationEvaluationPlanListPage{}, scanErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreUnavailable
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationPlanListPage{Plans: values, HasMore: hasMore}, nil
}

func (repository *sqliteApplicationEvaluationRepository) ReadPlanVersion(ctx ApplicationEvaluationContext, planID string, version int) (ApplicationEvaluationPlanVersion, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) || version < 1 {
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreContract
	}
	return readSQLiteApplicationEvaluationPlanVersion(ctx, repository.database.QueryRowContext(ctx.RequestContext, `SELECT sanitized_plan_version_record FROM application_evaluation_plan_versions
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND plan_id=? AND plan_version=?`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID, version))
}

func (repository *sqliteApplicationEvaluationRepository) ListPlanVersions(ctx ApplicationEvaluationContext, planID string, filter ApplicationEvaluationVersionListFilter) (ApplicationEvaluationVersionListPage, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreContract
	}
	if _, found, err := repository.ReadPlan(ctx, planID); err != nil {
		return ApplicationEvaluationVersionListPage{}, err
	} else if !found {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationNotFound
	}
	rows, err := repository.database.QueryContext(ctx.RequestContext, `SELECT sanitized_plan_version_record FROM application_evaluation_plan_versions
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND plan_id=? AND (?=0 OR plan_version<?)
ORDER BY plan_version DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID, filter.BeforeVersion, filter.BeforeVersion, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationPlanVersion, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readSQLiteApplicationEvaluationPlanVersion(ctx, rows)
		if scanErr != nil {
			return ApplicationEvaluationVersionListPage{}, scanErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreUnavailable
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationVersionListPage{Versions: values, HasMore: hasMore}, nil
}

func (repository *sqliteApplicationEvaluationRepository) CreateCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.database == nil || validateApplicationEvaluationCampaign(ctx, campaign) != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	payload, createdAt, err := encodeApplicationEvaluationCampaign(campaign)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	existing, found, err := readSQLiteApplicationEvaluationCampaign(ctx, tx.QueryRowContext(ctx.RequestContext, `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND client_campaign_key=?`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.ClientCampaignKey))
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, err
	}
	if found {
		if existing.PlanID != campaign.PlanID || existing.PlanVersion != campaign.PlanVersion || existing.PlanDigest != campaign.PlanDigest || existing.QuotaAPIKeyID != campaign.QuotaAPIKeyID {
			return existing, false, errApplicationEvaluationCampaignConflict
		}
		return existing, false, nil
	}
	if _, found, err = readSQLiteApplicationEvaluationCampaign(ctx, tx.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationCampaignReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID)); err != nil {
		return ApplicationEvaluationCampaign{}, false, err
	} else if found {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_campaigns
(tenant_ref,workspace_id,environment,application_id,campaign_id,client_campaign_key,record_version,plan_id,plan_version,plan_digest,quota_api_key_id,campaign_state,created_at_unix_nano,sanitized_campaign_record)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID, campaign.ClientCampaignKey, campaign.RecordVersion,
		campaign.PlanID, campaign.PlanVersion, campaign.PlanDigest, campaign.QuotaAPIKeyID, campaign.State, createdAt, payload); err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *sqliteApplicationEvaluationRepository) UpdateCampaign(ctx ApplicationEvaluationContext, expected int, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.database == nil || expected < 1 || validateApplicationEvaluationCampaign(ctx, campaign) != nil || campaign.RecordVersion != expected+1 {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationCampaign(ctx, tx.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationCampaignReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID))
	if err != nil || !found {
		return ApplicationEvaluationCampaign{}, false, err
	}
	if current.RecordVersion != expected {
		return current, false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.State}
	}
	if current.PlanID != campaign.PlanID || current.PlanVersion != campaign.PlanVersion || current.PlanDigest != campaign.PlanDigest || current.ExecutionProfile != campaign.ExecutionProfile || current.QuotaAPIKeyID != campaign.QuotaAPIKeyID ||
		current.ClientCampaignKey != campaign.ClientCampaignKey || current.CreatedAt != campaign.CreatedAt || current.CreatedByActorRef != campaign.CreatedByActorRef || !validApplicationEvaluationCampaignUpdate(current, campaign) {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	payload, _, err := encodeApplicationEvaluationCampaign(campaign)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_campaigns SET record_version=?,campaign_state=?,sanitized_campaign_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND campaign_id=? AND record_version=?`, campaign.RecordVersion, campaign.State, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID, expected)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *sqliteApplicationEvaluationRepository) ReadCampaign(ctx ApplicationEvaluationContext, campaignID string) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	return readSQLiteApplicationEvaluationCampaign(ctx, repository.database.QueryRowContext(ctx.RequestContext, sqliteApplicationEvaluationCampaignReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaignID))
}

func (repository *sqliteApplicationEvaluationRepository) ListCampaigns(ctx ApplicationEvaluationContext, filter ApplicationEvaluationCampaignListFilter) (ApplicationEvaluationCampaignListPage, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreContract
	}
	before, err := optionalApplicationEvaluationUnixNano(filter.BeforeCreatedAt)
	if err != nil {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreContract
	}
	rows, err := repository.database.QueryContext(ctx.RequestContext, `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND (?='' OR plan_id=?)
AND (? IS NULL OR created_at_unix_nano<? OR (created_at_unix_nano=? AND campaign_id<?))
ORDER BY created_at_unix_nano DESC,campaign_id DESC LIMIT ?`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.PlanID, filter.PlanID,
		before, before, before, filter.BeforeCampaignID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationCampaign, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readSQLiteApplicationEvaluationCampaign(ctx, rows)
		if scanErr != nil {
			return ApplicationEvaluationCampaignListPage{}, scanErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreUnavailable
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationCampaignListPage{Campaigns: values, HasMore: hasMore}, nil
}

func readSQLiteApplicationEvaluationPlan(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationPlan, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationPlan{}, false, nil
		}
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	var plan ApplicationEvaluationPlan
	if decodeStrictApplicationEvaluationJSON(payload, &plan) != nil || validateApplicationEvaluationPlan(ctx, plan) != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func readSQLiteApplicationEvaluationPlanVersion(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationPlanVersion, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationPlanVersion{}, false, nil
		}
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreUnavailable
	}
	var version ApplicationEvaluationPlanVersion
	if decodeStrictApplicationEvaluationJSON(payload, &version) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil {
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationPlanVersion(version), true, nil
}

func readSQLiteApplicationEvaluationCampaign(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationCampaign, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationCampaign{}, false, nil
		}
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	var campaign ApplicationEvaluationCampaign
	if decodeStrictApplicationEvaluationJSON(payload, &campaign) != nil || validateApplicationEvaluationCampaign(ctx, campaign) != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func encodeApplicationEvaluationPlan(plan ApplicationEvaluationPlan) ([]byte, int64, error) {
	payload, err := json.Marshal(plan)
	if err != nil {
		return nil, 0, err
	}
	updatedAt, err := applicationEvaluationUnixNano(plan.UpdatedAt)
	return payload, updatedAt, err
}

func encodeApplicationEvaluationPlanVersion(version ApplicationEvaluationPlanVersion) ([]byte, int64, error) {
	payload, err := json.Marshal(version)
	if err != nil {
		return nil, 0, err
	}
	createdAt, err := applicationEvaluationUnixNano(version.CreatedAt)
	return payload, createdAt, err
}

func encodeApplicationEvaluationCampaign(campaign ApplicationEvaluationCampaign) ([]byte, int64, error) {
	payload, err := json.Marshal(campaign)
	if err != nil {
		return nil, 0, err
	}
	createdAt, err := applicationEvaluationUnixNano(campaign.CreatedAt)
	return payload, createdAt, err
}

func applicationEvaluationUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	return parsed.UTC().UnixNano(), nil
}

func optionalApplicationEvaluationUnixNano(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	return applicationEvaluationUnixNano(value)
}

const sqliteApplicationEvaluationPlanReadSQL = `SELECT sanitized_plan_record FROM application_evaluation_plans
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND plan_id=?`
const sqliteApplicationEvaluationCampaignReadSQL = `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND campaign_id=?`

var _ applicationEvaluationRepository = (*sqliteApplicationEvaluationRepository)(nil)
