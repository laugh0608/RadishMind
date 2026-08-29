package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	workflowTemplateListCursorSchemaVersion = "workspace_workflow_template_list_cursor.v1"
	workflowTemplateDefaultListLimit        = 25
	workflowTemplateMaximumListLimit        = 100
)

type WorkflowTemplateCandidateCreateInput struct {
	CandidateID             string
	TemplateID              string
	SourceApplicationID     string
	SourceDefinitionID      string
	SourceDefinitionVersion int
	Title                   string
	Summary                 string
	UsageNotes              string
	Labels                  []string
}

type WorkflowTemplateReviewInput struct {
	ExpectedReviewVersion int
	Decision              string
	Reason                string
}

type WorkflowTemplateListingInput struct {
	ExpectedPointerVersion int
	Decision               string
	Version                int
	Reason                 string
}

type WorkflowTemplateDerivationInput struct {
	ExpectedPointerVersion int
	TemplateVersion        int
	TargetApplicationID    string
	DraftID                string
	Name                   string
	Confirmed              bool
}

type WorkflowTemplateListInput struct {
	State  string
	Limit  int
	Cursor string
}

type WorkflowTemplateCatalogResult struct {
	Candidate             *WorkflowTemplateCandidate
	Version               *WorkflowTemplateVersion
	Lineage               *WorkflowTemplateLineage
	Draft                 *SavedWorkflowDraft
	FailureCode           string
	CurrentReviewVersion  int
	CurrentPointerVersion int
}

type WorkflowTemplateCandidateListResult struct {
	Candidates  []WorkflowTemplateCandidate
	NextCursor  string
	FailureCode string
}

type WorkflowTemplateLineageListResult struct {
	Lineages    []WorkflowTemplateLineage
	NextCursor  string
	FailureCode string
}

type WorkflowTemplateVersionListResult struct {
	Versions    []WorkflowTemplateVersion
	NextCursor  string
	FailureCode string
}

type workflowTemplateTargetBindingValidator interface {
	ValidateTargetBinding(WorkflowTemplateCatalogContext, ApplicationCatalogRecord, WorkflowDefinitionSnapshot) error
}

type strictWorkflowTemplateTargetBindingValidator struct{}

func (strictWorkflowTemplateTargetBindingValidator) ValidateTargetBinding(ctx WorkflowTemplateCatalogContext, application ApplicationCatalogRecord, snapshot WorkflowDefinitionSnapshot) error {
	if !validWorkflowTemplateContext(ctx) || application.TenantRef != ctx.TenantRef || application.WorkspaceID != ctx.WorkspaceID ||
		application.OwnerSubjectRef != ctx.OwnerSubjectRef || application.LifecycleState != applicationCatalogLifecycleActive ||
		application.ApplicationKind != "workflow_copilot" {
		return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
	}
	if !validWorkflowTemplateProviderBindings(snapshot) {
		return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
	}
	return nil
}

type configuredWorkflowTemplateTargetBindingValidator struct {
	providerRouteSource      string
	providerRouteEnvironment string
	providerRouteConfigID    string
	snapshotProvider         gatewayProviderRouteSnapshotProvider
	bridge                   bridgeClient
}

