package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	GatewayRequestQuotaSchemaVersion = "gateway_request_quota_v1"
	GatewayRequestQuotaPeriod        = "calendar_day_utc"

	GatewayRequestQuotaFailureDisabled              = "gateway_quota_disabled"
	GatewayRequestQuotaFailureScopeDenied           = "gateway_quota_scope_denied"
	GatewayRequestQuotaFailureEnvironmentForbidden  = "gateway_quota_environment_forbidden"
	GatewayRequestQuotaFailurePayloadInvalid        = "gateway_quota_payload_invalid"
	GatewayRequestQuotaFailurePolicyNotFound        = "gateway_quota_policy_not_found"
	GatewayRequestQuotaFailurePolicyVersionConflict = "gateway_quota_policy_version_conflict"
	GatewayRequestQuotaFailureAttemptConflict       = "gateway_quota_attempt_conflict"
	GatewayRequestQuotaFailureExceeded              = "gateway_quota_exceeded"
	GatewayRequestQuotaFailureStoreUnavailable      = "gateway_quota_store_unavailable"

	minimumGatewayRequestQuotaLimit = 1
	maximumGatewayRequestQuotaLimit = 1_000_000
)

var (
	errGatewayRequestQuotaContract              = errors.New("gateway request quota contract mismatch")
	errGatewayRequestQuotaPolicyNotFound        = errors.New("gateway request quota policy not found")
	errGatewayRequestQuotaPolicyVersionConflict = errors.New("gateway request quota policy version conflict")
	errGatewayRequestQuotaAttemptConflict       = errors.New("gateway request quota attempt conflict")
	errGatewayRequestQuotaExceeded              = errors.New("gateway request quota exceeded")
	errGatewayRequestQuotaStoreUnavailable      = errors.New("gateway request quota store unavailable")

	gatewayRequestQuotaIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)
	gatewayRequestQuotaRoutePattern      = regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH) /[A-Za-z0-9._~:/{}-]{1,240}$`)
)

type GatewayRequestQuotaContext struct {
	RequestContext context.Context `json:"-"`
	TenantRef      string          `json:"tenant_ref"`
	WorkspaceID    string          `json:"workspace_id"`
	Environment    string          `json:"environment"`
	ApplicationID  string          `json:"application_id"`
	ActorRef       string          `json:"actor_ref"`
	RequestID      string          `json:"request_id"`
	AuditRef       string          `json:"audit_ref"`
}

type GatewayRequestQuotaPolicy struct {
	SchemaVersion string    `json:"schema_version"`
	PolicyID      string    `json:"policy_id"`
	TenantRef     string    `json:"tenant_ref"`
	WorkspaceID   string    `json:"workspace_id"`
	Environment   string    `json:"environment"`
	ApplicationID string    `json:"application_id"`
	Period        string    `json:"period"`
	RequestLimit  int64     `json:"request_limit"`
	RecordVersion int64     `json:"record_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedBy     string    `json:"updated_by"`
	LastRequestID string    `json:"last_request_id"`
	LastAuditRef  string    `json:"last_audit_ref"`
}

