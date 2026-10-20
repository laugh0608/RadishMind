export const operationsInbox = {
  "heading": "Operations inbox",
  "items": "Workspace operations attention items",
  "openSelected": "Open selected",
  "viewWorkspace": "View workspace",
  "readOnlyEvidence": "Read-only evidence. Review continues in the owning surface; no action is executed here.",
  "evidencePath": "Evidence path",
  "projection": "Source projection",
  "severity": "Attention severity",
  "authority": "Authority",
  "readOnly": "Read only",
  "mutation": "Mutation",
  "remediation": "Remediation",
  "businessTruth": "Business truth",
  "openEvidence": "Open evidence",
  "select": "Select an attention item to inspect its evidence path.",
  "noAttention": "No attention items",
  "noResources": "No resource items available",
  "complete": "The current source windows are complete.",
  "checkCoverage": "Review source coverage before interpreting the empty queue.",
  "unavailable": "Unavailable",
  "enabled": "Enabled",
  "locked": "Locked",
  "reviewOnly": "Review only",
  "writable": "Writable",
  "sourceCount": "Loaded from {{count}} authorized source projections",
  "sources": {
    "applications": "Applications",
    "api_keys": "API keys",
    "workflow_definitions": "Workflows",
    "runs": "Runs"
  },
  "reasons": {
    "application_archived": {
      "title": "{{name}} is archived",
      "summary": "Archived applications remain available for read-only history review."
    },
    "application_run_attention": {
      "title": "{{name}} reports {{status}}",
      "summary": "Open the application and inspect its authoritative run history."
    },
    "api_key_rotation_required": {
      "title": "API key rotation is required",
      "summary": "Review the existing key lifecycle before issuing or revoking credentials."
    },
    "api_key_expired": {
      "title": "API key is expired",
      "summary": "Review the sanitized key record; no credential material is available here."
    },
    "api_key_expiring": {
      "title": "API key expires within 14 days",
      "summary": "Review the existing key lifecycle before the explicit expiry time."
    },
    "workflow_definition_review": {
      "title": "{{name}} is {{status}}",
      "summary": "Review the immutable definition state; this inbox cannot approve or activate it."
    },
    "run_failure": {
      "title": "{{name}} requires failure review",
      "summary": "Inspect failure {{code}}; replay, resume and automatic remediation remain disabled."
    },
    "run_outcome_unknown": {
      "title": "{{name}} outcome is unknown",
      "summary": "Inspect the run metadata before deciding whether another explicit invocation is appropriate."
    }
  }
} as const;