func (validator configuredWorkflowTemplateTargetBindingValidator) ValidateTargetBinding(
	ctx WorkflowTemplateCatalogContext,
	application ApplicationCatalogRecord,
	snapshot WorkflowDefinitionSnapshot,
) error {
	if err := (strictWorkflowTemplateTargetBindingValidator{}).ValidateTargetBinding(ctx, application, snapshot); err != nil {
		return err
	}
	if len(snapshot.ProviderRefs) == 0 {
		return nil
	}
	environment := strings.TrimSpace(validator.providerRouteEnvironment)
	configurationID := strings.TrimSpace(validator.providerRouteConfigID)
	if strings.TrimSpace(validator.providerRouteSource) != "admin_snapshot_dev_test" ||
		(environment != adminProviderRouteEnvironmentDevelopment && environment != adminProviderRouteEnvironmentTest) ||
		!adminProviderRouteIdentifierPattern.MatchString(configurationID) || validator.snapshotProvider == nil || validator.bridge == nil {
		return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
	}
	active, found, err := validator.snapshotProvider.ReadActiveSnapshot(gatewayProviderRouteScope{
		RequestContext: workflowTemplateRequestContext(ctx), RequestID: ctx.RequestID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: environment,
		ConfigurationID: configurationID, ActorRef: ctx.ActorRef,
	})
	if err != nil || !found || active.TenantRef != ctx.TenantRef || active.WorkspaceID != ctx.WorkspaceID ||
		active.Environment != environment || active.ConfigurationID != configurationID {
		return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
	}
	inventory, err := validator.bridge.DescribeInventory(workflowTemplateRequestContext(ctx))
	if err != nil {
		return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
	}
	for _, providerRef := range snapshot.ProviderRefs {
		profileID := strings.TrimPrefix(providerRef, "profile:")
		assignment, ok := gatewayProviderAttemptAssignment(active, profileID)
		if !ok {
			return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
		}
		expected, ok := gatewayProviderRouteBinding(active.InventoryBindings, assignment.ProfileID)
		if !ok || !expected.Enabled {
			return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
		}
		matchedProfile := -1
		profileKey, ok := adminProviderRouteRuntimeProfileKey(environment, assignment.RuntimeProfileRef)
		if !ok {
			return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
		}
		for index := range inventory.Profiles {
			profile := inventory.Profiles[index]
			if strings.TrimSpace(profile.ProviderID) != assignment.ProviderID || strings.TrimSpace(profile.Profile) != profileKey {
				continue
			}
			if matchedProfile >= 0 {
				return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
			}
			matchedProfile = index
		}
		if matchedProfile < 0 {
			return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
		}
		current, bindingErr := adminProviderRouteInventoryBindingFromProfile(
			environment, assignment.ProviderID, assignment.RuntimeProfileRef, inventory.Profiles[matchedProfile],
		)
		current.ProfileID = assignment.ProfileID
		if bindingErr != nil || !gatewayProviderRouteBindingsEqual(expected, current) {
			return errors.New(WorkflowTemplateFailureTargetBindingUnavailable)
		}
	}
	return nil
}

type workflowTemplateCatalogService struct {
	store         workflowTemplateCatalogRepository
	definitions   workflowDefinitionReleaseRepository
	applications  applicationCatalogRepository
	drafts        savedWorkflowDraftService
	targetBinding workflowTemplateTargetBindingValidator
	now           func() time.Time
}

func newWorkflowTemplateCatalogService(store workflowTemplateCatalogRepository, definitions workflowDefinitionReleaseRepository, applications applicationCatalogRepository, drafts savedWorkflowDraftStore) workflowTemplateCatalogService {
	return workflowTemplateCatalogService{
		store: store, definitions: definitions, applications: applications, drafts: newSavedWorkflowDraftService(drafts),
		targetBinding: strictWorkflowTemplateTargetBindingValidator{}, now: func() time.Time { return time.Now().UTC() },
	}
}

func (service workflowTemplateCatalogService) CreateCandidate(ctx WorkflowTemplateCatalogContext, input WorkflowTemplateCandidateCreateInput) WorkflowTemplateCatalogResult {
	input = normalizeWorkflowTemplateCandidateInput(input)
	if !validWorkflowTemplateContext(ctx) || !applicationDraftIdentifierPattern.MatchString(input.CandidateID) ||
		!applicationDraftIdentifierPattern.MatchString(input.TemplateID) || !applicationCatalogIDPattern.MatchString(input.SourceApplicationID) ||
		!applicationDraftIdentifierPattern.MatchString(input.SourceDefinitionID) || input.SourceDefinitionVersion < 1 {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	if workflowTemplateMetadataContainsSecret(input) {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureSecretMaterialForbidden}
	}
	if !validWorkflowTemplateMetadata(input.Title, input.Summary, input.UsageNotes, input.Labels) {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	if failure := service.requireActiveSourceApplication(ctx, input.SourceApplicationID, ctx.OwnerSubjectRef); failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: failure}
	}
	definition, failure := service.readExactDefinition(ctx, input.SourceApplicationID, ctx.OwnerSubjectRef, input.SourceDefinitionID, input.SourceDefinitionVersion, "")
	if failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: failure}
	}
	portability, failure := validateWorkflowTemplatePortability(definition.Snapshot)
	if failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: failure}
	}
	candidate, err := service.store.CreateCandidate(ctx, WorkflowTemplateCandidate{
		CandidateID: input.CandidateID, TemplateID: input.TemplateID, SourceApplicationID: input.SourceApplicationID,
		SourceOwnerSubjectRef: ctx.OwnerSubjectRef, SourceDefinitionID: definition.DefinitionID,
		SourceDefinitionVersion: definition.Version, SourceDefinitionDigest: definition.DefinitionDigest,
		Title: input.Title, Summary: input.Summary, UsageNotes: input.UsageNotes, Labels: input.Labels, Portability: portability,
	}, service.now())
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	return WorkflowTemplateCatalogResult{Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion}
}

