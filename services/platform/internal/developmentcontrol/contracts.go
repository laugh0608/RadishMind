// Package developmentcontrol defines offline rules for one bounded development
// attempt. It authenticates no bearer, persists nothing and executes no work.
package developmentcontrol

import (
	"errors"
	"regexp"

	"radishmind.local/services/platform/internal/strictjson"
)

const Audience = "development-attempt-control.v1"

type Scope struct {
	AccountID                  string `json:"account_id"`
	WorkspaceID                string `json:"workspace_id"`
	ApplicationID              string `json:"application_id"`
	ApplicationAuthorityDigest string `json:"application_authority_digest"`
	BindingID                  string `json:"binding_id"`
	BindingRevision            int64  `json:"binding_revision"`
	AuthorizationRevision      int64  `json:"authorization_revision"`
	RunID                      string `json:"run_id"`
	TaskID                     string `json:"task_id"`
	AttemptID                  string `json:"attempt_id"`
	PlanDigest                 string `json:"plan_digest"`
	BaselineDigest             string `json:"baseline_digest"`
	ModelRef                   string `json:"model_ref"`
	PricingDigest              string `json:"pricing_digest"`
	ToolPolicyDigest           string `json:"tool_policy_digest"`
}

type Limits struct {
	MaxModelCalls   int64 `json:"max_model_calls"`
	MaxToolCalls    int64 `json:"max_tool_calls"`
	MaxTokens       int64 `json:"max_tokens"`
	MaxCostMicrousd int64 `json:"max_cost_microusd"`
}

type Delegation struct {
	SchemaVersion      string `json:"schema_version"`
	DelegationID       string `json:"delegation_id"`
	Revision           int64  `json:"revision"`
	Audience           string `json:"audience"`
	ExecutorInstanceID string `json:"executor_instance_id"`
	Scope              Scope  `json:"scope"`
	IssuedAtMS         int64  `json:"issued_at_ms"`
	ExpiresAtMS        int64  `json:"expires_at_ms"`
	Limits             Limits `json:"limits"`
}

// Authority is a fresh, trusted owner snapshot, never a field accepted from SDK.
// Token authentication and resource-owner reads must precede calling these rules.
type Authority struct {
	Scope              Scope
	DelegationRevision int64
	ExecutorInstanceID string
	Revoked            bool
}

type Reservation struct {
	SchemaVersion      string `json:"schema_version"`
	DelegationID       string `json:"delegation_id"`
	DelegationRevision int64  `json:"delegation_revision"`
	AttemptID          string `json:"attempt_id"`
	RequestID          string `json:"request_id"`
	Sequence           int64  `json:"sequence"`
	Kind               string `json:"kind"`
	PayloadDigest      string `json:"payload_digest"`
	TokenBound         int64  `json:"token_bound"`
	CostBoundMicrousd  int64  `json:"cost_bound_microusd"`
}

// Receipt is produced by the trusted Gateway / executor, not SDK cost metrics.
type Receipt struct {
	SchemaVersion string `json:"schema_version"`
	AttemptID     string `json:"attempt_id"`
	RequestID     string `json:"request_id"`
	Status        string `json:"status"`
	Tokens        int64  `json:"tokens"`
	CostMicrousd  int64  `json:"cost_microusd"`
}

type Budget struct {
	ModelReservations int64 `json:"model_reservations"`
	ToolReservations  int64 `json:"tool_reservations"`
	HeldTokens        int64 `json:"held_tokens"`
	HeldCostMicrousd  int64 `json:"held_cost_microusd"`
	SpentTokens       int64 `json:"spent_tokens"`
	SpentCostMicrousd int64 `json:"spent_cost_microusd"`
	InFlight          int64 `json:"in_flight"`
	Unresolved        int64 `json:"unresolved"`
}

