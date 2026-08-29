package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const (
	workflowTemplateProductSourceCandidate   = "candidate_template_definition_product"
	workflowTemplateProductDerivedDraft      = "draft_from_template_product"
	workflowTemplateProductDerivedCandidate  = "candidate_template_derived_product"
	workflowTemplateProductDerivedDefinition = "definition_template_derived_product"
)

type workflowTemplateProductEvidence struct {
	SourceDefinitionDigest string
	TemplateDigest         string
	DerivedDraftVersion    int
	ProviderBridge         *fakeBridge
}

func TestSQLiteWorkflowTemplateConfiguredProductChainRestartAndNoExecution(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-template-product.db")
	cfg := workflowTemplateProductConfig(aggregateSQLiteDevServerConfig(databasePath))
	first, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-template-product-first"})
	if err != nil {
		t.Fatalf("start SQLite workflow template product server: %v", err)
	}
	evidence := exerciseWorkflowTemplateProductHTTPChain(t, first)
	assertWorkflowTemplateProductSideEffects(t, first, evidence.ProviderBridge)
	assertSQLiteWorkflowTemplateProductRows(t, first)

	closedRepository := first.workflowTemplateCatalogRepository
	first.Close()
	if _, err := closedRepository.ReadLineage(workflowTemplateProductContext(), workflowTemplateTestTemplate); !errorsIsWorkflowTemplateStoreUnavailable(err) {
		t.Fatalf("closed SQLite template repository fell back: %v", err)
	}

	restarted, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-workflow-template-product-restarted"})
	if err != nil {
		t.Fatalf("restart SQLite workflow template product server: %v", err)
	}
	defer restarted.Close()
	attachWorkflowTemplateProductInventory(restarted)
	assertWorkflowTemplateProductRestored(t, restarted, evidence)
}

func workflowTemplateProductConfig(cfg config.Config) config.Config {
	cfg.WorkflowDefinitionReleaseDevEnabled = true
	cfg.WorkflowTemplateCatalogDevEnabled = true
	cfg.AdminProviderRouteDevHTTPEnabled = true
	cfg.AdminProviderRouteDevWriteEnabled = true
	cfg.GatewayAuthMode = gatewayAPIKeyAuthenticationSource
	cfg.GatewayProviderRouteSource = "admin_snapshot_dev_test"
	cfg.GatewayProviderRouteEnvironment = "test"
	cfg.GatewayProviderRouteConfigurationID = "gateway-default"
	return cfg
}

