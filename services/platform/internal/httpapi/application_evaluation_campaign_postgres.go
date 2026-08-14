package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationEvaluationRepository struct{ pool *pgxpool.Pool }

func newPostgresApplicationEvaluationRepository(pool *pgxpool.Pool) *postgresApplicationEvaluationRepository {
	return &postgresApplicationEvaluationRepository{pool: pool}
}

func (repository *postgresApplicationEvaluationRepository) CreatePlan(ctx ApplicationEvaluationContext, plan ApplicationEvaluationPlan, version ApplicationEvaluationPlanVersion) error {
	if repository == nil || repository.pool == nil || validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.PlanID != version.PlanID || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest || plan.Name != version.Name ||
		plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return errApplicationEvaluationStoreContract
	}
	planPayload, planAt, versionPayload, versionAt, err := encodePostgresApplicationEvaluationPlan(plan, version)
	if err != nil {
		return errApplicationEvaluationStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	command, err := tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_plans
(tenant_ref,workspace_id,environment,application_id,plan_id,record_version,latest_plan_version,latest_plan_digest,lifecycle_state,updated_at,sanitized_plan_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID,
		plan.RecordVersion, plan.LatestPlanVersion, plan.LatestPlanDigest, plan.LifecycleState, planAt, planPayload)
	if err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		current, found, readErr := readPostgresApplicationEvaluationPlan(ctx, tx.QueryRow(ctx.RequestContext, postgresApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
		if readErr != nil || !found {
			return errApplicationEvaluationStoreContract
		}
		return applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_plan_versions
(tenant_ref,workspace_id,environment,application_id,plan_id,plan_version,created_at,sanitized_plan_version_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.PlanID, version.PlanVersion, versionAt, versionPayload); err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return errApplicationEvaluationStoreUnavailable
	}
	return nil
}

func (repository *postgresApplicationEvaluationRepository) RevisePlan(ctx ApplicationEvaluationContext, expected int, plan ApplicationEvaluationPlan, version ApplicationEvaluationPlanVersion) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.pool == nil || expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.RecordVersion != expected+1 || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest || plan.PlanID != version.PlanID ||
		plan.Name != version.Name || plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	current, found, err := readPostgresApplicationEvaluationPlan(ctx, tx.QueryRow(ctx.RequestContext, postgresApplicationEvaluationPlanReadSQL+` FOR UPDATE`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
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
	planPayload, planAt, versionPayload, versionAt, err := encodePostgresApplicationEvaluationPlan(plan, version)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_plan_versions
(tenant_ref,workspace_id,environment,application_id,plan_id,plan_version,created_at,sanitized_plan_version_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.PlanID, version.PlanVersion, versionAt, versionPayload); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_plans SET record_version=$1,latest_plan_version=$2,latest_plan_digest=$3,lifecycle_state=$4,updated_at=$5,sanitized_plan_record=$6
WHERE tenant_ref=$7 AND workspace_id=$8 AND environment=$9 AND application_id=$10 AND plan_id=$11 AND record_version=$12`, plan.RecordVersion, plan.LatestPlanVersion, plan.LatestPlanDigest,
		plan.LifecycleState, planAt, planPayload, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID, expected)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *postgresApplicationEvaluationRepository) ArchivePlan(ctx ApplicationEvaluationContext, expected int, plan ApplicationEvaluationPlan) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.pool == nil || expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || plan.RecordVersion != expected+1 || plan.LifecycleState != applicationEvaluationPlanStateArchived {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	current, found, err := readPostgresApplicationEvaluationPlan(ctx, tx.QueryRow(ctx.RequestContext, postgresApplicationEvaluationPlanReadSQL+` FOR UPDATE`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID))
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
	payload, err := json.Marshal(plan)
	updatedAt, timeErr := applicationEvaluationTimestamp(plan.UpdatedAt)
	if err != nil || timeErr != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_plans SET record_version=$1,lifecycle_state=$2,updated_at=$3,sanitized_plan_record=$4
WHERE tenant_ref=$5 AND workspace_id=$6 AND environment=$7 AND application_id=$8 AND plan_id=$9 AND record_version=$10`, plan.RecordVersion, plan.LifecycleState, updatedAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, plan.PlanID, expected)
	if err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *postgresApplicationEvaluationRepository) ReadPlan(ctx ApplicationEvaluationContext, planID string) (ApplicationEvaluationPlan, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	return readPostgresApplicationEvaluationPlan(ctx, repository.pool.QueryRow(ctx.RequestContext, postgresApplicationEvaluationPlanReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID))
}

func (repository *postgresApplicationEvaluationRepository) ListPlans(ctx ApplicationEvaluationContext, filter ApplicationEvaluationPlanListFilter) (ApplicationEvaluationPlanListPage, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreContract
	}
	before, err := optionalApplicationEvaluationTimestamp(filter.BeforeUpdatedAt)
	if err != nil {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreContract
	}
	rows, err := repository.pool.Query(ctx.RequestContext, `SELECT sanitized_plan_record FROM application_evaluation_plans
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND lifecycle_state=$5
AND ($6::timestamptz IS NULL OR updated_at<$6 OR (updated_at=$6 AND plan_id<$7)) ORDER BY updated_at DESC,plan_id DESC LIMIT $8`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.LifecycleState, before, filter.BeforePlanID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationPlan, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readPostgresApplicationEvaluationPlan(ctx, rows)
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

func (repository *postgresApplicationEvaluationRepository) ReadPlanVersion(ctx ApplicationEvaluationContext, planID string, version int) (ApplicationEvaluationPlanVersion, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) || version < 1 {
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreContract
	}
	return readPostgresApplicationEvaluationPlanVersion(ctx, repository.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_plan_version_record FROM application_evaluation_plan_versions
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND plan_id=$5 AND plan_version=$6`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID, version))
}

func (repository *postgresApplicationEvaluationRepository) ListPlanVersions(ctx ApplicationEvaluationContext, planID string, filter ApplicationEvaluationVersionListFilter) (ApplicationEvaluationVersionListPage, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreContract
	}
	if _, found, err := repository.ReadPlan(ctx, planID); err != nil {
		return ApplicationEvaluationVersionListPage{}, err
	} else if !found {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationNotFound
	}
	rows, err := repository.pool.Query(ctx.RequestContext, `SELECT sanitized_plan_version_record FROM application_evaluation_plan_versions
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND plan_id=$5 AND ($6=0 OR plan_version<$6)
ORDER BY plan_version DESC LIMIT $7`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, planID, filter.BeforeVersion, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationPlanVersion, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readPostgresApplicationEvaluationPlanVersion(ctx, rows)
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

func (repository *postgresApplicationEvaluationRepository) CreateCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.pool == nil || validateApplicationEvaluationCampaign(ctx, campaign) != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	payload, err := json.Marshal(campaign)
	createdAt, timeErr := applicationEvaluationTimestamp(campaign.CreatedAt)
	if err != nil || timeErr != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	existing, found, err := readPostgresApplicationEvaluationCampaign(ctx, tx.QueryRow(ctx.RequestContext, `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND client_campaign_key=$5`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.ClientCampaignKey))
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, err
	}
	if found {
		return replayPostgresApplicationEvaluationCampaign(existing, campaign)
	}
	command, err := tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_campaigns
(tenant_ref,workspace_id,environment,application_id,campaign_id,client_campaign_key,record_version,plan_id,plan_version,plan_digest,quota_api_key_id,campaign_state,created_at,sanitized_campaign_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID,
		campaign.ClientCampaignKey, campaign.RecordVersion, campaign.PlanID, campaign.PlanVersion, campaign.PlanDigest, campaign.QuotaAPIKeyID, campaign.State, createdAt, payload)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		existing, found, err = readPostgresApplicationEvaluationCampaign(ctx, tx.QueryRow(ctx.RequestContext, `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND client_campaign_key=$5`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.ClientCampaignKey))
		if err != nil || !found {
			return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
		}
		return replayPostgresApplicationEvaluationCampaign(existing, campaign)
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *postgresApplicationEvaluationRepository) UpdateCampaign(ctx ApplicationEvaluationContext, expected int, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.pool == nil || expected < 1 || validateApplicationEvaluationCampaign(ctx, campaign) != nil || campaign.RecordVersion != expected+1 {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	current, found, err := readPostgresApplicationEvaluationCampaign(ctx, tx.QueryRow(ctx.RequestContext, postgresApplicationEvaluationCampaignReadSQL+` FOR UPDATE`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID))
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
	payload, err := json.Marshal(campaign)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_campaigns SET record_version=$1,campaign_state=$2,sanitized_campaign_record=$3
WHERE tenant_ref=$4 AND workspace_id=$5 AND environment=$6 AND application_id=$7 AND campaign_id=$8 AND record_version=$9`, campaign.RecordVersion, campaign.State, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaign.CampaignID, expected)
	if err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *postgresApplicationEvaluationRepository) ReadCampaign(ctx ApplicationEvaluationContext, campaignID string) (ApplicationEvaluationCampaign, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	return readPostgresApplicationEvaluationCampaign(ctx, repository.pool.QueryRow(ctx.RequestContext, postgresApplicationEvaluationCampaignReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, campaignID))
}

func (repository *postgresApplicationEvaluationRepository) ListCampaigns(ctx ApplicationEvaluationContext, filter ApplicationEvaluationCampaignListFilter) (ApplicationEvaluationCampaignListPage, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreContract
	}
	before, err := optionalApplicationEvaluationTimestamp(filter.BeforeCreatedAt)
	if err != nil {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreContract
	}
	rows, err := repository.pool.Query(ctx.RequestContext, `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND ($5='' OR plan_id=$5)
AND ($6::timestamptz IS NULL OR created_at<$6 OR (created_at=$6 AND campaign_id<$7)) ORDER BY created_at DESC,campaign_id DESC LIMIT $8`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.PlanID, before, filter.BeforeCampaignID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationCampaign, 0, filter.Limit+1)
	for rows.Next() {
		value, _, scanErr := readPostgresApplicationEvaluationCampaign(ctx, rows)
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

func encodePostgresApplicationEvaluationPlan(plan ApplicationEvaluationPlan, version ApplicationEvaluationPlanVersion) ([]byte, time.Time, []byte, time.Time, error) {
	planPayload, err := json.Marshal(plan)
	if err != nil {
		return nil, time.Time{}, nil, time.Time{}, err
	}
	planAt, err := applicationEvaluationTimestamp(plan.UpdatedAt)
	if err != nil {
		return nil, time.Time{}, nil, time.Time{}, err
	}
	versionPayload, err := json.Marshal(version)
	if err != nil {
		return nil, time.Time{}, nil, time.Time{}, err
	}
	versionAt, err := applicationEvaluationTimestamp(version.CreatedAt)
	return planPayload, planAt, versionPayload, versionAt, err
}

func readPostgresApplicationEvaluationPlan(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationPlan, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func readPostgresApplicationEvaluationPlanVersion(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationPlanVersion, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func readPostgresApplicationEvaluationCampaign(ctx ApplicationEvaluationContext, row applicationEvaluationSQLScanner) (ApplicationEvaluationCampaign, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func replayPostgresApplicationEvaluationCampaign(existing, requested ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if existing.PlanID != requested.PlanID || existing.PlanVersion != requested.PlanVersion || existing.PlanDigest != requested.PlanDigest || existing.QuotaAPIKeyID != requested.QuotaAPIKeyID {
		return existing, false, errApplicationEvaluationCampaignConflict
	}
	return existing, false, nil
}

func applicationEvaluationTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed.UTC(), err
}

func optionalApplicationEvaluationTimestamp(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	return applicationEvaluationTimestamp(value)
}

const postgresApplicationEvaluationPlanReadSQL = `SELECT sanitized_plan_record FROM application_evaluation_plans
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND plan_id=$5`
const postgresApplicationEvaluationCampaignReadSQL = `SELECT sanitized_campaign_record FROM application_evaluation_campaigns
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND campaign_id=$5`

var _ applicationEvaluationRepository = (*postgresApplicationEvaluationRepository)(nil)
