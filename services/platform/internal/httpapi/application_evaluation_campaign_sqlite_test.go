package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteApplicationEvaluationPlanAndCampaignSurviveRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application-evaluation.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}

	baseService, evaluationContext := newApplicationEvaluationPlanTestService(t, "workflow_copilot")
	repository := newSQLiteApplicationEvaluationRepository(runtime.DB())
	baseService.repository = repository
	baseService.newPlanID = func() (string, error) { return "aeplan_aaaaaaaaaaaaaaaa", nil }
	created := baseService.Create(evaluationContext, applicationEvaluationWorkflowPlanInput("Definition campaign"))
	if created.FailureCode != "" || created.Plan == nil || created.Version == nil {
		t.Fatalf("create SQLite evaluation plan: %+v", created)
	}
	revised := baseService.Revise(evaluationContext, created.Plan.PlanID, ApplicationEvaluationPlanReviseInput{
		ExpectedVersion: 1, Name: "Definition campaign revised", ExecutionProfile: workflowDefinitionExecutorProfile,
		Target: created.Version.Target, Items: append(created.Version.Items, ApplicationEvaluationPlanItem{
			ItemKey: "second", Name: "Second definition input", ExpectedClassification: WorkflowRunComparisonUnchanged,
			WorkflowDefinition: &ApplicationEvaluationDefinitionFixture{InputText: "second durable input", ConditionValues: map[string]bool{}},
		}),
	})
	if revised.FailureCode != "" || revised.Plan == nil || revised.Version == nil || revised.Version.PlanVersion != 2 {
		t.Fatalf("revise SQLite evaluation plan: %+v", revised)
	}

	campaign := applicationEvaluationPendingCampaign(evaluationContext, "campaign-durable", "2026-08-09T08:00:00Z")
	campaign.PlanID = revised.Plan.PlanID
	campaign.PlanVersion = revised.Version.PlanVersion
	campaign.PlanDigest = revised.Version.PlanDigest
	createdCampaign, inserted, err := repository.CreateCampaign(evaluationContext, campaign)
	if err != nil || !inserted {
		t.Fatalf("create SQLite campaign: inserted=%v err=%v value=%+v", inserted, err, createdCampaign)
	}
	running := campaign
	running.RecordVersion = 2
	running.State = applicationEvaluationCampaignStateRunning
	running.StartedAt = "2026-08-09T08:00:01Z"
	authority := applicationEvaluationWorkflowAuthority(t, evaluationContext)
	running.Authority = &authority
	storedCampaign, updated, err := repository.UpdateCampaign(evaluationContext, 1, running)
	if err != nil || !updated || storedCampaign.State != applicationEvaluationCampaignStateRunning {
		t.Fatalf("checkpoint SQLite campaign: updated=%v err=%v value=%+v", updated, err, storedCampaign)
	}

	if err = runtime.Close(); err != nil {
		t.Fatal(err)
	}
	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	restartedRepository := newSQLiteApplicationEvaluationRepository(restarted.DB())
	readPlan, found, err := restartedRepository.ReadPlan(evaluationContext, revised.Plan.PlanID)
	if err != nil || !found || readPlan.RecordVersion != 2 || readPlan.LatestPlanVersion != 2 {
		t.Fatalf("read restarted plan: found=%v err=%v value=%+v", found, err, readPlan)
	}
	firstVersion, found, err := restartedRepository.ReadPlanVersion(evaluationContext, revised.Plan.PlanID, 1)
	if err != nil || !found || firstVersion.Name != "Definition campaign" {
		t.Fatalf("read immutable version after restart: found=%v err=%v value=%+v", found, err, firstVersion)
	}
	readCampaign, found, err := restartedRepository.ReadCampaign(evaluationContext, campaign.CampaignID)
	if err != nil || !found || readCampaign.State != applicationEvaluationCampaignStateRunning || readCampaign.RecordVersion != 2 {
		t.Fatalf("read restarted campaign: found=%v err=%v value=%+v", found, err, readCampaign)
	}
	finished := readCampaign
	finished.RecordVersion = 3
	finished.State = applicationEvaluationCampaignStateSucceeded
	finished.SucceededItems = 1
	finished.CompletedAt = "2026-08-09T08:00:02Z"
	finished.Items[0].State = applicationEvaluationCampaignItemSucceeded
	finished.Items[0].StartedAt = finished.StartedAt
	finished.Items[0].CompletedAt = finished.CompletedAt
	finished.Items[0].RunSchemaVersion = workflowRunRecordDefinitionSchemaVersion
	finished.Items[0].RunProfile = workflowDefinitionExecutorProfile
	finished.Items[0].AuthorityDigest = authority.AuthorityDigest
	finished, updated, err = restartedRepository.UpdateCampaign(evaluationContext, 2, finished)
	if err != nil || !updated {
		t.Fatalf("finish restarted campaign: updated=%v err=%v value=%+v", updated, err, finished)
	}
	withHandoff := finished
	withHandoff.RecordVersion = 4
	withHandoff.Handoff = &ApplicationEvaluationHandoffRef{
		BaselineCampaignID: "aecamp_bbbbbbbbbbbbbbbb", CandidateCampaignID: withHandoff.CampaignID,
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: "eval_sqlite_case", Version: 1}}, State: "partial", AuditRef: "audit-sqlite-handoff",
	}
	withHandoff, updated, err = restartedRepository.UpdateCampaign(evaluationContext, 3, withHandoff)
	if err != nil || !updated || withHandoff.Handoff == nil || len(withHandoff.Handoff.CaseRefs) != 1 {
		t.Fatalf("checkpoint SQLite handoff: updated=%v err=%v value=%+v", updated, err, withHandoff)
	}
	replayed, inserted, err := restartedRepository.CreateCampaign(evaluationContext, campaign)
	if err != nil || inserted || replayed.CampaignID != campaign.CampaignID || replayed.State != applicationEvaluationCampaignStateSucceeded || replayed.Handoff == nil {
		t.Fatalf("replay durable campaign key: inserted=%v err=%v value=%+v", inserted, err, replayed)
	}
	if _, err = restarted.DB().ExecContext(context.Background(), `UPDATE application_evaluation_plan_versions SET sanitized_plan_version_record=sanitized_plan_version_record WHERE plan_id=?`, revised.Plan.PlanID); err == nil {
		t.Fatal("immutable plan version update must be rejected by SQLite")
	}
	if _, err = restarted.DB().ExecContext(context.Background(), `DELETE FROM application_evaluation_campaigns WHERE campaign_id=?`, campaign.CampaignID); err == nil {
		t.Fatal("campaign deletion must be rejected by SQLite")
	}
}

