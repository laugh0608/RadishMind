package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type controlPlaneReadResponseFixtureForTest struct {
	ResponseExamples []controlPlaneReadResponseExampleForTest `json:"response_examples"`
}

type controlPlaneReadResponseExampleForTest struct {
	RouteID string                                 `json:"route_id"`
	Success controlPlaneReadResponseSuccessForTest `json:"success"`
}

type controlPlaneReadResponseSuccessForTest struct {
	Items      []map[string]any `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

func TestControlPlaneReadFakeStoreMatchesResponseFixture(t *testing.T) {
	store := newControlPlaneReadFakeStore()
	fixture := loadControlPlaneReadResponseFixtureForTest(t)

	for _, routeID := range []string{
		"tenant-summary-route",
		"application-summary-list-route",
		"api-key-summary-list-route",
		"quota-summary-route",
		"workflow-definition-summary-list-route",
		"run-record-summary-list-route",
		"audit-summary-list-route",
	} {
		t.Run(routeID, func(t *testing.T) {
			expected, ok := fixture[routeID]
			if !ok {
				t.Fatalf("missing response fixture route: %s", routeID)
			}
			actualItems, actualCursor := controlPlaneReadFakeStoreRouteSamplesForTest(t, store, routeID)
			assertControlPlaneReadProductSamplesForTest(t, routeID, actualItems, actualCursor, expected)
		})
	}
}

func loadControlPlaneReadResponseFixtureForTest(t *testing.T) map[string]controlPlaneReadResponseSuccessForTest {
	t.Helper()

	path := filepath.Join(
		"..",
		"..",
		"..",
		"..",
		"scripts",
		"checks",
		"fixtures",
		"control-plane-read-response-fixtures-v1.json",
	)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read response fixture: %v", err)
	}

	var fixture controlPlaneReadResponseFixtureForTest
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatalf("decode response fixture: %v", err)
	}

	examples := make(map[string]controlPlaneReadResponseSuccessForTest, len(fixture.ResponseExamples))
	for _, example := range fixture.ResponseExamples {
		examples[example.RouteID] = example.Success
	}
	return examples
}

func controlPlaneReadFakeStoreRouteSamplesForTest(
	t *testing.T,
	store controlPlaneReadStore,
	routeID string,
) ([]map[string]any, *string) {
	t.Helper()

	switch routeID {
	case "tenant-summary-route":
		item, ok := store.TenantSummary("tenant_demo")
		if !ok {
			t.Fatalf("tenant summary fixture missing tenant_demo")
		}
		return []map[string]any{item}, nil
	case "application-summary-list-route":
		return store.ApplicationSummaries("tenant_demo", nil)
	case "api-key-summary-list-route":
		return store.APIKeySummaries("tenant_demo", nil)
	case "quota-summary-route":
		item, ok := store.QuotaSummary("tenant_demo")
		if !ok {
			t.Fatalf("quota summary fixture missing tenant_demo")
		}
		return []map[string]any{item}, nil
	case "workflow-definition-summary-list-route":
		return store.WorkflowDefinitionSummaries("tenant_demo", nil)
	case "run-record-summary-list-route":
		return store.RunRecordSummaries("tenant_demo", nil)
	case "audit-summary-list-route":
		return store.AuditSummaries("tenant_demo", nil)
	default:
		t.Fatalf("unsupported control plane read route: %s", routeID)
		return nil, nil
	}
}

func assertControlPlaneReadProductSamplesForTest(
	t *testing.T,
	routeID string,
	actualItems []map[string]any,
	actualCursor *string,
	expected controlPlaneReadResponseSuccessForTest,
) {
	t.Helper()

	if !reflect.DeepEqual(normalizedControlPlaneReadItemsForTest(actualItems), normalizedControlPlaneReadItemsForTest(expected.Items)) {
		t.Fatalf("%s product samples drifted:\nactual=%#v\nexpected=%#v", routeID, actualItems, expected.Items)
	}
	assertControlPlaneReadNextCursor(t, actualCursor, expected.NextCursor)
}

func normalizedControlPlaneReadItemsForTest(items []map[string]any) []map[string]any {
	payload, err := json.Marshal(items)
	if err != nil {
		panic(err)
	}
	var normalized []map[string]any
	if err := json.Unmarshal(payload, &normalized); err != nil {
		panic(err)
	}
	return normalized
}