func exerciseWorkflowTemplateProductHTTPChain(t *testing.T, server *Server) workflowTemplateProductEvidence {
	t.Helper()
	seedWorkflowTemplateProductApplications(t, server)
	providerBridge := seedWorkflowTemplateProviderAuthority(t, server)

	sourceDraft := portableWorkflowTemplateDraftPayload()
	sourceDraft.Name = "Template product source draft"
	savedSource := performWorkflowTemplateProductDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-drafts",
		savedWorkflowDraftSaveHTTPBody{Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(sourceDraft)}, workflowTemplateTestSourceApplication)
	if savedSource.FailureCode != nil || savedSource.Draft == nil || savedSource.CurrentDraftVersion != 1 || !savedSource.Draft.ValidationSummary.ValidForReview {
		t.Fatalf("save exact template source draft: %#v", savedSource)
	}

	createdDefinition := performWorkflowTemplateProductDefinitionRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-definition-candidates", workflowDefinitionCandidateCreateBody{
			CandidateID: workflowTemplateProductSourceCandidate, DefinitionID: workflowTemplateTestDefinition,
			DraftID: sourceDraft.DraftID, ExpectedDraftVersion: 1, ExpectedLifecycleVersion: 1,
		}, "workflow_definitions:write", workflowTemplateTestSourceApplication)
	if createdDefinition.FailureCode != nil || createdDefinition.Candidate == nil {
		t.Fatalf("create product source Definition candidate: %#v", createdDefinition)
	}
	approvedDefinition := performWorkflowTemplateProductDefinitionRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-definition-candidates/"+workflowTemplateProductSourceCandidate+"/decisions",
		workflowDefinitionCandidateDecisionBody{ExpectedReviewVersion: 0, Decision: "approve", Reason: "Approve exact product template source."},
		"workflow_definitions:review", workflowTemplateTestSourceApplication)
	if approvedDefinition.FailureCode != nil || approvedDefinition.Version == nil || approvedDefinition.Version.Version != 1 {
		t.Fatalf("approve product source Definition: %#v", approvedDefinition)
	}

	createdTemplate := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-template-candidates",
		workflowTemplateCandidateCreateBodyFromInput(workflowTemplateCandidateTestInput()),
		"workflow_definitions:read,workflow_definitions:write")
	if createdTemplate.FailureCode != nil || createdTemplate.Candidate == nil ||
		createdTemplate.Candidate.SourceDefinitionDigest != approvedDefinition.Version.DefinitionDigest {
		t.Fatalf("create product template candidate from exact Definition: %#v", createdTemplate)
	}
	reviewedTemplate := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-template-candidates/"+workflowTemplateTestCandidate+"/decisions",
		workflowTemplateCandidateDecisionBody{ExpectedReviewVersion: 0, Decision: "approve", Reason: "Approve exact workspace template."},
		"workflow_definitions:read,workflow_definitions:review")
	if reviewedTemplate.FailureCode != nil || reviewedTemplate.Version == nil || reviewedTemplate.Version.Version != 1 {
		t.Fatalf("approve product template candidate: %#v", reviewedTemplate)
	}
	staleReview := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-template-candidates/"+workflowTemplateTestCandidate+"/decisions",
		workflowTemplateCandidateDecisionBody{ExpectedReviewVersion: 0, Decision: "reject", Reason: "Reject from a stale review surface."},
		"workflow_definitions:read,workflow_definitions:review")
	if staleReview.FailureCode == nil || staleReview.CurrentReviewVersion != 1 {
		t.Fatalf("stale product review did not preserve current authority: %#v", staleReview)
	}

	listedTemplate := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/listing-decisions",
		workflowTemplateListingDecisionBody{ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "List exact workspace template version."},
		"workflow_definitions:read,workflow_definitions:activate")
	if listedTemplate.FailureCode != nil || listedTemplate.Lineage == nil || listedTemplate.Lineage.PointerVersion != 1 {
		t.Fatalf("list product template version: %#v", listedTemplate)
	}
	staleListing := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/listing-decisions",
		workflowTemplateListingDecisionBody{ExpectedPointerVersion: 0, Decision: "unlist", Version: 0, Reason: "Unlist from a stale listing surface."},
		"workflow_definitions:read,workflow_definitions:activate")
	if staleListing.FailureCode == nil || staleListing.CurrentPointerVersion != 1 {
		t.Fatalf("stale product listing did not preserve current authority: %#v", staleListing)
	}

	derived := performWorkflowTemplateRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/derivations",
		workflowTemplateDerivationBody{ExpectedPointerVersion: 1, TemplateVersion: 1,
			TargetApplicationID: workflowTemplateTestTargetApplication, DraftID: workflowTemplateProductDerivedDraft,
			Name: "Template product derived draft", Confirmed: true},
		"workflow_definitions:read,workflow_drafts:write")
	if derived.FailureCode != nil || derived.Draft == nil || derived.Draft.ProvenanceKind != string(SavedWorkflowDraftProvenanceTemplate) {
		t.Fatalf("derive product Saved Draft: %#v", derived)
	}
	derivedRecord := savedWorkflowDraftFromDocument(*derived.Draft)
	derivedPayload := savedWorkflowDraftPayloadFromDraft(derivedRecord)
	validated := performWorkflowTemplateProductDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-drafts/validate",
		savedWorkflowDraftValidateHTTPBody{Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(derivedPayload)}, workflowTemplateTestTargetApplication)
	if validated.FailureCode != nil || !validated.ValidationSummary.ValidForReview {
		t.Fatalf("validate product template-derived draft: %#v", validated)
	}
	resaved := performWorkflowTemplateProductDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-drafts",
		savedWorkflowDraftSaveHTTPBody{ExpectedDraftVersion: 1, ExpectedLifecycleVersion: 1,
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(derivedPayload)},
		workflowTemplateTestTargetApplication)
	if resaved.FailureCode != nil || resaved.Draft == nil || resaved.CurrentDraftVersion != 2 ||
		resaved.Draft.ProvenanceKind != string(SavedWorkflowDraftProvenanceTemplate) {
		t.Fatalf("resave exact template derivation: %#v", resaved)
	}

	derivedDefinition := performWorkflowTemplateProductDefinitionRequest(t, server, http.MethodPost,
		"/v1/user-workspace/workflow-definition-candidates", workflowDefinitionCandidateCreateBody{
			CandidateID: workflowTemplateProductDerivedCandidate, DefinitionID: workflowTemplateProductDerivedDefinition,
			DraftID: workflowTemplateProductDerivedDraft, ExpectedDraftVersion: 2, ExpectedLifecycleVersion: 1,
		}, "workflow_definitions:write", workflowTemplateTestTargetApplication)
	if derivedDefinition.FailureCode != nil || derivedDefinition.Candidate == nil ||
		derivedDefinition.Candidate.SourceDraftVersion != 2 {
		t.Fatalf("handoff template-derived draft to Definition candidate: %#v", derivedDefinition)
	}

	return workflowTemplateProductEvidence{
		SourceDefinitionDigest: approvedDefinition.Version.DefinitionDigest,
		TemplateDigest:         reviewedTemplate.Version.TemplateDigest,
		DerivedDraftVersion:    2,
		ProviderBridge:         providerBridge,
	}
}