func (service workflowTemplateCatalogService) ReviewCandidate(ctx WorkflowTemplateCatalogContext, candidateID string, input WorkflowTemplateReviewInput) WorkflowTemplateCatalogResult {
	candidate, err := service.store.ReadCandidate(ctx, strings.TrimSpace(candidateID))
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	if failure := service.requireActiveSourceApplication(ctx, candidate.SourceApplicationID, candidate.SourceOwnerSubjectRef); failure != "" {
		return WorkflowTemplateCatalogResult{Candidate: &candidate, FailureCode: failure, CurrentReviewVersion: candidate.ReviewVersion}
	}
	definition, failure := service.readExactDefinition(ctx, candidate.SourceApplicationID, candidate.SourceOwnerSubjectRef, candidate.SourceDefinitionID, candidate.SourceDefinitionVersion, candidate.SourceDefinitionDigest)
	if failure != "" {
		return WorkflowTemplateCatalogResult{Candidate: &candidate, FailureCode: failure, CurrentReviewVersion: candidate.ReviewVersion}
	}
	if _, failure = validateWorkflowTemplatePortability(definition.Snapshot); failure != "" {
		return WorkflowTemplateCatalogResult{Candidate: &candidate, FailureCode: failure, CurrentReviewVersion: candidate.ReviewVersion}
	}
	updated, version, err := service.store.ReviewCandidate(ctx, candidate.CandidateID, input.ExpectedReviewVersion, strings.TrimSpace(input.Decision), strings.TrimSpace(input.Reason), definition.DefinitionDigest, service.now())
	if err != nil {
		result := workflowTemplateResultFromError(err)
		result.CurrentReviewVersion = candidate.ReviewVersion
		return result
	}
	return WorkflowTemplateCatalogResult{Candidate: &updated, Version: version, CurrentReviewVersion: updated.ReviewVersion}
}

func (service workflowTemplateCatalogService) DecideListing(ctx WorkflowTemplateCatalogContext, templateID string, input WorkflowTemplateListingInput) WorkflowTemplateCatalogResult {
	lineage, err := service.store.DecideListing(ctx, strings.TrimSpace(templateID), input.ExpectedPointerVersion, strings.TrimSpace(input.Decision), input.Version, strings.TrimSpace(input.Reason), service.now())
	if err != nil {
		result := workflowTemplateResultFromError(err)
		if current, readErr := service.store.ReadLineage(ctx, strings.TrimSpace(templateID)); readErr == nil {
			result.CurrentPointerVersion = current.PointerVersion
		}
		return result
	}
	return WorkflowTemplateCatalogResult{Lineage: &lineage, CurrentPointerVersion: lineage.PointerVersion}
}

func (service workflowTemplateCatalogService) Derive(ctx WorkflowTemplateCatalogContext, templateID string, input WorkflowTemplateDerivationInput) WorkflowTemplateCatalogResult {
	templateID = strings.TrimSpace(templateID)
	input.TargetApplicationID = strings.TrimSpace(input.TargetApplicationID)
	input.DraftID, input.Name = strings.TrimSpace(input.DraftID), strings.TrimSpace(input.Name)
	if !validWorkflowTemplateContext(ctx) || !applicationDraftIdentifierPattern.MatchString(templateID) ||
		input.ExpectedPointerVersion < 1 || input.TemplateVersion < 1 || !applicationCatalogIDPattern.MatchString(input.TargetApplicationID) ||
		!applicationDraftIdentifierPattern.MatchString(input.DraftID) || utf8RuneCount(input.Name) < 2 || utf8RuneCount(input.Name) > 120 ||
		applicationDraftStringContainsSecret(input.Name) || !input.Confirmed {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	lineage, err := service.store.ReadLineage(ctx, templateID)
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	if lineage.PointerVersion != input.ExpectedPointerVersion {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePointerVersionConflict, CurrentPointerVersion: lineage.PointerVersion}
	}
	if lineage.Lifecycle != workflowTemplateLineageListed || lineage.ListedVersion != input.TemplateVersion {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureNotListed, CurrentPointerVersion: lineage.PointerVersion}
	}
	version, err := service.store.ReadVersion(ctx, templateID, input.TemplateVersion)
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	if version.TemplateDigest != lineage.ListedDigest {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureStoreUnavailable}
	}
	if failure := service.requireActiveSourceApplication(ctx, version.SourceApplicationID, version.SourceOwnerSubjectRef); failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: failure}
	}
	definition, failure := service.readExactDefinition(ctx, version.SourceApplicationID, version.SourceOwnerSubjectRef, version.SourceDefinitionID, version.SourceDefinitionVersion, version.SourceDefinitionDigest)
	if failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: failure}
	}
	target, failure := service.activeApplication(ctx, input.TargetApplicationID, ctx.OwnerSubjectRef)
	if failure != "" {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureTargetApplicationUnavailable}
	}
	if service.targetBinding == nil || service.targetBinding.ValidateTargetBinding(ctx, target, definition.Snapshot) != nil {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureTargetBindingUnavailable}
	}
	payload := workflowTemplateDraftPayload(ctx, input, version, definition)
	draftContext := SavedWorkflowDraftContext{
		RequestContext: workflowTemplateRequestContext(ctx),
		RequestID:      ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: input.TargetApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef,
		ScopeGrants: []string{"workflow_drafts:read", "workflow_drafts:write"}, AuditRef: ctx.AuditRef, WriteEnabled: true,
	}
	created := service.drafts.SaveDraft(draftContext, SaveWorkflowDraftRequest{Payload: payload})
	if created.FailureCode != "" || created.Draft == nil {
		if created.FailureCode == SavedWorkflowDraftFailureVersionConflict || created.FailureCode == SavedWorkflowDraftFailureNotFound {
			return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureDraftIDConflict}
		}
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureStoreUnavailable}
	}
	return WorkflowTemplateCatalogResult{Lineage: &lineage, Version: &version, Draft: created.Draft, CurrentPointerVersion: lineage.PointerVersion}
}

