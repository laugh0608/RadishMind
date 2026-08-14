package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAdminProviderRouteHTTPV2LifecycleActivationRollbackAndV1Compatibility(t *testing.T) {
	fixture := newAdminProviderRouteHTTPV2Fixture()
	const configurationPath = "/v1/admin/provider-route-configurations/gateway-default"

	v1Input := adminProviderRouteTestDraftInput(0, "mock-primary")
	v1Draft := fixture.serve(t, http.MethodPut, configurationPath, adminProviderRouteDraftPutBody{
		ExpectedRevision: v1Input.ExpectedRevision,
		DisplayName:      v1Input.DisplayName,
		ProviderProfiles: v1Input.ProviderProfiles,
		ModelRoutes:      v1Input.ModelRoutes,
	}, fixture.auth, http.StatusOK)
	if v1Draft.Draft == nil || v1Draft.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersion {
		t.Fatalf("create v1 compatibility draft: %#v", v1Draft)
	}
	v1Candidate := fixture.serve(t, http.MethodPost, configurationPath+"/candidates",
		adminProviderRouteCandidateCreateBody{CandidateID: "candidate-v1", ExpectedDraftRevision: 1},
		fixture.auth, http.StatusCreated)
	if v1Candidate.Candidate == nil || v1Candidate.Candidate.SchemaVersion != adminProviderRouteCandidateSchemaVersion {
		t.Fatalf("create v1 compatibility candidate: %#v", v1Candidate)
	}
	fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v1/reviews",
		adminProviderRouteReviewBody{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Reviewed the original single attempt route.",
		}, fixture.auth, http.StatusOK)
	v1Activation := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v1/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 0,
			Action:             adminProviderRouteActionActivate,
			Reason:             "Activate the reviewed single attempt route.",
		}, fixture.auth, http.StatusOK)
	if v1Activation.Snapshot == nil || v1Activation.Snapshot.Generation != 1 ||
		v1Activation.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersion {
		t.Fatalf("activate v1 compatibility snapshot: %#v", v1Activation)
	}

	v2Input := adminProviderRouteV2TestDraftInput(1)
	v2Input.ConfigurationID = "gateway-default"
	v2Draft := fixture.serve(t, http.MethodPut, configurationPath, adminProviderRouteDraftPutBody{
		ExpectedRevision: v2Input.ExpectedRevision,
		DisplayName:      v2Input.DisplayName,
		ProviderProfiles: v2Input.ProviderProfiles,
		ModelRoutes:      v2Input.ModelRoutes,
	}, fixture.auth, http.StatusOK)
	if v2Draft.Draft == nil || v2Draft.Draft.DraftRevision != 2 ||
		v2Draft.Draft.SchemaVersion != adminProviderRouteDraftSchemaVersionV2 ||
		!adminProviderRouteHTTPAttemptOrder(v2Draft.Draft.ModelRoutes[0], "primary", "secondary") {
		t.Fatalf("create ordered v2 draft: %#v", v2Draft)
	}
	v2Candidate := fixture.serve(t, http.MethodPost, configurationPath+"/candidates",
		adminProviderRouteCandidateCreateBody{CandidateID: "candidate-v2", ExpectedDraftRevision: 2},
		fixture.auth, http.StatusCreated)
	if v2Candidate.Candidate == nil || v2Candidate.Candidate.SchemaVersion != adminProviderRouteCandidateSchemaVersionV2 ||
		len(v2Candidate.Candidate.InventoryBindings) != 2 ||
		v2Candidate.Candidate.InventoryBindings[0].ProfileID != "primary" ||
		v2Candidate.Candidate.InventoryBindings[1].ProfileID != "secondary" ||
		v2Candidate.Candidate.InventoryBindings[0].InventoryDigest == v2Candidate.Candidate.InventoryBindings[1].InventoryDigest {
		t.Fatalf("create v2 candidate with frozen target inventory: %#v", v2Candidate)
	}
	fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v2/reviews",
		adminProviderRouteReviewBody{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Reviewed ordered targets, capabilities, quota and cost risk.",
		}, fixture.auth, http.StatusOK)
	afterReview := fixture.serve(t, http.MethodGet, configurationPath+"/active-snapshot", nil,
		fixture.auth, http.StatusOK)
	if afterReview.Snapshot == nil || afterReview.Snapshot.Generation != 1 ||
		afterReview.Snapshot.CandidateID != "candidate-v1" ||
		afterReview.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersion {
		t.Fatalf("review changed active runtime snapshot: %#v", afterReview)
	}

	v2Activation := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v2/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 1,
			Action:             adminProviderRouteActionActivate,
			Reason:             "Activate the reviewed sequential fallback route.",
		}, fixture.auth, http.StatusOK)
	if v2Activation.Snapshot == nil || v2Activation.Snapshot.Generation != 2 ||
		v2Activation.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersionV2 ||
		!adminProviderRouteHTTPAttemptOrder(v2Activation.Snapshot.Configuration.ModelRoutes[0], "primary", "secondary") {
		t.Fatalf("activate v2 snapshot: %#v", v2Activation)
	}
	staleGeneration := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v1/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 1,
			Action:             adminProviderRouteActionRollback,
			Reason:             "A stale generation must not roll back the active route.",
		}, fixture.auth, http.StatusConflict)
	if staleGeneration.FailureCode == nil ||
		*staleGeneration.FailureCode != AdminProviderRouteFailureGenerationConflict ||
		staleGeneration.CurrentGeneration != 2 {
		t.Fatalf("activation CAS lost current generation: %#v", staleGeneration)
	}
	rollback := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v1/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 2,
			Action:             adminProviderRouteActionRollback,
			Reason:             "Restore the previously active single attempt snapshot.",
		}, fixture.auth, http.StatusOK)
	if rollback.Snapshot == nil || rollback.Snapshot.Generation != 3 ||
		rollback.Snapshot.CandidateID != "candidate-v1" ||
		rollback.Snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersion {
		t.Fatalf("rollback to v1 snapshot: %#v", rollback)
	}
	history := fixture.serve(t, http.MethodGet, configurationPath+"/activation-history", nil,
		fixture.auth, http.StatusOK)
	if len(history.ActivationHistory) != 3 ||
		history.ActivationHistory[0].AfterCandidateID != "candidate-v1" ||
		history.ActivationHistory[1].AfterCandidateID != "candidate-v2" ||
		history.ActivationHistory[2].Action != adminProviderRouteActionRollback ||
		history.ActivationHistory[2].AfterCandidateID != "candidate-v1" {
		t.Fatalf("unexpected mixed-version activation history: %#v", history)
	}
}