func seedWorkflowTemplateProductApplications(t *testing.T, server *Server) {
	t.Helper()
	service := newApplicationCatalogService(server.applicationCatalogRepository)
	applicationIDs := []string{workflowTemplateTestSourceApplication, workflowTemplateTestTargetApplication}
	nextID := 0
	service.newID = func() (string, error) {
		identifier := applicationIDs[nextID]
		nextID++
		return identifier, nil
	}
	ctx := applicationCatalogTestContext("subject_demo_user")
	for index, name := range []string{"Template source application", "Template target application"} {
		created := service.Create(ctx, ApplicationCatalogCreateInput{
			DisplayName: name, Description: "Workflow template product continuity authority.", ApplicationKind: "workflow_copilot",
		})
		if created.FailureCode != "" || created.Record == nil || created.Record.ApplicationID != applicationIDs[index] {
			t.Fatalf("create workflow template product application %d: %#v", index, created)
		}
	}
}

func attachWorkflowTemplateProductInventory(server *Server) *fakeBridge {
	profile := seedWorkflowTemplateProviderAuthorityProfile()
	productBridge := &fakeBridge{inventory: profile}
	server.bridge = productBridge
	server.workflowTemplateTargetBindingValidator = configuredWorkflowTemplateTargetBindingValidator{
		providerRouteSource: "admin_snapshot_dev_test", providerRouteEnvironment: "test", providerRouteConfigID: "gateway-default",
		snapshotProvider: server.providerRouteSnapshotProvider, bridge: productBridge,
	}
	return productBridge
}

func seedWorkflowTemplateProviderAuthorityProfile() bridge.ProviderInventory {
	return bridge.ProviderInventory{Profiles: []bridge.ProviderProfileDescription{{
		Profile: "radishmind-default-workflow", NormalizedProfile: "radishmind-default-workflow", ProviderID: "mock",
		ResolvedModel: "mock-workflow-model", APIStyle: "openai_compatible", HasBaseURL: true, HasAPIKey: true,
		RequestTimeoutSeconds: 30, Active: true, Enabled: true,
		Capabilities: map[string]any{"chat": true, "responses": true}, NorthboundProtocols: []string{"chat.completions", "responses"},
		NorthboundRoutes: []string{"/v1/chat/completions", "/v1/responses"}, CredentialState: "configured",
		DeploymentMode: "local", AuthMode: "test", Streaming: true,
	}}}
}

func assertWorkflowTemplateProductSideEffects(t *testing.T, server *Server, productBridge *fakeBridge) {
	t.Helper()
	if productBridge.handleCalls != 0 || productBridge.streamCalled {
		t.Fatalf("template product actions reached Provider execution: handle=%d stream=%t", productBridge.handleCalls, productBridge.streamCalled)
	}
	if sideEffects := server.savedWorkflowDraftStore.SideEffects(); sideEffects.DraftWriteCount != 3 ||
		sideEffects.ExternalRepositoryWrites != 3 || sideEffects.ExecutorCallCount != 0 ||
		sideEffects.ConfirmationCallCount != 0 || sideEffects.BusinessWritebackCount != 0 ||
		sideEffects.ReplayCallCount != 0 || sideEffects.MaterializedResultReads != 0 {
		t.Fatalf("template product actions crossed Saved Draft runtime boundary: %#v", sideEffects)
	}
}