func (service workflowTemplateCatalogService) ReadCandidate(ctx WorkflowTemplateCatalogContext, candidateID string) WorkflowTemplateCatalogResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(candidateID)) {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	candidate, err := service.store.ReadCandidate(ctx, strings.TrimSpace(candidateID))
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	return WorkflowTemplateCatalogResult{Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion}
}

func (service workflowTemplateCatalogService) ReadTemplate(ctx WorkflowTemplateCatalogContext, templateID string) WorkflowTemplateCatalogResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(templateID)) {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	lineage, err := service.store.ReadLineage(ctx, strings.TrimSpace(templateID))
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	result := WorkflowTemplateCatalogResult{Lineage: &lineage, CurrentPointerVersion: lineage.PointerVersion}
	if lineage.ListedVersion > 0 {
		version, versionErr := service.store.ReadVersion(ctx, lineage.TemplateID, lineage.ListedVersion)
		if versionErr != nil {
			return workflowTemplateResultFromError(versionErr)
		}
		result.Version = &version
	}
	return result
}

func (service workflowTemplateCatalogService) ReadVersion(ctx WorkflowTemplateCatalogContext, templateID string, version int) WorkflowTemplateCatalogResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(templateID)) || version < 1 {
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	value, err := service.store.ReadVersion(ctx, strings.TrimSpace(templateID), version)
	if err != nil {
		return workflowTemplateResultFromError(err)
	}
	return WorkflowTemplateCatalogResult{Version: &value}
}

func (service workflowTemplateCatalogService) ListCandidates(ctx WorkflowTemplateCatalogContext, input WorkflowTemplateListInput) WorkflowTemplateCandidateListResult {
	state, limit, snapshotAt, afterTime, afterID, failure := normalizeWorkflowTemplateListInput(ctx, input, "candidate")
	if failure != "" {
		return WorkflowTemplateCandidateListResult{Candidates: []WorkflowTemplateCandidate{}, FailureCode: failure}
	}
	values, err := service.store.ListCandidates(ctx)
	if err != nil {
		return WorkflowTemplateCandidateListResult{Candidates: []WorkflowTemplateCandidate{}, FailureCode: workflowTemplateResultFromError(err).FailureCode}
	}
	filtered := make([]WorkflowTemplateCandidate, 0, len(values))
	for _, value := range values {
		if (state != "" && value.State != state) || !workflowTemplateListRecordIncluded(value.UpdatedAt, value.CandidateID, snapshotAt, afterTime, afterID) {
			continue
		}
		filtered = append(filtered, value)
	}
	result := WorkflowTemplateCandidateListResult{Candidates: filtered}
	if len(result.Candidates) > limit {
		result.Candidates = result.Candidates[:limit]
		last := result.Candidates[len(result.Candidates)-1]
		result.NextCursor = encodeWorkflowTemplateListCursor(ctx, "candidate", state, limit, snapshotAt, last.UpdatedAt, last.CandidateID)
	}
	return result
}

