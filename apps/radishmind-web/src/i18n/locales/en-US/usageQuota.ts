export const usageQuota = {
  "unavailable": "Unavailable",
  "monthly": "Monthly",
  "workspace": "User Workspace",
  "title": "Usage quota",
  "readOnlyReady": "Read-only ready",
  "blocked": "Blocked",
  "route": "Quota summary route",
  "model": "Read model",
  "period": "Period",
  "request": "Request",
  "audit": "Audit",
  "snapshot": "Workspace usage quota snapshot",
  "quota": "Quota",
  "window": "Read-only quota window",
  "failure": "Failure code",
  "noEnforcement": "Reported only, not enforced by this page",
  "available": "Available",
  "references": "Request and audit references remain visible",
  "noPolicy": "No quota policy summary is available.",
  "limits": "Workspace usage quota limits",
  "limit": "Limit {{limit}} / {{percent}} used",
  "noCurrency": "The source provides no currency; this value is shown without conversion or an inferred currency.",
  "overQuota": "Over-quota failure code",
  "boundary": "Displayed as read-side metadata only; enforcement, rate limits and cost record writes remain outside this page.",
  "states": "Quota state examples",
  "failureCount": "{{count}} items / failure {{code}}",
  "measures": {
    "requests": "Requests",
    "tokens": "Tokens",
    "cost": "Cost value"
  },
  "previews": {
    "ready": {
      "label": "Ready",
      "summary": "Quota summaries use the current read projection without enforcement."
    },
    "empty": {
      "label": "Empty",
      "summary": "Route metadata remains visible when no quota policy summary is returned."
    },
    "denied": {
      "label": "Denied",
      "summary": "Usage scope denial exposes no partial quota or usage rows."
    },
    "stale": {
      "label": "Stale",
      "summary": "Cached quota summaries remain separate from live cost records and Gateway enforcement."
    },
    "partial_failure": {
      "label": "Partial failure",
      "summary": "Failure metadata remains explicit when quota details cannot be fully read."
    },
    "forbidden_projection": {
      "label": "Forbidden projection",
      "summary": "The shared output guard blocks cost writeback payload projections."
    }
  }
} as const;
