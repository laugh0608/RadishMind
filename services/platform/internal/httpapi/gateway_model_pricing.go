package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GatewayModelPricingPolicySchemaVersion       = "gateway_model_pricing_policy.v1"
	GatewayModelPricingCurrency                  = "USD"
	GatewayModelPricingTokenUnit           int64 = 1_000_000

	GatewayModelPricingFailureDisabled             = "gateway_pricing_disabled"
	GatewayModelPricingFailureScopeDenied          = "gateway_pricing_scope_denied"
	GatewayModelPricingFailureEnvironmentForbidden = "gateway_pricing_environment_forbidden"
	GatewayModelPricingFailurePayloadInvalid       = "gateway_pricing_payload_invalid"
	GatewayModelPricingFailurePolicyNotFound       = "gateway_pricing_policy_not_found"
	GatewayModelPricingFailureVersionConflict      = "gateway_pricing_policy_version_conflict"
	GatewayModelPricingFailureScopeConflict        = "gateway_pricing_policy_scope_conflict"
	GatewayModelPricingFailureStoreUnavailable     = "gateway_pricing_store_unavailable"

	GatewayRequestCostEstimateSchemaVersion = "gateway_request_cost_estimate.v1"
	GatewayRequestCostEstimated             = "estimated"
	GatewayRequestCostUsageNotReported      = "usage_not_reported"
	GatewayRequestCostPriceNotConfigured    = "price_not_configured"
	GatewayRequestCostPriceUnavailable      = "price_unavailable"
	GatewayRequestCostNotApplicable         = "not_applicable"
	GatewayRequestCostLegacyNotCaptured     = "legacy_not_captured"
	GatewayRequestCostRoundingHalfUp        = "half_up_to_currency_micro"

	GatewayPricingSnapshotConfigured    = "configured"
	GatewayPricingSnapshotNotConfigured = "not_configured"
	GatewayPricingSnapshotUnavailable   = "unavailable"
)

var (
	errGatewayModelPricingContract         = errors.New("gateway model pricing contract mismatch")
	errGatewayModelPricingPolicyNotFound   = errors.New("gateway model pricing policy not found")
	errGatewayModelPricingVersionConflict  = errors.New("gateway model pricing policy version conflict")
	errGatewayModelPricingStoreUnavailable = errors.New("gateway model pricing store unavailable")

	gatewayModelPricingIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)
	gatewayModelPricingDigestPattern     = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

type GatewayModelPricingContext struct {
	RequestContext context.Context `json:"-"`
	TenantRef      string          `json:"tenant_ref"`
	WorkspaceID    string          `json:"workspace_id"`
	Environment    string          `json:"environment"`
	ProviderID     string          `json:"provider_id"`
	ProfileID      string          `json:"profile_id"`
	ModelID        string          `json:"model_id"`
	ActorRef       string          `json:"actor_ref"`
	RequestID      string          `json:"request_id"`
	AuditRef       string          `json:"audit_ref"`
}

type GatewayModelPricingPolicy struct {
	SchemaVersion                 string    `json:"schema_version"`
	PolicyID                      string    `json:"policy_id"`
	RecordVersion                 int64     `json:"record_version"`
	TenantRef                     string    `json:"tenant_ref"`
	WorkspaceID                   string    `json:"workspace_id"`
	Environment                   string    `json:"environment"`
	ProviderID                    string    `json:"provider_id"`
	ProfileID                     string    `json:"profile_id"`
	ModelID                       string    `json:"model_id"`
	Currency                      string    `json:"currency"`
	TokenUnit                     int64     `json:"token_unit"`
	InputPriceMicrosPerTokenUnit  int64     `json:"input_price_micros_per_token_unit"`
	OutputPriceMicrosPerTokenUnit int64     `json:"output_price_micros_per_token_unit"`
	PolicyDigest                  string    `json:"policy_digest"`
	Reason                        string    `json:"reason"`
	UpdatedAt                     time.Time `json:"updated_at"`
	UpdatedByActorRef             string    `json:"updated_by_actor_ref"`
	RequestID                     string    `json:"request_id"`
	AuditRef                      string    `json:"audit_ref"`
}

