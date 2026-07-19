package httpapi

import (
	"context"
	"testing"
	"time"
)

func TestExactApplicationInteractionAuthorityResolverWorkflowDefinition(t *testing.T) {
	definitionService, runContext, request, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
	ctx := ApplicationInteractionContext{RequestContext: context.Background(), RequestID: "request_session_authority", TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: "audit_session_authority"}
	resolver := newExactApplicationInteractionAuthorityResolver(definitionService.applications, definitionService.repository, nil, workflowRAGApplicationAuthorityResolver{})

	authority, failure := resolver.Resolve(ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflow, DefinitionID: request.DefinitionID})
	if failure != "" || authority.WorkflowDefinition == nil || authority.ApplicationRAG != nil || authority.WorkflowDefinition.DefinitionVersion != request.ExpectedDefinitionVersion || authority.WorkflowDefinition.DefinitionDigest != request.ExpectedDefinitionDigest || authority.WorkflowDefinition.ActivationPointerVersion != request.ExpectedPointerVersion || !workflowRAGDigestPattern.MatchString(authority.AuthorityDigest) {
		t.Fatalf("resolve exact workflow definition authority: authority=%#v failure=%s", authority, failure)
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("authority resolution called provider: %d", bridgeClient.callCount())
	}

	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: ctx.RequestContext, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	if _, err := definitionService.repository.DecideActivation(releaseContext, request.DefinitionID, request.ExpectedPointerVersion, "deactivate", 0, "disable session execution authority", time.Now().UTC()); err != nil {
		t.Fatalf("deactivate definition authority: %v", err)
	}
	if _, failure = resolver.Resolve(ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflow, DefinitionID: request.DefinitionID}); failure != ApplicationInteractionFailureAuthorityMissing {
		t.Fatalf("inactive definition authority did not fail closed: %s", failure)
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("failed authority resolution called provider: %d", bridgeClient.callCount())
	}
}

func TestExactApplicationInteractionAuthorityResolverApplicationRAG(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "activate exact application session authority"})
	if activated.FailureCode != "" || activated.Assignment == nil {
		t.Fatalf("activate RAG authority: %#v", activated)
	}
	applications := newMemoryApplicationCatalogRepository()
	seedWorkflowRAGApplicationRuntimeCatalogRecord(t, applications, fixture)
	ctx := ApplicationInteractionContext{RequestContext: context.Background(), RequestID: "request_session_rag_authority", TenantRef: fixture.runtimeContext.TenantRef, WorkspaceID: fixture.runtimeContext.WorkspaceID, ApplicationID: fixture.runtimeContext.ApplicationID, ActorRef: fixture.runtimeContext.ActorRef, OwnerSubjectRef: fixture.runtimeContext.OwnerSubjectRef, AuditRef: "audit_session_rag_authority"}
	resolver := newExactApplicationInteractionAuthorityResolver(applications, nil, fixture.runtimeRepository, fixture.resolver)

	authority, failure := resolver.Resolve(ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileRAG})
	if failure != "" || authority.ApplicationRAG == nil || authority.WorkflowDefinition != nil || authority.ApplicationRAG.AssignmentID != activated.Assignment.AssignmentID || authority.ApplicationRAG.AssignmentDigest != activated.Assignment.AssignmentDigest || authority.ApplicationRAG.SnapshotID == "" || !workflowRAGDigestPattern.MatchString(authority.AuthorityDigest) {
		t.Fatalf("resolve exact application RAG authority: authority=%#v failure=%s", authority, failure)
	}
	if fixture.bridge.callCount() != 0 {
		t.Fatalf("RAG authority resolution called provider: %d", fixture.bridge.callCount())
	}

	fixture.advanceRuntimeRequest("revoke_for_session")
	revoked := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: activated.Assignment.RecordVersion, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "revoke exact application session authority"})
	if revoked.FailureCode != "" {
		t.Fatalf("revoke RAG authority: %#v", revoked)
	}
	if _, failure = resolver.Resolve(ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileRAG}); failure != ApplicationInteractionFailureAuthorityMissing {
		t.Fatalf("revoked RAG authority did not fail closed: %s", failure)
	}
	if fixture.bridge.callCount() != 0 {
		t.Fatalf("revoked RAG authority resolution called provider: %d", fixture.bridge.callCount())
	}
}