func (service workflowTemplateCatalogService) ListTemplates(ctx WorkflowTemplateCatalogContext, input WorkflowTemplateListInput) WorkflowTemplateLineageListResult {
	_, limit, snapshotAt, afterTime, afterID, failure := normalizeWorkflowTemplateListInput(ctx, WorkflowTemplateListInput{Limit: input.Limit, Cursor: input.Cursor}, "template")
	if failure != "" || strings.TrimSpace(input.State) != "" {
		if failure == "" {
			failure = WorkflowTemplateFailurePayloadInvalid
		}
		return WorkflowTemplateLineageListResult{Lineages: []WorkflowTemplateLineage{}, FailureCode: failure}
	}
	values, err := service.store.ListLineages(ctx)
	if err != nil {
		return WorkflowTemplateLineageListResult{Lineages: []WorkflowTemplateLineage{}, FailureCode: workflowTemplateResultFromError(err).FailureCode}
	}
	filtered := make([]WorkflowTemplateLineage, 0, len(values))
	for _, value := range values {
		if value.Lifecycle != workflowTemplateLineageListed || !workflowTemplateListRecordIncluded(value.UpdatedAt, value.TemplateID, snapshotAt, afterTime, afterID) {
			continue
		}
		filtered = append(filtered, value)
	}
	result := WorkflowTemplateLineageListResult{Lineages: filtered}
	if len(result.Lineages) > limit {
		result.Lineages = result.Lineages[:limit]
		last := result.Lineages[len(result.Lineages)-1]
		result.NextCursor = encodeWorkflowTemplateListCursor(ctx, "template", "", limit, snapshotAt, last.UpdatedAt, last.TemplateID)
	}
	return result
}

func (service workflowTemplateCatalogService) ListVersions(ctx WorkflowTemplateCatalogContext, templateID string, input WorkflowTemplateListInput) WorkflowTemplateVersionListResult {
	templateID = strings.TrimSpace(templateID)
	if !applicationDraftIdentifierPattern.MatchString(templateID) || strings.TrimSpace(input.State) != "" {
		return WorkflowTemplateVersionListResult{Versions: []WorkflowTemplateVersion{}, FailureCode: WorkflowTemplateFailurePayloadInvalid}
	}
	_, limit, snapshotAt, afterTime, afterID, failure := normalizeWorkflowTemplateListInput(ctx, WorkflowTemplateListInput{Limit: input.Limit, Cursor: input.Cursor}, "version:"+templateID)
	if failure != "" {
		return WorkflowTemplateVersionListResult{Versions: []WorkflowTemplateVersion{}, FailureCode: failure}
	}
	values, err := service.store.ListVersions(ctx, templateID)
	if err != nil {
		return WorkflowTemplateVersionListResult{Versions: []WorkflowTemplateVersion{}, FailureCode: workflowTemplateResultFromError(err).FailureCode}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Version > values[j].Version })
	filtered := make([]WorkflowTemplateVersion, 0, len(values))
	for _, value := range values {
		id := integerString(value.Version)
		if workflowTemplateListRecordIncluded(value.CreatedAt, id, snapshotAt, afterTime, afterID) {
			filtered = append(filtered, value)
		}
	}
	result := WorkflowTemplateVersionListResult{Versions: filtered}
	if len(result.Versions) > limit {
		result.Versions = result.Versions[:limit]
		last := result.Versions[len(result.Versions)-1]
		result.NextCursor = encodeWorkflowTemplateListCursor(ctx, "version:"+templateID, "", limit, snapshotAt, last.CreatedAt, integerString(last.Version))
	}
	return result
}