type Event struct {
	SchemaVersion     string `json:"schema_version"`
	EventID           string `json:"event_id"`
	AttemptID         string `json:"attempt_id"`
	DelegationID      string `json:"delegation_id"`
	Sequence          int64  `json:"sequence"`
	OccurredAtMS      int64  `json:"occurred_at_ms"`
	Type              string `json:"type"`
	RequestID         string `json:"request_id"`
	Phase             string `json:"phase"`
	Reason            string `json:"reason"`
	UsageStatus       string `json:"usage_status"`
	TerminationStatus string `json:"termination_status"`
	Budget            Budget `json:"budget"`
}

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,63}$`)
var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var modelRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)

func (d Delegation) Validate() error {
	if d.SchemaVersion != "development_attempt_delegation.v1" || !idPattern.MatchString(d.DelegationID) || d.Revision < 1 || d.Revision > 1000000000 || d.Audience != Audience || !idPattern.MatchString(d.ExecutorInstanceID) {
		return errors.New("invalid delegation identity or audience")
	}
	s := d.Scope
	for _, id := range []string{s.AccountID, s.WorkspaceID, s.ApplicationID, s.BindingID, s.RunID, s.TaskID, s.AttemptID} {
		if !idPattern.MatchString(id) {
			return errors.New("invalid delegation resource reference")
		}
	}
	if !modelRefPattern.MatchString(s.ModelRef) {
		return errors.New("invalid model catalogue reference")
	}
	for _, digest := range []string{s.ApplicationAuthorityDigest, s.PlanDigest, s.BaselineDigest, s.PricingDigest, s.ToolPolicyDigest} {
		if !digestPattern.MatchString(digest) {
			return errors.New("invalid delegation digest")
		}
	}
	if s.BindingRevision < 1 || s.BindingRevision > 1000000000 || s.AuthorizationRevision < 1 || s.AuthorizationRevision > 1000000000 || d.IssuedAtMS < 1 || d.ExpiresAtMS > 9007199254740991 || d.ExpiresAtMS <= d.IssuedAtMS || d.ExpiresAtMS-d.IssuedAtMS > 600000 {
		return errors.New("invalid delegation revision or lifetime")
	}
	l := d.Limits
	if l.MaxModelCalls < 1 || l.MaxModelCalls > 8 || l.MaxToolCalls < 1 || l.MaxToolCalls > 32 || l.MaxTokens < 1 || l.MaxTokens > 262144 || l.MaxCostMicrousd < 1 || l.MaxCostMicrousd > 1000000000 {
		return errors.New("invalid delegation limits")
	}
	return nil
}

func (r Reservation) Validate() error {
	if r.SchemaVersion != "development_attempt_reservation.v1" || !idPattern.MatchString(r.DelegationID) || r.DelegationRevision < 1 || r.DelegationRevision > 1000000000 || !idPattern.MatchString(r.AttemptID) || !idPattern.MatchString(r.RequestID) || r.Sequence < 1 || r.Sequence > 40 || !digestPattern.MatchString(r.PayloadDigest) {
		return errors.New("invalid reservation identity")
	}
	switch r.Kind {
	case "model":
		if r.TokenBound < 1 || r.TokenBound > 262144 || r.CostBoundMicrousd < 1 || r.CostBoundMicrousd > 1000000000 {
			return errors.New("model reservation requires bounded tokens and price")
		}
	case "tool":
		if r.TokenBound != 0 || r.CostBoundMicrousd != 0 {
			return errors.New("tool reservation cannot claim model usage")
		}
	default:
		return errors.New("unknown reservation kind")
	}
	return nil
}

func (r Receipt) Validate() error {
	if r.SchemaVersion != "development_attempt_receipt.v1" || !idPattern.MatchString(r.AttemptID) || !idPattern.MatchString(r.RequestID) || r.Tokens < 0 || r.Tokens > 1000000000000 || r.CostMicrousd < 0 || r.CostMicrousd > 1000000000000 {
		return errors.New("invalid receipt identity or amount")
	}
	switch r.Status {
	case "reported":
	case "missing", "unknown", "not_dispatched", "not_applicable":
		if r.Tokens != 0 || r.CostMicrousd != 0 {
			return errors.New("unreported usage must not contain invented amounts")
		}
	default:
		return errors.New("unknown receipt status")
	}
	return nil
}

func ParseDelegation(data []byte) (Delegation, error) {
	var d Delegation
	if err := strictjson.Decode(data, &d); err != nil {
		return d, err
	}
	fields, err := strictjson.ObjectFields(data, "schema_version", "delegation_id", "revision", "audience", "executor_instance_id", "scope", "issued_at_ms", "expires_at_ms", "limits")
	if err != nil {
		return d, err
	}
	if _, err = strictjson.ObjectFields(fields["scope"], "account_id", "workspace_id", "application_id", "application_authority_digest", "binding_id", "binding_revision", "authorization_revision", "run_id", "task_id", "attempt_id", "plan_digest", "baseline_digest", "model_ref", "pricing_digest", "tool_policy_digest"); err != nil {
		return d, err
	}
	if _, err = strictjson.ObjectFields(fields["limits"], "max_model_calls", "max_tool_calls", "max_tokens", "max_cost_microusd"); err != nil {
		return d, err
	}
	return d, d.Validate()
}

func ParseReservation(data []byte) (Reservation, error) {
	var r Reservation
	if err := strictjson.Decode(data, &r); err != nil {
		return r, err
	}
	if _, err := strictjson.ObjectFields(data, "schema_version", "delegation_id", "delegation_revision", "attempt_id", "request_id", "sequence", "kind", "payload_digest", "token_bound", "cost_bound_microusd"); err != nil {
		return r, err
	}
	return r, r.Validate()
}

func ParseReceipt(data []byte) (Receipt, error) {
	var r Receipt
	if err := strictjson.Decode(data, &r); err != nil {
		return r, err
	}
	if _, err := strictjson.ObjectFields(data, "schema_version", "attempt_id", "request_id", "status", "tokens", "cost_microusd"); err != nil {
		return r, err
	}
	return r, r.Validate()
}
