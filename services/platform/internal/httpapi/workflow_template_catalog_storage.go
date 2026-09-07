package httpapi

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func workflowTemplateRequestContext(value WorkflowTemplateCatalogContext) context.Context {
	if value.RequestContext != nil {
		return value.RequestContext
	}
	return context.Background()
}

func workflowTemplateUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, errWorkflowTemplateStoreUnavailable
	}
	unixNano := parsed.UTC().UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(parsed.UTC()) {
		return 0, errWorkflowTemplateStoreUnavailable
	}
	return unixNano, nil
}

func workflowTemplatePostgresTimeMatches(stored time.Time, encoded string) bool {
	decoded, err := time.Parse(time.RFC3339Nano, encoded)
	return err == nil && stored.UTC().UnixMicro() == decoded.UTC().UnixMicro()
}

func workflowTemplateAuditSequence(value WorkflowTemplateAudit) (int, error) {
	const prefix = "template-audit-"
	if !strings.HasPrefix(value.AuditID, prefix) {
		return 0, errWorkflowTemplateStoreUnavailable
	}
	sequence, err := strconv.Atoi(strings.TrimPrefix(value.AuditID, prefix))
	if err != nil || sequence < 1 {
		return 0, errWorkflowTemplateStoreUnavailable
	}
	return sequence, nil
}

func loadWorkflowTemplateCandidateRecord(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, payload []byte) (WorkflowTemplateCandidate, error) {
	var value WorkflowTemplateCandidate
	if store == nil || decodeWorkflowTemplateRecord(payload, &value) != nil {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, value.CandidateID)
	if _, duplicate := store.candidates[key]; duplicate {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateStoreUnavailable
	}
	store.candidates[key] = cloneWorkflowTemplateCandidate(value)
	return value, nil
}

func loadWorkflowTemplateVersionRecord(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, payload []byte) (WorkflowTemplateVersion, error) {
	var value WorkflowTemplateVersion
	if store == nil || decodeWorkflowTemplateRecord(payload, &value) != nil {
		return WorkflowTemplateVersion{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, value.TemplateID)
	if value.Version != len(store.versions[key])+1 {
		return WorkflowTemplateVersion{}, errWorkflowTemplateStoreUnavailable
	}
	store.versions[key] = append(store.versions[key], cloneWorkflowTemplateVersion(value))
	return value, nil
}

func loadWorkflowTemplateLineageRecord(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, payload []byte) (WorkflowTemplateLineage, error) {
	var value WorkflowTemplateLineage
	if store == nil || decodeWorkflowTemplateRecord(payload, &value) != nil ||
		value.TenantRef != ctx.TenantRef || value.WorkspaceID != ctx.WorkspaceID ||
		validateWorkflowTemplateLineageVersion(value, store.versions[workflowTemplateScopeKey(ctx, value.TemplateID)]) != nil {
		return WorkflowTemplateLineage{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, value.TemplateID)
	if _, duplicate := store.lineages[key]; duplicate {
		return WorkflowTemplateLineage{}, errWorkflowTemplateStoreUnavailable
	}
	store.lineages[key] = cloneWorkflowTemplateLineage(value)
	return value, nil
}

func loadWorkflowTemplateAuditRecord(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, sequence int, payload []byte) (WorkflowTemplateAudit, error) {
	var value WorkflowTemplateAudit
	if store == nil || decodeWorkflowTemplateRecord(payload, &value) != nil {
		return WorkflowTemplateAudit{}, errWorkflowTemplateStoreUnavailable
	}
	expected := len(store.audits[workflowTemplateScopeKey(ctx, "audits")]) + 1
	decodedSequence, err := workflowTemplateAuditSequence(value)
	if err != nil || sequence != expected || decodedSequence != sequence {
		return WorkflowTemplateAudit{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, "audits")
	store.audits[key] = append(store.audits[key], value)
	return value, nil
}

func validateWorkflowTemplateDecisionEvidence(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, candidateID string, reviewVersion int, payload []byte) error {
	var value WorkflowTemplateReviewDecision
	if decodeWorkflowTemplateRecord(payload, &value) != nil || value.ReviewVersion != reviewVersion {
		return errWorkflowTemplateStoreUnavailable
	}
	candidate, found := store.candidates[workflowTemplateScopeKey(ctx, candidateID)]
	if !found || reviewVersion < 1 || reviewVersion > len(candidate.Decisions) || !reflect.DeepEqual(value, candidate.Decisions[reviewVersion-1]) {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateWorkflowTemplateListingEventEvidence(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, templateID string, pointerVersion int, payload []byte) error {
	var value WorkflowTemplateListingEvent
	if decodeWorkflowTemplateRecord(payload, &value) != nil || value.AfterPointerVersion != pointerVersion {
		return errWorkflowTemplateStoreUnavailable
	}
	lineage, found := store.lineages[workflowTemplateScopeKey(ctx, templateID)]
	if !found || pointerVersion < 1 || pointerVersion > len(lineage.Events) || !reflect.DeepEqual(value, lineage.Events[pointerVersion-1]) {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateWorkflowTemplateEvidenceCounts(store *memoryWorkflowTemplateCatalogRepository, ctx WorkflowTemplateCatalogContext, decisions, events map[string]int) error {
	for _, candidate := range store.candidates {
		if decisions[candidate.CandidateID] != len(candidate.Decisions) {
			return errWorkflowTemplateStoreUnavailable
		}
	}
	for _, lineage := range store.lineages {
		if events[lineage.TemplateID] != len(lineage.Events) {
			return errWorkflowTemplateStoreUnavailable
		}
	}
	return nil
}

func workflowTemplateStoredFieldMismatch(values ...bool) error {
	for _, matches := range values {
		if !matches {
			return errWorkflowTemplateStoreUnavailable
		}
	}
	return nil
}
