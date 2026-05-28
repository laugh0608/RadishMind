package httpapi

type controlPlaneReadStore interface {
	TenantSummary(tenantRef string) (map[string]any, bool)
	QuotaSummary(tenantRef string) (map[string]any, bool)
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
	tenantSummaries map[string]map[string]any
	quotaSummaries  map[string]map[string]any
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

func (store fixtureBackedControlPlaneReadStore) QuotaSummary(tenantRef string) (map[string]any, bool) {
	item, ok := store.quotaSummaries[tenantRef]
	if !ok {
		return nil, false
	}
	return cloneControlPlaneReadMap(item), true
}

func (store fixtureBackedControlPlaneReadStore) SideEffects() controlPlaneReadSideEffects {
	return controlPlaneReadSideEffects{}
}

func cloneControlPlaneReadMap(input map[string]any) map[string]any {
	cloned := make(map[string]any, len(input))
	for key, value := range input {
		if nested, ok := value.(map[string]any); ok {
			cloned[key] = cloneControlPlaneReadMap(nested)
			continue
		}
		cloned[key] = value
	}
	return cloned
}