type GatewayModelPricingPolicyInput struct {
	ExpectedVersion               int64
	Currency                      string
	InputPriceMicrosPerTokenUnit  int64
	OutputPriceMicrosPerTokenUnit int64
	Reason                        string
}

type GatewayModelPricingRepository interface {
	ReadCurrent(GatewayModelPricingContext) (GatewayModelPricingPolicy, bool, error)
	ReadRevision(GatewayModelPricingContext, int64) (GatewayModelPricingPolicy, bool, error)
	PutRevision(GatewayModelPricingContext, GatewayModelPricingPolicyInput, time.Time) (GatewayModelPricingPolicy, error)
}

type GatewayModelPricingResult struct {
	Policy         *GatewayModelPricingPolicy
	FailureCode    string
	CurrentVersion int64
}

type gatewayModelPricingService struct {
	repository GatewayModelPricingRepository
	now        func() time.Time
}

func newGatewayModelPricingService(repository GatewayModelPricingRepository) gatewayModelPricingService {
	return gatewayModelPricingService{
		repository: repository,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (service gatewayModelPricingService) ReadCurrent(
	pricingContext GatewayModelPricingContext,
) GatewayModelPricingResult {
	if failureCode := gatewayModelPricingContextFailure(pricingContext); failureCode != "" {
		return GatewayModelPricingResult{FailureCode: failureCode}
	}
	if service.repository == nil {
		return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailureStoreUnavailable}
	}
	policy, found, err := service.repository.ReadCurrent(pricingContext)
	if err != nil {
		return GatewayModelPricingResult{FailureCode: gatewayModelPricingRepositoryFailure(err)}
	}
	if !found {
		return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailurePolicyNotFound}
	}
	if !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailureStoreUnavailable}
	}
	return GatewayModelPricingResult{Policy: &policy, CurrentVersion: policy.RecordVersion}
}

func (service gatewayModelPricingService) PutRevision(
	pricingContext GatewayModelPricingContext,
	input GatewayModelPricingPolicyInput,
) GatewayModelPricingResult {
	if failureCode := gatewayModelPricingContextFailure(pricingContext); failureCode != "" {
		return GatewayModelPricingResult{FailureCode: failureCode}
	}
	input = normalizeGatewayModelPricingPolicyInput(input)
	if !validGatewayModelPricingPolicyInput(input) {
		return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailurePayloadInvalid}
	}
	if service.repository == nil {
		return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailureStoreUnavailable}
	}
	policy, err := service.repository.PutRevision(pricingContext, input, service.now())
	if err == nil {
		if !validGatewayModelPricingPolicy(pricingContext, policy) {
			return GatewayModelPricingResult{FailureCode: GatewayModelPricingFailureStoreUnavailable}
		}
		return GatewayModelPricingResult{Policy: &policy, CurrentVersion: policy.RecordVersion}
	}
	result := GatewayModelPricingResult{FailureCode: gatewayModelPricingRepositoryFailure(err)}
	if errors.Is(err, errGatewayModelPricingVersionConflict) {
		if current, found, readErr := service.repository.ReadCurrent(pricingContext); readErr == nil && found {
			result.CurrentVersion = current.RecordVersion
		}
	}
	return result
}

func gatewayModelPricingContextFailure(pricingContext GatewayModelPricingContext) string {
	if pricingContext.Environment != "development" && pricingContext.Environment != "test" {
		return GatewayModelPricingFailureEnvironmentForbidden
	}
	if !validGatewayModelPricingContext(pricingContext) {
		return GatewayModelPricingFailureScopeConflict
	}
	return ""
}

func gatewayModelPricingRepositoryFailure(err error) string {
	switch {
	case errors.Is(err, errGatewayModelPricingPolicyNotFound):
		return GatewayModelPricingFailurePolicyNotFound
	case errors.Is(err, errGatewayModelPricingVersionConflict):
		return GatewayModelPricingFailureVersionConflict
	default:
		return GatewayModelPricingFailureStoreUnavailable
	}
}

type memoryGatewayModelPricingRepository struct {
	mu        sync.RWMutex
	current   map[string]int64
	revisions map[string]map[int64]GatewayModelPricingPolicy
}