func (service workflowTemplateCatalogService) readExactDefinition(ctx WorkflowTemplateCatalogContext, applicationID, ownerRef, definitionID string, version int, expectedDigest string) (WorkflowDefinitionVersion, string) {
	if service.definitions == nil {
		return WorkflowDefinitionVersion{}, WorkflowTemplateFailureStoreUnavailable
	}
	definitionContext := WorkflowDefinitionReleaseContext{
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: applicationID,
		OwnerSubjectRef: ownerRef, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	value, err := service.definitions.ReadVersion(definitionContext, definitionID, version)
	if errors.Is(err, errWorkflowDefinitionNotFound) {
		return WorkflowDefinitionVersion{}, WorkflowTemplateFailureSourceDefinitionNotFound
	}
	if err != nil || validateStoredWorkflowDefinitionVersion(value) != nil {
		return WorkflowDefinitionVersion{}, WorkflowTemplateFailureStoreUnavailable
	}
	if expectedDigest != "" && value.DefinitionDigest != expectedDigest {
		return WorkflowDefinitionVersion{}, WorkflowTemplateFailureSourceDefinitionChanged
	}
	digest, digestErr := workflowDefinitionSnapshotDigest(value.Snapshot)
	if digestErr != nil || digest != value.DefinitionDigest {
		return WorkflowDefinitionVersion{}, WorkflowTemplateFailureSourceDefinitionChanged
	}
	return value, ""
}

func (service workflowTemplateCatalogService) requireActiveSourceApplication(ctx WorkflowTemplateCatalogContext, applicationID, ownerRef string) string {
	_, failure := service.activeApplication(ctx, applicationID, ownerRef)
	if failure == WorkflowTemplateFailureTargetApplicationUnavailable {
		return WorkflowTemplateFailureSourceApplicationUnavailable
	}
	return failure
}

func (service workflowTemplateCatalogService) activeApplication(ctx WorkflowTemplateCatalogContext, applicationID, ownerRef string) (ApplicationCatalogRecord, string) {
	if service.applications == nil {
		return ApplicationCatalogRecord{}, WorkflowTemplateFailureStoreUnavailable
	}
	result := newApplicationCatalogService(service.applications).RequireActive(ApplicationCatalogContext{
		RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ActorRef: ctx.ActorRef, OwnerSubjectRef: ownerRef, AuditRef: ctx.AuditRef,
	}, applicationID)
	if result.FailureCode == ApplicationCatalogFailureNotFound || result.FailureCode == ApplicationCatalogFailureArchived {
		return ApplicationCatalogRecord{}, WorkflowTemplateFailureTargetApplicationUnavailable
	}
	if result.FailureCode != "" || result.Record == nil {
		return ApplicationCatalogRecord{}, WorkflowTemplateFailureStoreUnavailable
	}
	if result.Record.TenantRef != ctx.TenantRef || result.Record.WorkspaceID != ctx.WorkspaceID || result.Record.OwnerSubjectRef != ownerRef {
		return ApplicationCatalogRecord{}, WorkflowTemplateFailureScopeDenied
	}
	return *result.Record, ""
}

func validateWorkflowTemplatePortability(snapshot WorkflowDefinitionSnapshot) (WorkflowTemplatePortabilitySummary, string) {
	if workflowDefinitionSnapshotContainsForbiddenMaterial(snapshot) {
		return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureSecretMaterialForbidden
	}
	if snapshot.ExecutionProfile != workflowDefinitionExecutorProfile || snapshot.SchemaVersion != savedWorkflowDraftSchemaVersion {
		return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureSourceProfileUnsupported
	}
	if len(snapshot.ToolRefs) > 0 || len(snapshot.RAGRefs) > 0 || len(snapshot.RequestedCapabilities) > 0 {
		return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureForbiddenCapability
	}
	if !validWorkflowTemplateProviderBindings(snapshot) {
		return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureSourceProfileUnsupported
	}
	nodeKinds := make([]string, 0, 4)
	seenKinds := map[string]struct{}{}
	risk := "low"
	for _, node := range snapshot.Nodes {
		switch node.NodeType {
		case "prompt", "llm", "condition", "output":
		default:
			return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureForbiddenCapability
		}
		if node.RequiresConfirmation || node.ToolRef != "" || node.RAGRef != "" {
			return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureForbiddenCapability
		}
		if _, found := seenKinds[node.NodeType]; !found {
			seenKinds[node.NodeType] = struct{}{}
			nodeKinds = append(nodeKinds, node.NodeType)
		}
		if node.RiskLevel == "high" || (node.RiskLevel == "medium" && risk == "low") {
			risk = node.RiskLevel
		}
	}
	sort.Strings(nodeKinds)
	providerRefs := cloneStringSlice(snapshot.ProviderRefs)
	sort.Strings(providerRefs)
	executionDraft := workflowDefinitionSnapshotAsDraft(
		WorkflowRunContext{WorkspaceID: "workspace_portability_check", ApplicationID: "application_portability_check"},
		WorkflowDefinitionVersion{SourceDraftVersion: 1, Snapshot: snapshot},
	)
	if !workflowDefinitionDraftMatchesDefaultContract(executionDraft) || workflowDefinitionExecutionBlocker(executionDraft) != "" {
		return WorkflowTemplatePortabilitySummary{}, WorkflowTemplateFailureForbiddenCapability
	}
	return WorkflowTemplatePortabilitySummary{
		ExecutionProfile: snapshot.ExecutionProfile, NodeKinds: nodeKinds, ProviderRefs: providerRefs,
		RiskLevel: risk, Portable: true, Blockers: []string{},
	}, ""
}

func validWorkflowTemplateProviderBindings(snapshot WorkflowDefinitionSnapshot) bool {
	known := make(map[string]struct{}, len(snapshot.ProviderRefs))
	for _, raw := range snapshot.ProviderRefs {
		reference := strings.TrimSpace(raw)
		if reference != raw || !validWorkflowTemplateProviderRef(reference) ||
			applicationDraftStringContainsSecret(reference) {
			return false
		}
		if _, duplicate := known[reference]; duplicate {
			return false
		}
		known[reference] = struct{}{}
	}
	for _, node := range snapshot.Nodes {
		reference := strings.TrimSpace(node.ProviderRef)
		if node.NodeType != "llm" {
			if reference != "" {
				return false
			}
			continue
		}
		if reference == "" && len(known) == 0 {
			continue
		}
		if reference != node.ProviderRef {
			return false
		}
		if _, found := known[reference]; !found {
			return false
		}
	}
	return true
}

func validWorkflowTemplateProviderRef(value string) bool {
	if !strings.HasPrefix(value, "profile:") || len(value) > 160 {
		return false
	}
	return applicationDraftIdentifierPattern.MatchString(strings.TrimPrefix(value, "profile:"))
}

func workflowTemplateDraftPayload(ctx WorkflowTemplateCatalogContext, input WorkflowTemplateDerivationInput, template WorkflowTemplateVersion, definition WorkflowDefinitionVersion) SavedWorkflowDraftPayload {
	source := workflowDefinitionSnapshotAsDraft(WorkflowRunContext{WorkspaceID: ctx.WorkspaceID, ApplicationID: input.TargetApplicationID}, definition)
	return SavedWorkflowDraftPayload{
		DraftID: input.DraftID, WorkspaceID: ctx.WorkspaceID, ApplicationID: input.TargetApplicationID,
		SourceDefinitionID: definition.DefinitionID, BaseDefinitionVersion: definition.Version,
		SchemaVersion: source.SchemaVersion, DraftStatus: SavedWorkflowDraftStatusValidForReview,
		Name: input.Name, Description: template.Summary, Nodes: cloneSavedWorkflowDraftNodes(source.Nodes), Edges: cloneSavedWorkflowDraftEdges(source.Edges),
		InputContract: cloneSavedWorkflowDraftContract(source.InputContract), OutputContract: cloneSavedWorkflowDraftContract(source.OutputContract),
		ProviderRefs: cloneStringSlice(source.ProviderRefs), ToolRefs: cloneStringSlice(source.ToolRefs), RAGRefs: cloneStringSlice(source.RAGRefs),
		RequestedCapabilities: cloneStringSlice(source.RequestedCapabilities),
		AdditionalFields: map[string]any{savedWorkflowTemplateDerivationAdditionalField: map[string]any{
			"version": savedWorkflowTemplateDerivationVersion, "source_kind": savedWorkflowTemplateDerivationSourceKind,
			"template_id": template.TemplateID, "template_version": template.Version, "template_digest": template.TemplateDigest,
			"source_definition_id": definition.DefinitionID, "source_definition_version": definition.Version,
			"source_definition_digest": definition.DefinitionDigest,
		}},
	}
}

func normalizeWorkflowTemplateCandidateInput(input WorkflowTemplateCandidateCreateInput) WorkflowTemplateCandidateCreateInput {
	input.CandidateID, input.TemplateID = strings.TrimSpace(input.CandidateID), strings.TrimSpace(input.TemplateID)
	input.SourceApplicationID, input.SourceDefinitionID = strings.TrimSpace(input.SourceApplicationID), strings.TrimSpace(input.SourceDefinitionID)
	input.Title, input.Summary, input.UsageNotes = strings.TrimSpace(input.Title), strings.TrimSpace(input.Summary), strings.TrimSpace(input.UsageNotes)
	labels := make([]string, 0, len(input.Labels))
	for _, label := range input.Labels {
		labels = append(labels, strings.ToLower(strings.TrimSpace(label)))
	}
	sort.Strings(labels)
	input.Labels = labels
	return input
}

func workflowTemplateMetadataContainsSecret(input WorkflowTemplateCandidateCreateInput) bool {
	payload, err := json.Marshal(input)
	return err != nil || applicationDraftStringContainsSecret(string(payload))
}

func workflowTemplateResultFromError(err error) WorkflowTemplateCatalogResult {
	switch {
	case errors.Is(err, errWorkflowTemplateCandidateNotFound):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureCandidateNotFound}
	case errors.Is(err, errWorkflowTemplateCandidateConflict):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureCandidateVersionConflict}
	case errors.Is(err, errWorkflowTemplateReviewTransition):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureReviewTransitionInvalid}
	case errors.Is(err, errWorkflowTemplateVersionNotFound):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureVersionNotFound}
	case errors.Is(err, errWorkflowTemplatePointerConflict):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePointerVersionConflict}
	case errors.Is(err, errWorkflowTemplateNotListed):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureNotListed}
	case errors.Is(err, errWorkflowTemplatePayloadInvalid):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid}
	case errors.Is(err, errWorkflowTemplateSecretMaterialForbidden):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureSecretMaterialForbidden}
	case errors.Is(err, errWorkflowTemplateSourceDefinitionChanged):
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureSourceDefinitionChanged}
	default:
		return WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailureStoreUnavailable}
	}
}