func TestSQLiteApplicationEvaluationStoredContractCloses(t *testing.T) {
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "application-evaluation-corruption.db"), Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	service, evaluationContext := newApplicationEvaluationPlanTestService(t, "prompt_application")
	repository := newSQLiteApplicationEvaluationRepository(runtime.DB())
	service.repository = repository
	created := service.Create(evaluationContext, applicationEvaluationPromptPlanInput("Stored contract", WorkflowRunComparisonUnchanged))
	if created.FailureCode != "" || created.Plan == nil {
		t.Fatalf("create plan before corruption: %+v", created)
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `PRAGMA ignore_check_constraints=ON`); err != nil {
		t.Fatal(err)
	}
	_, err = runtime.DB().ExecContext(context.Background(), `UPDATE application_evaluation_plans
SET record_version=record_version+1,sanitized_plan_record=json_set(sanitized_plan_record,'$.record_version',record_version+1,'$.unexpected',1)
WHERE plan_id=?`, created.Plan.PlanID)
	if err != nil {
		t.Fatal(err)
	}
	if _, found, readErr := repository.ReadPlan(evaluationContext, created.Plan.PlanID); found || !errors.Is(readErr, errApplicationEvaluationStoreContract) {
		t.Fatalf("corrupted stored record must fail closed: found=%v err=%v", found, readErr)
	}
}

func applicationEvaluationWorkflowPlanInput(name string) ApplicationEvaluationPlanCreateInput {
	digest := "sha256:" + strings.Repeat("b", 64)
	return ApplicationEvaluationPlanCreateInput{
		Name: name, ExecutionProfile: workflowDefinitionExecutorProfile,
		Target: ApplicationEvaluationPlanTarget{WorkflowDefinition: &ApplicationEvaluationDefinitionTarget{
			DefinitionID: "definition-one", ExpectedPointerVersion: 1, ExpectedDefinitionVersion: 1, ExpectedDefinitionDigest: digest,
		}},
		Items: []ApplicationEvaluationPlanItem{{
			ItemKey: "first", Name: "First definition input", ExpectedClassification: WorkflowRunComparisonUnchanged,
			WorkflowDefinition: &ApplicationEvaluationDefinitionFixture{InputText: "first durable input", ConditionValues: map[string]bool{}},
		}},
	}
}