func TestAdminProviderRouteHTTPV2RejectsNestedJSONTargetOrderAndCapabilities(t *testing.T) {
	for _, test := range []struct {
		name        string
		invalidJSON bool
		requestBody func(t *testing.T) string
	}{
		{
			name: "unknown nested target field", invalidJSON: true,
			requestBody: func(t *testing.T) string {
				return adminProviderRouteHTTPV2RawBody(t, func(payload string) string {
					return strings.Replace(payload, `"provider_profile_id":"primary"`,
						`"provider_profile_id":"primary","unexpected":true`, 1)
				})
			},
		},
		{
			name: "duplicate nested ordinal field", invalidJSON: true,
			requestBody: func(t *testing.T) string {
				return adminProviderRouteHTTPV2RawBody(t, func(payload string) string {
					return strings.Replace(payload, `"ordinal":1`, `"ordinal":1,"ordinal":2`, 1)
				})
			},
		},
		{
			name: "target array order differs from ordinal",
			requestBody: func(t *testing.T) string {
				input := adminProviderRouteV2TestDraftInput(0)
				input.ModelRoutes[0].AttemptTargets[0], input.ModelRoutes[0].AttemptTargets[1] =
					input.ModelRoutes[0].AttemptTargets[1], input.ModelRoutes[0].AttemptTargets[0]
				return adminProviderRouteHTTPDraftBodyJSON(t, input)
			},
		},
		{
			name: "backup profile lacks route capability",
			requestBody: func(t *testing.T) string {
				input := adminProviderRouteV2TestDraftInput(0)
				input.ProviderProfiles[1].Capabilities = []string{"responses"}
				return adminProviderRouteHTTPDraftBodyJSON(t, input)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newAdminProviderRouteHTTPV2Fixture()
			recorder := fixture.rawServe(t, http.MethodPut,
				"/v1/admin/provider-route-configurations/gateway-fallback", test.requestBody(t), fixture.auth)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("unexpected strict payload status: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if test.invalidJSON {
				if !strings.Contains(recorder.Body.String(), "INVALID_JSON") {
					t.Fatalf("strict decoder did not reject nested JSON: %s", recorder.Body.String())
				}
			} else {
				envelope := decodeAdminProviderRouteEnvelope(t, recorder, http.StatusBadRequest)
				if envelope.FailureCode == nil || *envelope.FailureCode != AdminProviderRouteFailurePayloadInvalid {
					t.Fatalf("unexpected route contract failure: %#v", envelope)
				}
			}
			if len(fixture.repository.drafts) != 0 || fixture.bridge.inventoryCalls.Load() != 0 {
				t.Fatalf("invalid v2 payload reached owner side effects: drafts=%d inventory_calls=%d",
					len(fixture.repository.drafts), fixture.bridge.inventoryCalls.Load())
			}
		})
	}
}