type GatewayRequestQuotaUsage struct {
	SchemaVersion         string    `json:"schema_version"`
	TenantRef             string    `json:"tenant_ref"`
	WorkspaceID           string    `json:"workspace_id"`
	Environment           string    `json:"environment"`
	ApplicationID         string    `json:"application_id"`
	Period                string    `json:"period"`
	PeriodStart           string    `json:"period_start"`
	PolicyID              string    `json:"policy_id"`
	PolicyVersion         int64     `json:"policy_version"`
	RequestLimit          int64     `json:"request_limit"`
	AdmittedRequestCount  int64     `json:"admitted_request_count"`
	RemainingRequestCount int64     `json:"remaining_request_count"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type GatewayRequestQuotaAdmissionInput struct {
	APIKeyID   string
	RequestID  string
	Route      string
	AdmittedAt time.Time
}

type GatewayRequestQuotaAdmissionDecision struct {
	SchemaVersion string                   `json:"schema_version"`
	AdmissionID   string                   `json:"admission_id"`
	APIKeyID      string                   `json:"api_key_id"`
	RequestID     string                   `json:"request_id"`
	Route         string                   `json:"route"`
	AdmittedAt    time.Time                `json:"admitted_at"`
	PolicyID      string                   `json:"policy_id"`
	PolicyVersion int64                    `json:"policy_version"`
	Usage         GatewayRequestQuotaUsage `json:"usage"`
}

type GatewayRequestQuotaRepository interface {
	ReadPolicy(GatewayRequestQuotaContext) (GatewayRequestQuotaPolicy, bool, error)
	PutPolicy(GatewayRequestQuotaContext, int64, int64, time.Time) (GatewayRequestQuotaPolicy, error)
	ReadUsage(GatewayRequestQuotaContext, string) (GatewayRequestQuotaUsage, bool, error)
	AdmitProviderAttempt(GatewayRequestQuotaContext, GatewayRequestQuotaAdmissionInput) (GatewayRequestQuotaAdmissionDecision, error)
}

type memoryGatewayRequestQuotaRepository struct {
	mu         sync.Mutex
	policies   map[string]GatewayRequestQuotaPolicy
	usage      map[string]GatewayRequestQuotaUsage
	admissions map[string]GatewayRequestQuotaAdmissionDecision
}

func newMemoryGatewayRequestQuotaRepository() *memoryGatewayRequestQuotaRepository {
	return &memoryGatewayRequestQuotaRepository{
		policies:   make(map[string]GatewayRequestQuotaPolicy),
		usage:      make(map[string]GatewayRequestQuotaUsage),
		admissions: make(map[string]GatewayRequestQuotaAdmissionDecision),
	}
}

func (repository *memoryGatewayRequestQuotaRepository) ReadPolicy(
	quotaContext GatewayRequestQuotaContext,
) (GatewayRequestQuotaPolicy, bool, error) {
	if !validGatewayRequestQuotaContext(quotaContext) {
		return GatewayRequestQuotaPolicy{}, false, errGatewayRequestQuotaContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	policy, found := repository.policies[gatewayRequestQuotaScopeKey(quotaContext)]
	return policy, found, nil
}

func (repository *memoryGatewayRequestQuotaRepository) PutPolicy(
	quotaContext GatewayRequestQuotaContext,
	expectedVersion int64,
	requestLimit int64,
	now time.Time,
) (GatewayRequestQuotaPolicy, error) {
	if !validGatewayRequestQuotaContext(quotaContext) || expectedVersion < 0 ||
		requestLimit < minimumGatewayRequestQuotaLimit || requestLimit > maximumGatewayRequestQuotaLimit || now.IsZero() {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaContract
	}
	now = now.UTC()
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := gatewayRequestQuotaScopeKey(quotaContext)
	current, found := repository.policies[key]
	if (!found && expectedVersion != 0) || (found && current.RecordVersion != expectedVersion) {
		return GatewayRequestQuotaPolicy{}, errGatewayRequestQuotaPolicyVersionConflict
	}
	policy := current
	if !found {
		policy = GatewayRequestQuotaPolicy{
			SchemaVersion: GatewayRequestQuotaSchemaVersion,
			PolicyID:      gatewayRequestQuotaPolicyID(quotaContext),
			TenantRef:     quotaContext.TenantRef,
			WorkspaceID:   quotaContext.WorkspaceID,
			Environment:   quotaContext.Environment,
			ApplicationID: quotaContext.ApplicationID,
			Period:        GatewayRequestQuotaPeriod,
			CreatedAt:     now,
			CreatedBy:     quotaContext.ActorRef,
		}
	}
	policy.RequestLimit = requestLimit
	policy.RecordVersion = expectedVersion + 1
	policy.UpdatedAt = now
	policy.UpdatedBy = quotaContext.ActorRef
	policy.LastRequestID = quotaContext.RequestID
	policy.LastAuditRef = quotaContext.AuditRef
	repository.policies[key] = policy
	return policy, nil
}

func (repository *memoryGatewayRequestQuotaRepository) ReadUsage(
	quotaContext GatewayRequestQuotaContext,
	periodStart string,
) (GatewayRequestQuotaUsage, bool, error) {
	if !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaPeriodStart(periodStart) {
		return GatewayRequestQuotaUsage{}, false, errGatewayRequestQuotaContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	usage, found := repository.usage[gatewayRequestQuotaUsageKey(quotaContext, periodStart)]
	if !found {
		return GatewayRequestQuotaUsage{}, false, nil
	}
	if policy, policyFound := repository.policies[gatewayRequestQuotaScopeKey(quotaContext)]; policyFound {
		usage.PolicyID = policy.PolicyID
		usage.PolicyVersion = policy.RecordVersion
		usage.RequestLimit = policy.RequestLimit
		usage.RemainingRequestCount = maxInt64(policy.RequestLimit-usage.AdmittedRequestCount, 0)
	}
	return usage, true, nil
}

func (repository *memoryGatewayRequestQuotaRepository) AdmitProviderAttempt(
	quotaContext GatewayRequestQuotaContext,
	input GatewayRequestQuotaAdmissionInput,
) (GatewayRequestQuotaAdmissionDecision, error) {
	if !validGatewayRequestQuotaContext(quotaContext) || !validGatewayRequestQuotaAdmissionInput(input) {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	scopeKey := gatewayRequestQuotaScopeKey(quotaContext)
	policy, found := repository.policies[scopeKey]
	if !found {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaPolicyNotFound
	}
	admissionKey := gatewayRequestQuotaAdmissionKey(quotaContext, input.RequestID)
	if _, found := repository.admissions[admissionKey]; found {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaAttemptConflict
	}
	periodStart := gatewayRequestQuotaPeriodStart(input.AdmittedAt)
	usageKey := gatewayRequestQuotaUsageKey(quotaContext, periodStart)
	usage := repository.usage[usageKey]
	if usage.AdmittedRequestCount >= policy.RequestLimit {
		return GatewayRequestQuotaAdmissionDecision{}, errGatewayRequestQuotaExceeded
	}
	usage = GatewayRequestQuotaUsage{
		SchemaVersion:        GatewayRequestQuotaSchemaVersion,
		TenantRef:            quotaContext.TenantRef,
		WorkspaceID:          quotaContext.WorkspaceID,
		Environment:          quotaContext.Environment,
		ApplicationID:        quotaContext.ApplicationID,
		Period:               GatewayRequestQuotaPeriod,
		PeriodStart:          periodStart,
		PolicyID:             policy.PolicyID,
		PolicyVersion:        policy.RecordVersion,
		RequestLimit:         policy.RequestLimit,
		AdmittedRequestCount: usage.AdmittedRequestCount + 1,
		UpdatedAt:            input.AdmittedAt.UTC(),
	}
	usage.RemainingRequestCount = policy.RequestLimit - usage.AdmittedRequestCount
	decision := GatewayRequestQuotaAdmissionDecision{
		SchemaVersion: GatewayRequestQuotaSchemaVersion,
		AdmissionID:   gatewayRequestQuotaAdmissionID(quotaContext, input.RequestID),
		APIKeyID:      input.APIKeyID,
		RequestID:     input.RequestID,
		Route:         input.Route,
		AdmittedAt:    input.AdmittedAt.UTC(),
		PolicyID:      policy.PolicyID,
		PolicyVersion: policy.RecordVersion,
		Usage:         usage,
	}
	repository.usage[usageKey] = usage
	repository.admissions[admissionKey] = decision
	return decision, nil
}

func validGatewayRequestQuotaContext(quotaContext GatewayRequestQuotaContext) bool {
	if quotaContext.RequestContext == nil || quotaContext.Environment != "development" && quotaContext.Environment != "test" {
		return false
	}
	for _, value := range []string{
		quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.ApplicationID,
		quotaContext.ActorRef, quotaContext.RequestID, quotaContext.AuditRef,
	} {
		if !gatewayRequestQuotaIdentifierPattern.MatchString(strings.TrimSpace(value)) {
			return false
		}
	}
	return true
}

func validGatewayRequestQuotaAdmissionInput(input GatewayRequestQuotaAdmissionInput) bool {
	if input.AdmittedAt.IsZero() || !gatewayRequestQuotaRoutePattern.MatchString(strings.TrimSpace(input.Route)) {
		return false
	}
	for _, value := range []string{input.APIKeyID, input.RequestID} {
		if !gatewayRequestQuotaIdentifierPattern.MatchString(strings.TrimSpace(value)) {
			return false
		}
	}
	return true
}

func validGatewayRequestQuotaPeriodStart(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func gatewayRequestQuotaPeriodStart(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

func gatewayRequestQuotaScopeKey(quotaContext GatewayRequestQuotaContext) string {
	return strings.Join([]string{quotaContext.TenantRef, quotaContext.WorkspaceID, quotaContext.Environment, quotaContext.ApplicationID}, "\x1f")
}

func gatewayRequestQuotaUsageKey(quotaContext GatewayRequestQuotaContext, periodStart string) string {
	return gatewayRequestQuotaScopeKey(quotaContext) + "\x1f" + periodStart
}

func gatewayRequestQuotaAdmissionKey(quotaContext GatewayRequestQuotaContext, requestID string) string {
	return gatewayRequestQuotaScopeKey(quotaContext) + "\x1f" + requestID
}

func gatewayRequestQuotaPolicyID(quotaContext GatewayRequestQuotaContext) string {
	return "quota_" + gatewayRequestQuotaDigest(gatewayRequestQuotaScopeKey(quotaContext))
}

func gatewayRequestQuotaAdmissionID(quotaContext GatewayRequestQuotaContext, requestID string) string {
	return "qadm_" + gatewayRequestQuotaDigest(gatewayRequestQuotaAdmissionKey(quotaContext, requestID))
}

func gatewayRequestQuotaDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:12])
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