func newMemoryGatewayModelPricingRepository() *memoryGatewayModelPricingRepository {
	return &memoryGatewayModelPricingRepository{
		current:   make(map[string]int64),
		revisions: make(map[string]map[int64]GatewayModelPricingPolicy),
	}
}

func (repository *memoryGatewayModelPricingRepository) ReadCurrent(
	pricingContext GatewayModelPricingContext,
) (GatewayModelPricingPolicy, bool, error) {
	if !validGatewayModelPricingContext(pricingContext) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	key := gatewayModelPricingScopeKey(pricingContext)
	version, found := repository.current[key]
	if !found {
		return GatewayModelPricingPolicy{}, false, nil
	}
	policy, found := repository.revisions[key][version]
	if !found || !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	return policy, true, nil
}

func (repository *memoryGatewayModelPricingRepository) ReadRevision(
	pricingContext GatewayModelPricingContext,
	recordVersion int64,
) (GatewayModelPricingPolicy, bool, error) {
	if !validGatewayModelPricingContext(pricingContext) || recordVersion < 1 {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingContract
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	policy, found := repository.revisions[gatewayModelPricingScopeKey(pricingContext)][recordVersion]
	if !found {
		return GatewayModelPricingPolicy{}, false, nil
	}
	if !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, false, errGatewayModelPricingStoreUnavailable
	}
	return policy, true, nil
}