func TestAdminProviderRouteHTTPV2CASAndInventoryDriftFailClosed(t *testing.T) {
	fixture := newAdminProviderRouteHTTPV2Fixture()
	const configurationPath = "/v1/admin/provider-route-configurations/gateway-fallback"
	input := adminProviderRouteV2TestDraftInput(0)
	body := adminProviderRouteDraftPutBody{
		ExpectedRevision: input.ExpectedRevision,
		DisplayName:      input.DisplayName,
		ProviderProfiles: input.ProviderProfiles,
		ModelRoutes:      input.ModelRoutes,
	}
	fixture.serve(t, http.MethodPut, configurationPath, body, fixture.auth, http.StatusOK)
	staleDraft := fixture.serve(t, http.MethodPut, configurationPath, body, fixture.auth, http.StatusConflict)
	if staleDraft.FailureCode == nil || *staleDraft.FailureCode != AdminProviderRouteFailureDraftRevisionConflict ||
		staleDraft.CurrentDraftRevision != 1 {
		t.Fatalf("v2 draft CAS lost current revision: %#v", staleDraft)
	}
	fixture.serve(t, http.MethodPost, configurationPath+"/candidates",
		adminProviderRouteCandidateCreateBody{CandidateID: "candidate-v2-cas", ExpectedDraftRevision: 1},
		fixture.auth, http.StatusCreated)
	reviewBody := adminProviderRouteReviewBody{
		ExpectedReviewVersion: 0,
		Decision:              adminProviderRouteDecisionApprove,
		Reason:                "Reviewed frozen target inventory before the CAS check.",
	}
	fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v2-cas/reviews",
		reviewBody, fixture.auth, http.StatusOK)
	staleReview := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v2-cas/reviews",
		reviewBody, fixture.auth, http.StatusConflict)
	if staleReview.FailureCode == nil || *staleReview.FailureCode != AdminProviderRouteFailureReviewVersionConflict ||
		staleReview.CurrentReviewVersion != 1 ||
		staleReview.CurrentCandidateState != adminProviderRouteCandidateApproved {
		t.Fatalf("v2 review CAS lost current labels: %#v", staleReview)
	}

	drifted := fixture.bridge.inventory.Profiles[1]
	drifted.ResolvedModel = "mock-model-drifted"
	fixture.bridge.inventory.Profiles[1] = drifted
	blocked := fixture.serve(t, http.MethodPost, configurationPath+"/candidates/candidate-v2-cas/activations",
		adminProviderRouteActivationBody{
			ExpectedGeneration: 0,
			Action:             adminProviderRouteActionActivate,
			Reason:             "Activation must reject the changed backup inventory digest.",
		}, fixture.auth, http.StatusUnprocessableEntity)
	if blocked.FailureCode == nil || *blocked.FailureCode != AdminProviderRouteFailureInventoryMismatch ||
		blocked.Snapshot != nil || blocked.Activation != nil {
		t.Fatalf("inventory drift did not fail closed: %#v", blocked)
	}
	active := fixture.serve(t, http.MethodGet, configurationPath+"/active-snapshot", nil,
		fixture.auth, http.StatusOK)
	history := fixture.serve(t, http.MethodGet, configurationPath+"/activation-history", nil,
		fixture.auth, http.StatusOK)
	if active.Snapshot != nil || active.CurrentGeneration != 0 || len(history.ActivationHistory) != 0 {
		t.Fatalf("blocked activation changed runtime state: active=%#v history=%#v", active, history)
	}
}

func newAdminProviderRouteHTTPV2Fixture() adminProviderRouteHTTPFixture {
	fixture := newAdminProviderRouteHTTPFixture()
	secondary := adminProviderRouteBridgeProfile()
	secondary.Profile = "mock-secondary"
	secondary.NormalizedProfile = "mock-secondary"
	fixture.bridge.inventory.Profiles = append(fixture.bridge.inventory.Profiles, secondary)
	return fixture
}

func adminProviderRouteHTTPAttemptOrder(route AdminModelRouteDefinition, profileIDs ...string) bool {
	if route.ExecutionMode != AdminProviderRouteExecutionSequentialFallback ||
		len(route.AttemptTargets) != len(profileIDs) {
		return false
	}
	for index, profileID := range profileIDs {
		if route.AttemptTargets[index].Ordinal != index+1 ||
			route.AttemptTargets[index].ProviderProfileID != profileID {
			return false
		}
	}
	return true
}

func adminProviderRouteHTTPV2RawBody(t *testing.T, mutate func(string) string) string {
	t.Helper()
	return mutate(adminProviderRouteHTTPDraftBodyJSON(t, adminProviderRouteV2TestDraftInput(0)))
}

func adminProviderRouteHTTPDraftBodyJSON(t *testing.T, input AdminProviderRouteDraftInput) string {
	t.Helper()
	payload, err := json.Marshal(adminProviderRouteDraftPutBody{
		ExpectedRevision: input.ExpectedRevision,
		DisplayName:      input.DisplayName,
		ProviderProfiles: input.ProviderProfiles,
		ModelRoutes:      input.ModelRoutes,
	})
	if err != nil {
		t.Fatalf("marshal admin provider route draft body: %v", err)
	}
	return string(payload)
}
