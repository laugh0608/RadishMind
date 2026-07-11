package httpapi

import (
	"context"
	"net/http"
	"testing"

	"radishmind.local/services/platform/internal/bridge"
)

func TestBridgeFailureCodesKeepStableNorthboundStatus(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "timeout",
			err:            context.DeadlineExceeded,
			expectedCode:   bridge.ErrorCodeWorkerTimeout,
			expectedStatus: http.StatusGatewayTimeout,
		},
		{
			name:           "cancellation",
			err:            context.Canceled,
			expectedCode:   bridge.ErrorCodeWorkerCanceled,
			expectedStatus: http.StatusRequestTimeout,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			code := bridgeFailureCode(testCase.err)
			if code != testCase.expectedCode {
				t.Fatalf("unexpected bridge failure code: %s", code)
			}
			if definition := lookupPlatformErrorDefinition(code); definition.statusCode != testCase.expectedStatus {
				t.Fatalf("unexpected bridge failure status: %#v", definition)
			}
		})
	}
}
