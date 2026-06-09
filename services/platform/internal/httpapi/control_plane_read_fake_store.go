package httpapi

type controlPlaneReadStore interface {
	TenantSummary(tenantRef string) (map[string]any, bool)
	ApplicationSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string)
	APIKeySummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string)
	QuotaSummary(tenantRef string) (map[string]any, bool)
	WorkflowDefinitionSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string)
	RunRecordSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string)
	AuditSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string)
	SideEffects() controlPlaneReadSideEffects
}

type controlPlaneReadSideEffects struct {
	FakeStoreWriteCount    int `json:"fake_store_write_count"`
	ExecutorCallCount      int `json:"executor_call_count"`
	ConfirmationCallCount  int `json:"confirmation_call_count"`
	BusinessWritebackCount int `json:"business_writeback_count"`
	ReplayCallCount        int `json:"replay_call_count"`
}

type fixtureBackedControlPlaneReadStore struct {
	tenantSummaries             map[string]map[string]any
	applicationSummaries        map[string]controlPlaneReadCursorListFixture
	apiKeySummaries             map[string]controlPlaneReadCursorListFixture
	quotaSummaries              map[string]map[string]any
	workflowDefinitionSummaries map[string]controlPlaneReadCursorListFixture
	runRecordSummaries          map[string]controlPlaneReadCursorListFixture
	auditSummaries              map[string]controlPlaneReadCursorListFixture
}

type controlPlaneReadCursorListFixture struct {
	items      []map[string]any
	nextCursor string
}

func newControlPlaneReadFakeStore() controlPlaneReadStore {
	return fixtureBackedControlPlaneReadStore{
		tenantSummaries: map[string]map[string]any{
			"tenant_demo": {
				"tenant_ref":            "tenant_demo",
				"tenant_display_name":   "Demo tenant",
				"tenant_state":          "active",
				"plan_ref":              "plan_internal",
				"quota_summary_ref":     "quota_demo_current",
				"deployment_status_ref": "deployment_local_read_only",
				"audit_summary_ref":     "audit_demo_latest",
			},
		},
		applicationSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"application_ref":                "app_flow_copilot",
						"tenant_ref":                     "tenant_demo",
						"application_kind":               "workflow_copilot",
						"display_name":                   "RadishFlow Copilot",
						"owner_subject_ref":              "subject_platform_ops",
						"latest_workflow_definition_ref": "wf_radishflow_copilot_latest",
						"last_run_status":                "succeeded",
						"updated_at":                     "2026-05-31T10:20:00Z",
					},
					{
						"application_ref":                "app_docs_assistant",
						"tenant_ref":                     "tenant_demo",
						"application_kind":               "docs_qa",
						"display_name":                   "Radish Docs Assistant",
						"owner_subject_ref":              "subject_docs_team",
						"latest_workflow_definition_ref": "wf_radish_docs_qa_draft",
						"last_run_status":                "blocked",
						"updated_at":                     "2026-05-31T09:40:00Z",
					},
				},
				nextCursor: "cursor_workspace_applications_next",
			},
		},
		apiKeySummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"api_key_id":        "key_ops_readonly",
						"tenant_ref":        "tenant_demo",
						"owner_subject_ref": "subject_platform_ops",
						"scopes": []string{
							"models:read",
							"runs:read",
						},
						"state":        "active",
						"created_at":   "2026-05-20T08:00:00Z",
						"expires_at":   "2026-08-20T08:00:00Z",
						"last_used_at": "2026-05-31T08:45:00Z",
					},
					{
						"api_key_id":        "key_docs_pending_rotation",
						"tenant_ref":        "tenant_demo",
						"owner_subject_ref": "subject_docs_team",
						"scopes": []string{
							"applications:read",
						},
						"state":        "rotation_required",
						"created_at":   "2026-04-12T11:10:00Z",
						"expires_at":   nil,
						"last_used_at": nil,
					},
				},
			},
		},
		quotaSummaries: map[string]map[string]any{
			"tenant_demo": {
				"quota_id":      "quota_demo_current",
				"tenant_ref":    "tenant_demo",
				"period":        "monthly",
				"request_limit": 10000,
				"token_limit":   1000000,
				"cost_limit":    100,
				"usage_snapshot": map[string]any{
					"request_count":  12,
					"token_count":    3456,
					"estimated_cost": 0.42,
				},
				"over_quota_failure_code": "quota_exceeded",
			},
		},
		workflowDefinitionSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"workflow_definition_id":        "wf_radishflow_copilot_latest",
						"tenant_ref":                    "tenant_demo",
						"application_ref":               "app_flow_copilot",
						"version":                       3,
						"definition_status":             "published",
						"node_count":                    8,
						"risk_level":                    "medium",
						"requires_confirmation_capable": true,
						"updated_at":                    "2026-05-31T10:25:00Z",
					},
					{
						"workflow_definition_id":        "wf_radish_docs_qa_draft",
						"tenant_ref":                    "tenant_demo",
						"application_ref":               "app_docs_assistant",
						"version":                       1,
						"definition_status":             "draft",
						"node_count":                    4,
						"risk_level":                    "low",
						"requires_confirmation_capable": false,
						"updated_at":                    "2026-05-31T09:05:00Z",
					},
				},
				nextCursor: "cursor_workspace_workflow_definitions_next",
			},
		},
		runRecordSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"run_id":                 "run_radishflow_copilot_20260531_001",
						"tenant_ref":             "tenant_demo",
						"workflow_definition_id": "wf_radishflow_copilot_latest",
						"application_ref":        "app_flow_copilot",
						"status":                 "succeeded",
						"failure_code":           nil,
						"cost_summary": map[string]any{
							"estimated_cost": 0.08,
							"currency":       "USD",
						},
						"trace_id":     "trace_run_radishflow_copilot_001",
						"started_at":   "2026-05-31T10:31:00Z",
						"completed_at": "2026-05-31T10:31:03Z",
					},
					{
						"run_id":                 "run_radish_docs_qa_20260531_002",
						"tenant_ref":             "tenant_demo",
						"workflow_definition_id": "wf_radish_docs_qa_draft",
						"application_ref":        "app_docs_assistant",
						"status":                 "failed",
						"failure_code":           "read_store_unavailable",
						"cost_summary": map[string]any{
							"estimated_cost": 0.02,
							"currency":       "USD",
						},
						"trace_id":     "trace_run_radish_docs_qa_002",
						"started_at":   "2026-05-31T10:12:00Z",
						"completed_at": "2026-05-31T10:12:01Z",
					},
				},
				nextCursor: "cursor_workspace_run_history_next",
			},
		},
		auditSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"audit_ref":         "audit_admin_policy_read_20260601_001",
						"tenant_ref":        "tenant_demo",
						"actor_subject_ref": "subject_admin_ops",
						"event_kind":        "control_plane.read",
						"resource_ref":      "tenant_demo",
						"decision":          "allowed",
						"failure_code":      nil,
						"trace_id":          "trace_audit_admin_policy_read_001",
						"recorded_at":       "2026-06-01T09:14:00Z",
					},
					{
						"audit_ref":         "audit_cross_tenant_denied_20260601_002",
						"tenant_ref":        "tenant_demo",
						"actor_subject_ref": "subject_workspace_reader",
						"event_kind":        "control_plane.denied",
						"resource_ref":      "tenant_other",
						"decision":          "denied",
						"failure_code":      "scope_denied",
						"trace_id":          "trace_audit_cross_tenant_denied_002",
						"recorded_at":       "2026-06-01T09:10:00Z",
					},
				},
				nextCursor: "cursor_admin_audit_log_next",
			},
		},
	}
}

