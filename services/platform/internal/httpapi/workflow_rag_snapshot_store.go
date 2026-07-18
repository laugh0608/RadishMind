package httpapi

import (
	"encoding/json"
	"errors"
	"sort"
	"time"
)

type workflowRAGSnapshotRow interface{ Scan(...any) error }

func encodeWorkflowRAGResource(resource WorkflowRAGSnapshotResource) ([]byte, error) {
	if !workflowRAGSnapshotIDPattern.MatchString(resource.SnapshotID) || !workflowRAGSnapshotKeyPattern.MatchString(resource.SnapshotKey) ||
		(resource.LifecycleState != workflowRAGSnapshotActive && resource.LifecycleState != workflowRAGSnapshotArchived) ||
		resource.LatestVersion < 1 || !workflowRAGDigestPattern.MatchString(resource.LatestDigest) ||
		resource.LatestRAGRef == "" || resource.FragmentCount < 1 || resource.FragmentCount > workflowRAGMaxFragments ||
		resource.TotalContentBytes < 1 || resource.TotalContentBytes > workflowRAGMaxSnapshotBytes {
		return nil, errWorkflowRAGStoreContract
	}
	return json.Marshal(resource)
}

func decodeWorkflowRAGResource(payload []byte, ctx WorkflowRAGSnapshotContext) (WorkflowRAGSnapshotResource, error) {
	var resource WorkflowRAGSnapshotResource
	if err := json.Unmarshal(payload, &resource); err != nil || resource.TenantRef != ctx.TenantRef ||
		resource.WorkspaceID != ctx.WorkspaceID || resource.ApplicationID != ctx.ApplicationID {
		return WorkflowRAGSnapshotResource{}, errWorkflowRAGStoreContract
	}
	if _, err := encodeWorkflowRAGResource(resource); err != nil {
		return WorkflowRAGSnapshotResource{}, err
	}
	return resource, nil
}

func encodeWorkflowRAGSnapshotMetadata(record WorkflowRAGSnapshotRecord, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if err := validateStoredWorkflowRAGRecord(record, ctx); err != nil {
		return nil, err
	}
	metadata := cloneWorkflowRAGSnapshotRecord(record)
	metadata.Fragments = []WorkflowRAGFragment{}
	return json.Marshal(metadata)
}

func decodeWorkflowRAGSnapshotMetadata(payload []byte, fragments []WorkflowRAGFragment, ctx WorkflowRAGSnapshotContext) (WorkflowRAGSnapshotRecord, error) {
	var record WorkflowRAGSnapshotRecord
	if err := json.Unmarshal(payload, &record); err != nil || len(record.Fragments) != 0 {
		return WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreContract
	}
	record.Fragments = cloneWorkflowRAGFragments(fragments)
	sort.Slice(record.Fragments, func(i, j int) bool { return record.Fragments[i].FragmentRef < record.Fragments[j].FragmentRef })
	if err := validateStoredWorkflowRAGRecord(record, ctx); err != nil {
		return WorkflowRAGSnapshotRecord{}, err
	}
	return record, nil
}

func encodeWorkflowRAGFragment(fragment WorkflowRAGFragment) ([]byte, error) {
	if fragment.SchemaVersion != workflowRAGFragmentSchemaVersion || !workflowRAGFragmentRefPattern.MatchString(fragment.FragmentRef) ||
		!workflowRAGSourceTypeAllowed(fragment.SourceType) || !workflowRAGClassificationAllowed(fragment.ContentClassification) ||
		fragment.Content == "" || fragment.ContentBytes != len([]byte(fragment.Content)) || fragment.ContentBytes > workflowRAGMaxFragmentBytes ||
		fragment.ContentDigest != workflowRAGSHA256(fragment.Content) {
		return nil, errWorkflowRAGStoreContract
	}
	return json.Marshal(fragment)
}

func decodeWorkflowRAGFragment(payload []byte) (WorkflowRAGFragment, error) {
	var fragment WorkflowRAGFragment
	if err := json.Unmarshal(payload, &fragment); err != nil {
		return WorkflowRAGFragment{}, errWorkflowRAGStoreContract
	}
	if _, err := encodeWorkflowRAGFragment(fragment); err != nil {
		return WorkflowRAGFragment{}, err
	}
	return fragment, nil
}

func encodeWorkflowRAGAudit(audit WorkflowRAGExecutionAudit, ctx WorkflowRAGSnapshotContext) ([]byte, error) {
	if audit.SchemaVersion != workflowRAGAuditSchemaVersion ||
		!workflowRAGAuditEventAllowed(audit.EventKind) ||
		audit.TenantRef != ctx.TenantRef || audit.WorkspaceID != ctx.WorkspaceID || audit.ApplicationID != ctx.ApplicationID ||
		!workflowRAGSnapshotIDPattern.MatchString(audit.SnapshotID) || !workflowRAGSnapshotKeyPattern.MatchString(audit.SnapshotKey) ||
		audit.SnapshotVersion < 1 || !workflowRAGDigestPattern.MatchString(audit.SnapshotDigest) ||
		audit.ActorRef != ctx.ActorRef || audit.RequestID != ctx.RequestID || audit.AuditRef != ctx.AuditRef {
		return nil, errWorkflowRAGStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, audit.OccurredAt); err != nil {
		return nil, errWorkflowRAGStoreContract
	}
	return json.Marshal(audit)
}

func scanWorkflowRAGResource(row workflowRAGSnapshotRow, ctx WorkflowRAGSnapshotContext) (WorkflowRAGSnapshotResource, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return WorkflowRAGSnapshotResource{}, err
	}
	return decodeWorkflowRAGResource(payload, ctx)
}

func workflowRAGStoreError(err error) error {
	if errors.Is(err, errWorkflowRAGNotFound) || errors.Is(err, errWorkflowRAGScopeDenied) ||
		errors.Is(err, errWorkflowRAGVersionConflict) || errors.Is(err, errWorkflowRAGArchived) || errors.Is(err, errWorkflowRAGStoreContract) {
		return err
	}
	return errWorkflowRAGStoreUnavailable
}