type workflowTemplateListCursor struct {
	SchemaVersion string `json:"schema_version"`
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	ListKind      string `json:"list_kind"`
	State         string `json:"state"`
	Limit         int    `json:"limit"`
	SnapshotAt    string `json:"snapshot_at"`
	LastTime      string `json:"last_time"`
	LastID        string `json:"last_id"`
}

func normalizeWorkflowTemplateListInput(ctx WorkflowTemplateCatalogContext, input WorkflowTemplateListInput, listKind string) (state string, limit int, snapshotAt, afterTime, afterID string, failure string) {
	state = strings.TrimSpace(input.State)
	if state != "" && state != workflowTemplateCandidatePending && state != workflowTemplateCandidateApproved && state != workflowTemplateCandidateRejected && state != workflowTemplateCandidateChangesRequested && state != workflowTemplateCandidateWithdrawn {
		failure = WorkflowTemplateFailurePayloadInvalid
		return
	}
	limit = input.Limit
	if limit == 0 {
		limit = workflowTemplateDefaultListLimit
	}
	if limit < 1 || limit > workflowTemplateMaximumListLimit {
		failure = WorkflowTemplateFailurePayloadInvalid
		return
	}
	if strings.TrimSpace(input.Cursor) == "" {
		snapshotAt = time.Now().UTC().Format(time.RFC3339Nano)
		return
	}
	cursor, err := decodeWorkflowTemplateListCursor(input.Cursor)
	if err != nil || cursor.SchemaVersion != workflowTemplateListCursorSchemaVersion || cursor.TenantRef != ctx.TenantRef || cursor.WorkspaceID != ctx.WorkspaceID ||
		cursor.ListKind != listKind || cursor.State != state || cursor.Limit != limit {
		failure = WorkflowTemplateFailureCursorInvalid
		return
	}
	if _, err = time.Parse(time.RFC3339Nano, cursor.SnapshotAt); err != nil {
		failure = WorkflowTemplateFailureCursorInvalid
		return
	}
	if _, err = time.Parse(time.RFC3339Nano, cursor.LastTime); err != nil {
		failure = WorkflowTemplateFailureCursorInvalid
		return
	}
	snapshotAt, afterTime, afterID = cursor.SnapshotAt, cursor.LastTime, cursor.LastID
	return
}