func (s *Server) controlPlaneReadDataStore() controlPlaneReadStore {
	if s.controlPlaneReadStore != nil {
		return s.controlPlaneReadStore
	}
	return newControlPlaneReadFakeStore()
}

func (store fixtureBackedControlPlaneReadStore) TenantSummary(tenantRef string) (map[string]any, bool) {
	item, ok := store.tenantSummaries[tenantRef]
	if !ok {
		return nil, false
	}
	return cloneControlPlaneReadMap(item), true
}

func (store fixtureBackedControlPlaneReadStore) ApplicationSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	return controlPlaneReadCursorListSummaries(store.applicationSummaries, tenantRef, filters)
}

func (store fixtureBackedControlPlaneReadStore) APIKeySummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	return controlPlaneReadCursorListSummaries(store.apiKeySummaries, tenantRef, filters)
}

func (store fixtureBackedControlPlaneReadStore) QuotaSummary(tenantRef string) (map[string]any, bool) {
	item, ok := store.quotaSummaries[tenantRef]
	if !ok {
		return nil, false
	}
	return cloneControlPlaneReadMap(item), true
}

func (store fixtureBackedControlPlaneReadStore) WorkflowDefinitionSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	return controlPlaneReadCursorListSummaries(store.workflowDefinitionSummaries, tenantRef, filters)
}

func (store fixtureBackedControlPlaneReadStore) RunRecordSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	return controlPlaneReadCursorListSummaries(store.runRecordSummaries, tenantRef, filters)
}

func (store fixtureBackedControlPlaneReadStore) AuditSummaries(tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	return controlPlaneReadCursorListSummaries(store.auditSummaries, tenantRef, filters)
}

func (store fixtureBackedControlPlaneReadStore) SideEffects() controlPlaneReadSideEffects {
	return controlPlaneReadSideEffects{}
}

func controlPlaneReadCursorListSummaries(lists map[string]controlPlaneReadCursorListFixture, tenantRef string, filters map[string]string) ([]map[string]any, *string) {
	fixture, ok := lists[tenantRef]
	if !ok {
		return []map[string]any{}, nil
	}

	items := make([]map[string]any, 0, len(fixture.items))
	for _, item := range fixture.items {
		if !controlPlaneReadItemMatchesFilters(item, filters) {
			continue
		}
		items = append(items, cloneControlPlaneReadMap(item))
	}
	if len(items) == 0 || fixture.nextCursor == "" {
		return items, nil
	}
	nextCursor := fixture.nextCursor
	return items, &nextCursor
}

func controlPlaneReadItemMatchesFilters(item map[string]any, filters map[string]string) bool {
	for key, expected := range filters {
		value := item[key]
		if key == "scope" {
			value = item["scopes"]
		}
		if !controlPlaneReadValueMatchesFilter(value, expected) {
			return false
		}
	}
	return true
}

func controlPlaneReadValueMatchesFilter(value any, expected string) bool {
	switch typed := value.(type) {
	case string:
		return typed == expected
	case nil:
		return expected == ""
	case []string:
		for _, item := range typed {
			if item == expected {
				return true
			}
		}
		return false
	case []any:
		for _, item := range typed {
			if stringItem, ok := item.(string); ok && stringItem == expected {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func cloneControlPlaneReadMap(input map[string]any) map[string]any {
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = cloneControlPlaneReadValue(value)
	}
	return cloned
}

func cloneControlPlaneReadValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneControlPlaneReadMap(typed)
	case []string:
		cloned := make([]string, len(typed))
		copy(cloned, typed)
		return cloned
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneControlPlaneReadValue(item)
		}
		return cloned
	default:
		return typed
	}
}
