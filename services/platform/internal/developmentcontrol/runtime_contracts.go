package developmentcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"slices"
	"time"

	"radishmind.local/services/platform/internal/strictjson"
)

var (
	ErrNotFound    = errors.New("development_attempt_not_found")
	ErrConflict    = errors.New("development_attempt_version_conflict")
	ErrUnavailable = errors.New("development_store_unavailable")
	ErrIneligible  = errors.New("development_qualification_denied")
	ErrCredential  = errors.New("development_credential_denied")
	ErrTransport   = errors.New("development_native_transport_unavailable")
	ErrContext     = errors.New("development_context_mismatch")
)

type Identity struct {
	TenantRef   string `json:"tenant_ref"`
	AccountID   string `json:"account_id"`
	WorkspaceID string `json:"workspace_id"`
}

type ToolPolicy struct {
	ReadPaths           []string `json:"read_paths"`
	WritePaths          []string `json:"write_paths"`
	VerificationProfile string   `json:"verification_profile"`
}

func (p ToolPolicy) Validate() error {
	if len(p.ReadPaths) < 1 || len(p.ReadPaths) > 64 || p.WritePaths == nil || len(p.WritePaths) > 64 || p.VerificationProfile != "go_test_v1" {
		return ErrIneligible
	}
	seen := map[string]bool{}
	for _, name := range p.ReadPaths {
		if seen[name] || (ToolAction{SchemaVersion: "development_tool_action.v1", Action: "read_file", Path: name}).Validate() != nil {
			return ErrIneligible
		}
		seen[name] = true
	}
	seen = map[string]bool{}
	for _, name := range p.WritePaths {
		if seen[name] || !slices.Contains(p.ReadPaths, name) {
			return ErrIneligible
		}
		seen[name] = true
	}
	return nil
}
func (p ToolPolicy) Allows(a ToolAction) bool {
	if a.Validate() != nil {
		return false
	}
	switch a.Action {
	case "read_file":
		return slices.Contains(p.ReadPaths, a.Path)
	case "replace_file":
		return slices.Contains(p.WritePaths, a.Path)
	case "run_tests":
		return p.VerificationProfile == "go_test_v1"
	default:
		return false
	}
}

type Plan struct {
	SchemaVersion         string     `json:"schema_version"`
	ApplicationID         string     `json:"application_id"`
	ExecutionProfile      string     `json:"execution_profile"`
	DefinitionID          string     `json:"definition_id"`
	ModelRef              string     `json:"model_ref"`
	BaselineDigest        string     `json:"baseline_digest"`
	InitialMessagesDigest string     `json:"initial_messages_digest"`
	Tools                 ToolPolicy `json:"tools"`
	Limits                Limits     `json:"limits"`
	LifetimeMS            int64      `json:"lifetime_ms"`
	MaxOutputTokens       int64      `json:"max_output_tokens"`
}

func ParsePlan(data []byte) (Plan, error) {
	var plan Plan
	if err := strictjson.Decode(data, &plan); err != nil {
		return plan, err
	}
	fields, err := strictjson.ObjectFields(data, "schema_version", "application_id", "execution_profile", "definition_id", "model_ref", "baseline_digest", "initial_messages_digest", "tools", "limits", "lifetime_ms", "max_output_tokens")
	if err != nil {
		return plan, err
	}
	if _, err = strictjson.ObjectFields(fields["tools"], "read_paths", "write_paths", "verification_profile"); err != nil {
		return plan, err
	}
	if _, err = strictjson.ObjectFields(fields["limits"], "max_model_calls", "max_tool_calls", "max_tokens", "max_cost_microusd"); err != nil {
		return plan, err
	}
	return plan, plan.Validate()
}

func (p Plan) Validate() error {
	if !digestPattern.MatchString(p.InitialMessagesDigest) {
		return ErrIneligible
	}
	if p.SchemaVersion != "development_attempt_plan.v1" || !idPattern.MatchString(p.ApplicationID) || !idPattern.MatchString(p.ExecutionProfile) || p.DefinitionID != "" && !idPattern.MatchString(p.DefinitionID) || !modelRefPattern.MatchString(p.ModelRef) || !digestPattern.MatchString(p.BaselineDigest) || p.LifetimeMS < 1000 || p.LifetimeMS > 600000 || p.MaxOutputTokens < 1 || p.MaxOutputTokens > 262144 {
		return ErrIneligible
	}
	if err := p.Tools.Validate(); err != nil {
		return err
	}
	l := p.Limits
	if l.MaxModelCalls < 1 || l.MaxModelCalls > 8 || l.MaxToolCalls < 1 || l.MaxToolCalls > 32 || l.MaxTokens < 1 || l.MaxTokens > 262144 || l.MaxCostMicrousd < 1 || l.MaxCostMicrousd > 1000000000 {
		return ErrIneligible
	}
	return nil
}
func (p Plan) Digest() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return PayloadDigest(data), nil
}