func workflowTemplateListRecordIncluded(recordTime, recordID, snapshotAt, afterTime, afterID string) bool {
	recordTimestamp, recordErr := time.Parse(time.RFC3339Nano, recordTime)
	snapshotTimestamp, snapshotErr := time.Parse(time.RFC3339Nano, snapshotAt)
	if recordErr != nil || snapshotErr != nil || recordTimestamp.After(snapshotTimestamp) {
		return false
	}
	if afterTime == "" {
		return true
	}
	afterTimestamp, afterErr := time.Parse(time.RFC3339Nano, afterTime)
	if afterErr != nil {
		return false
	}
	return recordTimestamp.Before(afterTimestamp) || (recordTimestamp.Equal(afterTimestamp) && recordID < afterID)
}

func encodeWorkflowTemplateListCursor(ctx WorkflowTemplateCatalogContext, listKind, state string, limit int, snapshotAt, lastTime, lastID string) string {
	payload, err := json.Marshal(workflowTemplateListCursor{
		SchemaVersion: workflowTemplateListCursorSchemaVersion, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ListKind: listKind, State: state, Limit: limit, SnapshotAt: snapshotAt, LastTime: lastTime, LastID: lastID,
	})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(payload)
}

func decodeWorkflowTemplateListCursor(value string) (workflowTemplateListCursor, error) {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 4096 {
		return workflowTemplateListCursor{}, errors.New("invalid workflow template cursor")
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(payload) > 2048 || validateNoDuplicateJSONFields(payload) != nil {
		return workflowTemplateListCursor{}, errors.New("invalid workflow template cursor")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var cursor workflowTemplateListCursor
	if err = decoder.Decode(&cursor); err != nil {
		return workflowTemplateListCursor{}, err
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return workflowTemplateListCursor{}, errors.New("invalid workflow template cursor")
	}
	return cursor, nil
}

func integerString(value int) string {
	return fmt.Sprintf("%020d", value)
}

func utf8RuneCount(value string) int {
	return len([]rune(strings.TrimSpace(value)))
}
