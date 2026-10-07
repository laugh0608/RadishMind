export const adminQuota = {
  "policiesLabel": "Application quota policies",
  "policies": "Application policies",
  "oneDetail": "One current detail",
  "selectionRequired": "Selection required",
  "applicationList": "Applications in the active workspace",
  "devTest": "development/test",
  "notLoaded": "not loaded",
  "noApplications": "No application owner exists in the active workspace.",
  "selectionNote": "Only the application driving this detail receives the ink-blue selection track. Policy state remains a separate text badge.",
  "selectedApplication": "Selected application",
  "authority": "authority: admin_gateway_quotas:read / write",
  "eyebrow": "Quota · admin_gateway_quotas:read / write",
  "title": "UTC daily provider-attempt policy",
  "usageSource": "Usage comes from the quota owner, never from Request History or the legacy QuotaSummary.",
  "tenant": "Tenant",
  "workspace": "Workspace",
  "environment": "Environment",
  "application": "Application",
  "selectApplication": "Select one application",
  "selectRequired": "selection required",
  "boundary": "Production quota, token/cost limits, billing, delete/disable, automatic increase, automatic routing, formal membership and OIDC remain closed.",
  "admittedAttempts": "admitted provider attempts",
  "remaining": "remaining today",
  "policy": "Policy",
  "period": "Period",
  "recordVersion": "Record version",
  "updatedBy": "Updated by",
  "missingTitle": "No quota policy exists for this exact application scope",
  "policyUpdate": "Policy update",
  "createLimit": "Create UTC daily request limit",
  "changeLimit": "Change UTC daily request limit",
  "limitHelp": "Positive integers only · 1–1,000,000 · no delete or disable",
  "requestLimit": "Request limit",
  "reload": "Reload current policy",
  "reviewUpdate": "Review update",
  "confirmationLabel": "Quota policy update confirmation",
  "confirmationHelp": "The exact tenant, workspace, environment and application scope changes. The current admitted count is never reset or recalculated.",
  "staleHelp": "If the expected version is stale, the write is rejected and this owner must be reloaded.",
  "cancel": "Cancel",
  "updating": "Updating…",
  "confirmUpdate": "Confirm update",
  "casLabel": "Quota policy CAS boundary",
  "casGuard": "CAS guard",
  "reviewHelp": "Review opens an explicit confirmation. No update is sent from this editor directly.",
  "loading": "Loading the exact quota owner…",
  "loadingHelp": "Policy update remains disabled while the owner is unresolved.",
  "unavailableTitle": "Quota response unavailable",
  "noAcceptedPolicy": "No policy or usage was accepted.",
  "retry": "Retry exact owner",
  "version": "version {{version}}",
  "window": "UTC window · {{periodStart}}",
  "usage": "{{admitted}} / {{limit}}",
  "progress": "{{percent}}% of the request limit admitted",
  "expectedVersion": "Expected version {{version}}",
  "confirmationVersion": "Confirmation · expected version {{version}}",
  "createRequests": "Create a limit of {{limit}} requests",
  "updateRequests": "Update {{current}} → {{limit}} requests",
  "missingHelp": "Create a positive UTC daily request limit with expected_version = 0. Missing policy remains fail-closed and is not replaced by Request History or the legacy QuotaSummary.",
  "operations": {
    "invalidRead": "The quota response failed strict validation. No policy or usage was accepted.",
    "invalidLimit": "Request limit must be a positive integer from 1 to 1,000,000.",
    "unchangedLimit": "Enter a different request limit before reviewing an update.",
    "conflict": "The expected version is stale. Reload the quota owner before reviewing another update.",
    "updated": "Policy version {{version}} is now current.",
    "invalidUpdate": "The quota update response could not be verified. Reload the current policy before retrying."
  },
  "states": {
    "ready": "policy ready",
    "missing": "policy missing",
    "loading": "loading",
    "failed": "blocked",
    "limitReached": "limit reached",
    "active": "active",
    "draft": "draft",
    "archived": "archived",
    "unknown": "unknown"
  },
  "failures": {
    "gateway_quota_disabled": {
      "shortLabel": "HTTP closed",
      "title": "Development/test quota management is closed",
      "summary": "The Admin read or write gate is disabled. No fallback policy is available."
    },
    "gateway_quota_scope_denied": {
      "shortLabel": "permission blocked",
      "title": "Quota permission is denied",
      "summary": "The exact admin_gateway_quotas read or write permission is required for this workspace."
    },
    "gateway_quota_environment_forbidden": {
      "shortLabel": "environment blocked",
      "title": "Environment is outside the development/test boundary",
      "summary": "Only development or test is accepted. Production quota remains closed."
    },
    "gateway_quota_payload_invalid": {
      "shortLabel": "payload rejected",
      "title": "Quota update payload was rejected",
      "summary": "The request was rejected. Review the positive integer limit before retrying."
    },
    "gateway_quota_policy_not_found": {
      "shortLabel": "policy missing",
      "title": "Quota policy is missing",
      "summary": "Create the first policy with expected version 0. Missing policy remains fail-closed."
    },
    "gateway_quota_policy_version_conflict": {
      "shortLabel": "version conflict",
      "title": "Quota policy version changed",
      "summary": "The stale write was rejected. Reload the exact owner before reviewing another update."
    },
    "gateway_quota_attempt_conflict": {
      "shortLabel": "attempt conflict",
      "title": "Quota admission attempt conflicts",
      "summary": "The provider attempt did not proceed. Reload the quota owner before another management action."
    },
    "gateway_quota_exceeded": {
      "shortLabel": "limit reached",
      "title": "UTC daily request limit is reached",
      "summary": "New provider attempts are blocked. The status does not change the selected application."
    },
    "gateway_quota_store_unavailable": {
      "shortLabel": "store unavailable",
      "title": "Quota store is unavailable",
      "summary": "The exact quota owner cannot be read. The application is not treated as unlimited."
    }
  }
} as const;