func (repository *memoryGatewayModelPricingRepository) PutRevision(
	pricingContext GatewayModelPricingContext,
	input GatewayModelPricingPolicyInput,
	now time.Time,
) (GatewayModelPricingPolicy, error) {
	input = normalizeGatewayModelPricingPolicyInput(input)
	if !validGatewayModelPricingContext(pricingContext) || !validGatewayModelPricingPolicyInput(input) || now.IsZero() {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := gatewayModelPricingScopeKey(pricingContext)
	currentVersion, found := repository.current[key]
	if !found {
		currentVersion = 0
	}
	if input.ExpectedVersion != currentVersion {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingVersionConflict
	}
	policy := buildGatewayModelPricingPolicy(pricingContext, input, currentVersion+1, now)
	if !validGatewayModelPricingPolicy(pricingContext, policy) {
		return GatewayModelPricingPolicy{}, errGatewayModelPricingContract
	}
	if repository.revisions[key] == nil {
		repository.revisions[key] = make(map[int64]GatewayModelPricingPolicy)
	}
	repository.revisions[key][policy.RecordVersion] = policy
	repository.current[key] = policy.RecordVersion
	return policy, nil
}

func buildGatewayModelPricingPolicy(
	pricingContext GatewayModelPricingContext,
	input GatewayModelPricingPolicyInput,
	recordVersion int64,
	now time.Time,
) GatewayModelPricingPolicy {
	policy := GatewayModelPricingPolicy{
		SchemaVersion: GatewayModelPricingPolicySchemaVersion,
		PolicyID:      gatewayModelPricingPolicyID(pricingContext), RecordVersion: recordVersion,
		TenantRef: pricingContext.TenantRef, WorkspaceID: pricingContext.WorkspaceID,
		Environment: pricingContext.Environment, ProviderID: pricingContext.ProviderID,
		ProfileID: pricingContext.ProfileID, ModelID: pricingContext.ModelID,
		Currency: input.Currency, TokenUnit: GatewayModelPricingTokenUnit,
		InputPriceMicrosPerTokenUnit:  input.InputPriceMicrosPerTokenUnit,
		OutputPriceMicrosPerTokenUnit: input.OutputPriceMicrosPerTokenUnit,
		Reason:                        input.Reason, UpdatedAt: now.UTC(), UpdatedByActorRef: pricingContext.ActorRef,
		RequestID: pricingContext.RequestID, AuditRef: pricingContext.AuditRef,
	}
	policy.PolicyDigest = gatewayModelPricingPolicyDigest(policy)
	return policy
}

type GatewayModelPricingSnapshot struct {
	Availability                  string
	Reason                        string
	Currency                      string
	TokenUnit                     int64
	InputPriceMicrosPerTokenUnit  int64
	OutputPriceMicrosPerTokenUnit int64
	PricingPolicyID               string
	PricingPolicyVersion          int64
	PricingPolicyDigest           string
	integrityDigest               string
}

type GatewayRequestCostEstimate struct {
	SchemaVersion                 string `json:"schema_version"`
	Availability                  string `json:"availability"`
	Reason                        string `json:"reason"`
	Currency                      string `json:"currency,omitempty"`
	EstimatedCostMicros           *int64 `json:"estimated_cost_micros,omitempty"`
	TokenUnit                     *int64 `json:"token_unit,omitempty"`
	InputPriceMicrosPerTokenUnit  *int64 `json:"input_price_micros_per_token_unit,omitempty"`
	OutputPriceMicrosPerTokenUnit *int64 `json:"output_price_micros_per_token_unit,omitempty"`
	PricingPolicyID               string `json:"pricing_policy_id,omitempty"`
	PricingPolicyVersion          *int64 `json:"pricing_policy_version,omitempty"`
	PricingPolicyDigest           string `json:"pricing_policy_digest,omitempty"`
	RoundingMode                  string `json:"rounding_mode,omitempty"`
}

func gatewayModelPricingSnapshotFromPolicy(policy GatewayModelPricingPolicy) GatewayModelPricingSnapshot {
	snapshot := GatewayModelPricingSnapshot{
		Availability: GatewayPricingSnapshotConfigured,
		Currency:     policy.Currency, TokenUnit: policy.TokenUnit,
		InputPriceMicrosPerTokenUnit:  policy.InputPriceMicrosPerTokenUnit,
		OutputPriceMicrosPerTokenUnit: policy.OutputPriceMicrosPerTokenUnit,
		PricingPolicyID:               policy.PolicyID, PricingPolicyVersion: policy.RecordVersion,
		PricingPolicyDigest: policy.PolicyDigest,
	}
	snapshot.integrityDigest = gatewayModelPricingSnapshotIntegrityDigest(snapshot)
	return snapshot
}

func gatewayModelPricingUnavailableSnapshot(availability string, reason string) GatewayModelPricingSnapshot {
	return GatewayModelPricingSnapshot{Availability: availability, Reason: strings.TrimSpace(reason)}
}

func buildGatewayRequestCostEstimate(
	providerAttempted bool,
	usage GatewayRequestUsage,
	snapshot GatewayModelPricingSnapshot,
) GatewayRequestCostEstimate {
	if !providerAttempted {
		return gatewayRequestCostUnavailable(GatewayRequestCostNotApplicable, "provider_not_attempted")
	}
	if usage.Availability != GatewayRequestUsageReported || !validGatewayRequestUsage(usage) {
		return gatewayRequestCostUnavailable(GatewayRequestCostUsageNotReported, "provider_usage_not_reported")
	}
	if snapshot.Availability == GatewayPricingSnapshotNotConfigured {
		return gatewayRequestCostUnavailable(GatewayRequestCostPriceNotConfigured, "pricing_policy_not_configured")
	}
	if !validGatewayModelPricingSnapshot(snapshot) {
		return gatewayRequestCostUnavailable(GatewayRequestCostPriceUnavailable, "pricing_snapshot_unavailable")
	}
	estimatedCost, ok := gatewayRequestEstimatedCostMicros(usage, snapshot)
	if !ok {
		return gatewayRequestCostUnavailable(GatewayRequestCostPriceUnavailable, "pricing_calculation_unavailable")
	}
	return GatewayRequestCostEstimate{
		SchemaVersion: GatewayRequestCostEstimateSchemaVersion,
		Availability:  GatewayRequestCostEstimated, Reason: "", Currency: snapshot.Currency,
		EstimatedCostMicros: gatewayInt64Pointer(estimatedCost), TokenUnit: gatewayInt64Pointer(snapshot.TokenUnit),
		InputPriceMicrosPerTokenUnit:  gatewayInt64Pointer(snapshot.InputPriceMicrosPerTokenUnit),
		OutputPriceMicrosPerTokenUnit: gatewayInt64Pointer(snapshot.OutputPriceMicrosPerTokenUnit),
		PricingPolicyID:               snapshot.PricingPolicyID,
		PricingPolicyVersion:          gatewayInt64Pointer(snapshot.PricingPolicyVersion),
		PricingPolicyDigest:           snapshot.PricingPolicyDigest,
		RoundingMode:                  GatewayRequestCostRoundingHalfUp,
	}
}

func gatewayRequestLegacyCostEstimate() GatewayRequestCostEstimate {
	return gatewayRequestCostUnavailable(GatewayRequestCostLegacyNotCaptured, "legacy_record_without_cost_snapshot")
}

func gatewayRequestCostUnavailable(availability string, reason string) GatewayRequestCostEstimate {
	return GatewayRequestCostEstimate{
		SchemaVersion: GatewayRequestCostEstimateSchemaVersion,
		Availability:  availability,
		Reason:        strings.TrimSpace(reason),
	}
}

func gatewayRequestEstimatedCostMicros(
	usage GatewayRequestUsage,
	snapshot GatewayModelPricingSnapshot,
) (int64, bool) {
	inputCost := new(big.Int).Mul(big.NewInt(int64(usage.InputTokens)), big.NewInt(snapshot.InputPriceMicrosPerTokenUnit))
	outputCost := new(big.Int).Mul(big.NewInt(int64(usage.OutputTokens)), big.NewInt(snapshot.OutputPriceMicrosPerTokenUnit))
	numerator := new(big.Int).Add(inputCost, outputCost)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, big.NewInt(GatewayModelPricingTokenUnit), remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(big.NewInt(GatewayModelPricingTokenUnit)) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Int64(), quotient.IsInt64() && quotient.Sign() >= 0
}

func validGatewayModelPricingContext(pricingContext GatewayModelPricingContext) bool {
	if pricingContext.RequestContext == nil || pricingContext.Environment != "development" && pricingContext.Environment != "test" {
		return false
	}
	for _, value := range []string{
		pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.ProviderID,
		pricingContext.ProfileID, pricingContext.ModelID, pricingContext.ActorRef,
		pricingContext.RequestID, pricingContext.AuditRef,
	} {
		if !gatewayModelPricingIdentifierPattern.MatchString(strings.TrimSpace(value)) {
			return false
		}
	}
	return true
}

func normalizeGatewayModelPricingPolicyInput(input GatewayModelPricingPolicyInput) GatewayModelPricingPolicyInput {
	input.Currency = strings.TrimSpace(input.Currency)
	input.Reason = strings.TrimSpace(input.Reason)
	return input
}

func validGatewayModelPricingPolicyInput(input GatewayModelPricingPolicyInput) bool {
	return input.ExpectedVersion >= 0 && input.Currency == GatewayModelPricingCurrency &&
		input.InputPriceMicrosPerTokenUnit >= 0 && input.OutputPriceMicrosPerTokenUnit >= 0 &&
		validGatewayModelPricingReason(input.Reason)
}

func validGatewayModelPricingReason(reason string) bool {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 512 || strings.ContainsAny(reason, "\r\n\x00") {
		return false
	}
	lower := strings.ToLower(reason)
	for _, forbidden := range []string{
		"authorization:", "api_key", "apikey", "password", "secret", "credential",
		"postgres://", "postgresql://", "mysql://", "mongodb://", "https://", "http://",
	} {
		if strings.Contains(lower, forbidden) {
			return false
		}
	}
	return true
}

func validGatewayModelPricingPolicy(pricingContext GatewayModelPricingContext, policy GatewayModelPricingPolicy) bool {
	if policy.SchemaVersion != GatewayModelPricingPolicySchemaVersion || policy.PolicyID != gatewayModelPricingPolicyID(pricingContext) ||
		policy.RecordVersion < 1 || policy.TenantRef != pricingContext.TenantRef || policy.WorkspaceID != pricingContext.WorkspaceID ||
		policy.Environment != pricingContext.Environment || policy.ProviderID != pricingContext.ProviderID ||
		policy.ProfileID != pricingContext.ProfileID || policy.ModelID != pricingContext.ModelID ||
		policy.Currency != GatewayModelPricingCurrency || policy.TokenUnit != GatewayModelPricingTokenUnit ||
		policy.InputPriceMicrosPerTokenUnit < 0 || policy.OutputPriceMicrosPerTokenUnit < 0 ||
		!validGatewayModelPricingReason(policy.Reason) || policy.UpdatedAt.IsZero() ||
		!gatewayModelPricingIdentifierPattern.MatchString(policy.UpdatedByActorRef) ||
		!gatewayModelPricingIdentifierPattern.MatchString(policy.RequestID) ||
		!gatewayModelPricingIdentifierPattern.MatchString(policy.AuditRef) ||
		!gatewayModelPricingDigestPattern.MatchString(policy.PolicyDigest) {
		return false
	}
	return policy.PolicyDigest == gatewayModelPricingPolicyDigest(policy)
}

func validGatewayModelPricingSnapshot(snapshot GatewayModelPricingSnapshot) bool {
	return snapshot.Availability == GatewayPricingSnapshotConfigured && snapshot.Reason == "" &&
		snapshot.Currency == GatewayModelPricingCurrency && snapshot.TokenUnit == GatewayModelPricingTokenUnit &&
		snapshot.InputPriceMicrosPerTokenUnit >= 0 && snapshot.OutputPriceMicrosPerTokenUnit >= 0 &&
		strings.HasPrefix(snapshot.PricingPolicyID, "gmp_") && snapshot.PricingPolicyVersion >= 1 &&
		gatewayModelPricingDigestPattern.MatchString(snapshot.PricingPolicyDigest) &&
		snapshot.integrityDigest != "" && snapshot.integrityDigest == gatewayModelPricingSnapshotIntegrityDigest(snapshot)
}

func gatewayModelPricingScopeKey(pricingContext GatewayModelPricingContext) string {
	return strings.Join([]string{
		pricingContext.TenantRef, pricingContext.WorkspaceID, pricingContext.Environment,
		pricingContext.ProviderID, pricingContext.ProfileID, pricingContext.ModelID,
	}, "\x1f")
}

func gatewayModelPricingPolicyID(pricingContext GatewayModelPricingContext) string {
	digest := sha256.Sum256([]byte(gatewayModelPricingScopeKey(pricingContext)))
	return "gmp_" + hex.EncodeToString(digest[:12])
}

func gatewayModelPricingPolicyDigest(policy GatewayModelPricingPolicy) string {
	document := struct {
		SchemaVersion                 string `json:"schema_version"`
		PolicyID                      string `json:"policy_id"`
		RecordVersion                 int64  `json:"record_version"`
		TenantRef                     string `json:"tenant_ref"`
		WorkspaceID                   string `json:"workspace_id"`
		Environment                   string `json:"environment"`
		ProviderID                    string `json:"provider_id"`
		ProfileID                     string `json:"profile_id"`
		ModelID                       string `json:"model_id"`
		Currency                      string `json:"currency"`
		TokenUnit                     int64  `json:"token_unit"`
		InputPriceMicrosPerTokenUnit  int64  `json:"input_price_micros_per_token_unit"`
		OutputPriceMicrosPerTokenUnit int64  `json:"output_price_micros_per_token_unit"`
		Reason                        string `json:"reason"`
	}{
		policy.SchemaVersion, policy.PolicyID, policy.RecordVersion, policy.TenantRef,
		policy.WorkspaceID, policy.Environment, policy.ProviderID, policy.ProfileID,
		policy.ModelID, policy.Currency, policy.TokenUnit, policy.InputPriceMicrosPerTokenUnit,
		policy.OutputPriceMicrosPerTokenUnit, policy.Reason,
	}
	payload, _ := json.Marshal(document)
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func gatewayModelPricingSnapshotIntegrityDigest(snapshot GatewayModelPricingSnapshot) string {
	value := strings.Join([]string{
		snapshot.Availability, snapshot.Currency, strconv.FormatInt(snapshot.TokenUnit, 10),
		strconv.FormatInt(snapshot.InputPriceMicrosPerTokenUnit, 10),
		strconv.FormatInt(snapshot.OutputPriceMicrosPerTokenUnit, 10),
		snapshot.PricingPolicyID, strconv.FormatInt(snapshot.PricingPolicyVersion, 10), snapshot.PricingPolicyDigest,
	}, "\x1f")
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func gatewayInt64Pointer(value int64) *int64 {
	result := value
	return &result
}

var _ GatewayModelPricingRepository = (*memoryGatewayModelPricingRepository)(nil)