func assertSQLiteWorkflowTemplateProductRows(t *testing.T, server *Server) {
	t.Helper()
	database := server.localPersistenceRuntime.DB()
	for table, expected := range map[string]int{
		"workflow_template_candidates": 1, "workflow_template_decisions": 1, "workflow_template_versions": 1,
		"workflow_template_lineages": 1, "workflow_template_listing_events": 1, "workflow_template_audits": 4,
		"workflow_run_records": 0, "workflow_http_tool_confirmation_decisions": 0, "workflow_rag_execution_audits": 0,
	} {
		var count int
		if err := database.QueryRowContext(context.Background(), "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != expected {
			t.Fatalf("unexpected SQLite product row count: table=%s count=%d expected=%d err=%v", table, count, expected, err)
		}
	}
	if _, err := database.ExecContext(context.Background(), "UPDATE workflow_template_versions SET title=title WHERE template_id=?", workflowTemplateTestTemplate); err == nil {
		t.Fatal("SQLite product runtime mutated immutable template history")
	}
}

func assertWorkflowTemplateProductRestored(t *testing.T, server *Server, evidence workflowTemplateProductEvidence) {
	t.Helper()
	ctx := workflowTemplateProductContext()
	lineage, err := server.workflowTemplateCatalogRepository.ReadLineage(ctx, workflowTemplateTestTemplate)
	if err != nil || lineage.PointerVersion != 1 || lineage.ListedDigest != evidence.TemplateDigest {
		t.Fatalf("restore exact product template listing: lineage=%#v err=%v", lineage, err)
	}
	version, err := server.workflowTemplateCatalogRepository.ReadVersion(ctx, workflowTemplateTestTemplate, 1)
	if err != nil || version.TemplateDigest != evidence.TemplateDigest || version.SourceDefinitionDigest != evidence.SourceDefinitionDigest {
		t.Fatalf("restore immutable product template version: version=%#v err=%v", version, err)
	}
	draft := newSavedWorkflowDraftService(server.savedWorkflowDraftStore).ReadDraft(SavedWorkflowDraftContext{
		RequestContext: context.Background(), RequestID: "request_template_product_restore", TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: workflowTemplateTestTargetApplication, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, ScopeGrants: []string{"workflow_drafts:read"}, AuditRef: "audit_template_product_restore",
	}, ReadWorkflowDraftRequest{DraftID: workflowTemplateProductDerivedDraft})
	if draft.FailureCode != "" || draft.Draft == nil || draft.Draft.DraftVersion != evidence.DerivedDraftVersion ||
		draft.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("restore exact template-derived Saved Draft: %#v", draft)
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: workflowTemplateTestTargetApplication, OwnerSubjectRef: ctx.OwnerSubjectRef,
		ActorRef: ctx.ActorRef, RequestID: "request_template_product_restore", AuditRef: "audit_template_product_restore"}
	candidate, err := server.workflowDefinitionReleaseRepository.ReadCandidate(releaseContext, workflowTemplateProductDerivedCandidate)
	if err != nil || candidate.SourceDraftVersion != evidence.DerivedDraftVersion || candidate.State != workflowDefinitionStatePending {
		t.Fatalf("restore derived Definition candidate handoff: candidate=%#v err=%v", candidate, err)
	}
}

func workflowTemplateProductContext() WorkflowTemplateCatalogContext {
	return WorkflowTemplateCatalogContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		OwnerSubjectRef: "subject_demo_user", ActorRef: "subject_demo_user", RequestID: "request_template_product", AuditRef: "audit_template_product"}
}

func performWorkflowTemplateProductDraftRequest(t *testing.T, server *Server, method, target string, body any, applicationID string) savedWorkflowDraftEnvelope {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewReader(mustSavedWorkflowDraftJSON(t, body)))
	request.Header.Set("Content-Type", "application/json")
	setLocalProductWorkflowHeaders(request, "workflow_drafts:read,workflow_drafts:write", applicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	return decodeSavedWorkflowDraftEnvelope(t, recorder, http.StatusOK)
}

func performWorkflowTemplateProductDefinitionRequest(t *testing.T, server *Server, method, target string, body any, scopes, applicationID string) workflowDefinitionReleaseEnvelope {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(payload))
	setWorkflowDefinitionHeaders(request, scopes)
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected Definition product status %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope workflowDefinitionReleaseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode Definition product response: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}