// Selection is read from current identity, application, model and pricing owners.
// It contains no provider credentials or endpoint, and is never SDK-supplied.
type Selection struct {
	ApplicationAuthorityDigest string `json:"application_authority_digest"`
	IdentityDigest             string `json:"identity_digest"`
	ModelDigest                string `json:"model_digest"`
	WireModel                  string `json:"wire_model"`
	PricingDigest              string `json:"pricing_digest"`
	TokenUnit                  int64  `json:"token_unit"`
	InputPriceMicros           int64  `json:"input_price_micros"`
	OutputPriceMicros          int64  `json:"output_price_micros"`
}

func (s Selection) Validate() error {
	for _, digest := range []string{s.ApplicationAuthorityDigest, s.IdentityDigest, s.ModelDigest, s.PricingDigest} {
		if !digestPattern.MatchString(digest) {
			return ErrIneligible
		}
	}
	if s.WireModel == "" || len(s.WireModel) > 256 || s.TokenUnit != 1000000 || s.InputPriceMicros < 0 || s.OutputPriceMicros < 0 || s.InputPriceMicros > 1000000000000 || s.OutputPriceMicros > 1000000000000 || s.InputPriceMicros+s.OutputPriceMicros == 0 {
		return ErrIneligible
	}
	return nil
}
func (s Selection) Cost(input, output int64) (int64, error) {
	return s.calculateCost(input, output, false)
}
func (s Selection) ReserveCost(input, output int64) (int64, error) {
	return s.calculateCost(input, output, true)
}
func (s Selection) calculateCost(input, output int64, upper bool) (int64, error) {
	if s.Validate() != nil || input < 0 || output < 0 || input > 1000000000000 || output > 1000000000000 {
		return 0, ErrIneligible
	}
	total := new(big.Int).Mul(big.NewInt(input), big.NewInt(s.InputPriceMicros))
	total.Add(total, new(big.Int).Mul(big.NewInt(output), big.NewInt(s.OutputPriceMicros)))
	if upper {
		total.Add(total, big.NewInt(s.TokenUnit-1))
	} else {
		total.Add(total, big.NewInt(s.TokenUnit/2))
	}
	total.Div(total, big.NewInt(s.TokenUnit))
	if !total.IsInt64() {
		return 0, ErrIneligible
	}
	return total.Int64(), nil
}

type QualificationReader interface {
	Resolve(context.Context, Identity, Plan) (Selection, error)
}
type NativePolicy struct {
	InputTokenCeiling int64
	MaxOutputTokens   int64
	Timeout           time.Duration
}
type NativeBackend interface {
	Policy(context.Context, string) (NativePolicy, error)
	Complete(context.Context, string, NativeRequest) (NativeResult, error)
}

type Binding struct {
	SchemaVersion     string    `json:"schema_version"`
	Identity          Identity  `json:"identity"`
	Plan              Plan      `json:"plan"`
	Selection         Selection `json:"selection"`
	InputTokenCeiling int64     `json:"input_token_ceiling"`
}

func (b Binding) Validate() error {
	if b.SchemaVersion != "development_attempt_binding.v1" || b.Identity.TenantRef == "" || len(b.Identity.TenantRef) > 128 || !idPattern.MatchString(b.Identity.AccountID) || !idPattern.MatchString(b.Identity.WorkspaceID) || b.Plan.Validate() != nil || b.Selection.Validate() != nil || b.InputTokenCeiling < 1 || b.InputTokenCeiling+b.Plan.MaxOutputTokens > b.Plan.Limits.MaxTokens {
		return ErrIneligible
	}
	return nil
}

type Offer struct {
	RequestID         string `json:"request_id"`
	ParentRequestID   string `json:"parent_request_id"`
	ToolCallID        string `json:"tool_call_id"`
	PayloadDigest     string `json:"payload_digest"`
	ArgumentsDigest   string `json:"arguments_digest"`
	State             string `json:"state"`
	ObservationDigest string `json:"observation_digest"`
}

type Record struct {
	Version              int64
	OwnerInstanceID      string
	Revoked              bool
	Binding              Binding
	Control              *AttemptControl
	ContextDigest        string
	ContextCount         int64
	PendingToolRequestID string
	Offers               map[string]Offer
}
