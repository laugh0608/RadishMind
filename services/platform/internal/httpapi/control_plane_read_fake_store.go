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
				"tenant_display_name":   "Demo Tenant",
				"tenant_state":          "active",
				"plan_ref":              "plan_dev",
				"quota_summary_ref":     "quota_demo_current",
				"deployment_status_ref": "deployment_local_boundary",
				"audit_summary_ref":     "audit_tenant_demo_latest",
			},
		},
		applicationSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"application_ref":                "app_demo_workflow",
						"tenant_ref":                     "tenant_demo",
						"application_kind":               "workflow",
						"display_name":                   "Demo workflow",
						"owner_subject_ref":              "subject_demo_user",
						"latest_workflow_definition_ref": "wf_demo_v1",
						"last_run_status":                "succeeded",
						"updated_at":                     "2026-05-27T00:00:00Z",
					},
				},
				nextCursor: "cursor_apps_next",
			},
		},
		apiKeySummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"api_key_id":        "ak_demo_001",
						"tenant_ref":        "tenant_demo",
						"owner_subject_ref": "subject_demo_user",
						"scopes": []string{
							"models:read",
							"chat:invoke",
							"usage:read",
						},
						"state":        "active",
						"created_at":   "2026-05-27T00:00:00Z",
						"expires_at":   "2026-06-27T00:00:00Z",
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
						"workflow_definition_id":        "wf_demo_v1",
						"tenant_ref":                    "tenant_demo",
						"application_ref":               "app_demo_workflow",
						"version":                       1,
						"definition_status":             "draft",
						"node_count":                    4,
						"risk_level":                    "medium",
						"requires_confirmation_capable": true,
						"updated_at":                    "2026-05-27T00:00:00Z",
					},
				},
				nextCursor: "cursor_workflows_next",
			},
		},
		runRecordSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"run_id":                 "run_demo_001",
						"tenant_ref":             "tenant_demo",
						"workflow_definition_id": "wf_demo_v1",
						"application_ref":        "app_demo_workflow",
						"status":                 "succeeded",
						"failure_code":           nil,
						"cost_summary": map[string]any{
							"estimated_cost": 0.08,
							"currency":       "USD",
						},
						"trace_id":     "trace_demo_run_001",
						"started_at":   "2026-05-27T00:00:00Z",
						"completed_at": "2026-05-27T00:00:02Z",
					},
				},
				nextCursor: "cursor_runs_next",
			},
		},
		auditSummaries: map[string]controlPlaneReadCursorListFixture{
			"tenant_demo": {
				items: []map[string]any{
					{
						"audit_ref":         "audit_read_runs_001",
						"tenant_ref":        "tenant_demo",
						"actor_subject_ref": "subject_demo_admin",
						"event_kind":        "read_model_accessed",
						"resource_ref":      "run_demo_001",
						"decision":          "allowed",
						"failure_code":      nil,
						"trace_id":          "trace_demo_run_001",
						"recorded_at":       "2026-05-27T00:00:03Z",
					},
				},
				nextCursor: "cursor_audit_next",
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
