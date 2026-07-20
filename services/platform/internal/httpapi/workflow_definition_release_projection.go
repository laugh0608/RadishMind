package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
)

type workflowDefinitionSummarySource struct {
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	OwnerSubjectRef string
	Version         WorkflowDefinitionVersion
	Activation      *WorkflowDefinitionActivation
}

type workflowDefinitionSummaryCursor struct {
	UpdatedAt      string `json:"updated_at"`
	ApplicationRef string `json:"application_ref"`
	DefinitionID   string `json:"definition_id"`
}

func (store *workflowDefinitionReleaseStore) ListSummaries(ctx ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	sources := make([]workflowDefinitionSummarySource, 0, len(store.versions))
	for key, versions := range store.versions {
		parts := strings.Split(key, "\x00")
		if len(parts) != 5 || parts[0] != ctx.TenantRef || parts[3] != ctx.SubjectRef || len(versions) == 0 {
			continue
		}
		latest := cloneWorkflowDefinitionVersion(versions[len(versions)-1])
		var activation *WorkflowDefinitionActivation
		if current, ok := store.activations[key]; ok {
			copied := cloneWorkflowDefinitionActivation(current)
			activation = &copied
			if copied.State == workflowDefinitionActivationActive && copied.ActiveVersion > 0 && copied.ActiveVersion <= len(versions) {
				latest = cloneWorkflowDefinitionVersion(versions[copied.ActiveVersion-1])
			}
		}
		sources = append(sources, workflowDefinitionSummarySource{TenantRef: parts[0], WorkspaceID: parts[1], ApplicationID: parts[2], OwnerSubjectRef: parts[3], Version: latest, Activation: activation})
	}
	return projectWorkflowDefinitionSummaries(ctx, request, sources)
}

func projectWorkflowDefinitionSummaries(ctx ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest, sources []workflowDefinitionSummarySource) ListWorkflowDefinitionSummariesResult {
	items := make([]WorkflowDefinitionSummary, 0, len(sources))
	for _, source := range sources {
		version := source.Version
		status := workflowDefinitionActivationInactive
		updatedAt := version.CreatedAt
		if source.Activation != nil {
			updatedAt = source.Activation.UpdatedAt
			if source.Activation.State == workflowDefinitionActivationActive {
				status = workflowDefinitionActivationActive
			}
		}
		risk, confirmation := workflowDefinitionSnapshotRisk(version.Snapshot)
		item := WorkflowDefinitionSummary{WorkflowDefinitionID: version.DefinitionID, TenantRef: source.TenantRef, ApplicationRef: source.ApplicationID, Version: version.Version, DefinitionStatus: status, NodeCount: len(version.Snapshot.Nodes), RiskLevel: risk, RequiresConfirmationCapable: confirmation, UpdatedAt: updatedAt}
		if value := strings.TrimSpace(request.Filters["application_ref"]); value != "" && item.ApplicationRef != value {
			continue
		}
		if value := strings.TrimSpace(request.Filters["definition_status"]); value != "" && item.DefinitionStatus != value {
			continue
		}
		if value := strings.TrimSpace(request.Filters["risk_level"]); value != "" && item.RiskLevel != value {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt != items[j].UpdatedAt {
			return items[i].UpdatedAt > items[j].UpdatedAt
		}
		if items[i].ApplicationRef != items[j].ApplicationRef {
			return items[i].ApplicationRef < items[j].ApplicationRef
		}
		return items[i].WorkflowDefinitionID < items[j].WorkflowDefinitionID
	})
	if request.Sort != "" && request.Sort != "updated_at_desc" {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureInvalidFilter)
	}
	start := 0
	if strings.TrimSpace(request.Cursor) != "" {
		cursor, err := decodeWorkflowDefinitionSummaryCursor(request.Cursor)
		if err != nil {
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureInvalidFilter)
		}
		found := false
		for index, item := range items {
			if item.UpdatedAt == cursor.UpdatedAt && item.ApplicationRef == cursor.ApplicationRef && item.WorkflowDefinitionID == cursor.DefinitionID {
				start = index + 1
				found = true
				break
			}
		}
		if !found {
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureInvalidFilter)
		}
	}
	limit := request.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := append([]WorkflowDefinitionSummary(nil), items[start:end]...)
	var next *string
	if end < len(items) && len(page) > 0 {
		last := page[len(page)-1]
		encoded := encodeWorkflowDefinitionSummaryCursor(workflowDefinitionSummaryCursor{UpdatedAt: last.UpdatedAt, ApplicationRef: last.ApplicationRef, DefinitionID: last.WorkflowDefinitionID})
		next = &encoded
	}
	return ListWorkflowDefinitionSummariesResult{TenantRef: ctx.TenantRef, Items: page, NextCursor: next, AuditRef: ctx.AuditRef}
}

func workflowDefinitionSnapshotRisk(snapshot WorkflowDefinitionSnapshot) (string, bool) {
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	risk := "low"
	confirmation := false
	for _, node := range snapshot.Nodes {
		candidate := strings.ToLower(strings.TrimSpace(node.RiskLevel))
		if rank[candidate] > rank[risk] {
			risk = candidate
		}
		confirmation = confirmation || node.RequiresConfirmation
	}
	return risk, confirmation
}
func encodeWorkflowDefinitionSummaryCursor(value workflowDefinitionSummaryCursor) string {
	payload, _ := json.Marshal(value)
	return base64.RawURLEncoding.EncodeToString(payload)
}
func decodeWorkflowDefinitionSummaryCursor(value string) (workflowDefinitionSummaryCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return workflowDefinitionSummaryCursor{}, err
	}
	var cursor workflowDefinitionSummaryCursor
	if err = decodeWorkflowDefinitionRecord(payload, &cursor); err != nil {
		return workflowDefinitionSummaryCursor{}, err
	}
	if strings.TrimSpace(cursor.UpdatedAt) == "" || strings.TrimSpace(cursor.ApplicationRef) == "" || strings.TrimSpace(cursor.DefinitionID) == "" {
		return workflowDefinitionSummaryCursor{}, errWorkflowDefinitionPayloadInvalid
	}
	return cursor, nil
}
func workflowDefinitionSummaryFailure(ctx ReadRepositoryContext, code ReadRepositoryFailureCode) ListWorkflowDefinitionSummariesResult {
	return ListWorkflowDefinitionSummariesResult{TenantRef: ctx.TenantRef, Items: []WorkflowDefinitionSummary{}, FailureCode: code, AuditRef: ctx.AuditRef}
}
