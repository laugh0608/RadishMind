export const identitySecurity = {
  "eyebrow": "Account security · development/test",
  "title": "Sessions & local credential",
  "claim": "Claim invitation",
  "opening": "Opening…",
  "link": "Link Radish identity",
  "signingOut": "Signing out…",
  "signOut": "Sign out",
  "close": "Close account security",
  "currentMethod": "Current method",
  "sessionOwner": "Session reference",
  "recentAuthentication": "Recent authentication",
  "verified": "Verified",
  "required": "Required",
  "snapshot": "Directory snapshot",
  "pending": "Pending",
  "unknown": "Unknown",
  "committed": "Security change confirmed",
  "committing": "Committing security change…",
  "owner": "Your account sessions",
  "directory": "Session directory",
  "loadedMore": "{{countText}} loaded · more available",
  "loadedComplete": "{{countText}} loaded · complete snapshot",
  "loadingTitle": "Reading sessions",
  "loadingDescription": "Waiting for the current account’s session directory.",
  "emptyTitle": "No sessions returned",
  "emptyDescription": "You are signed in, but the service returned no session records. Reload the directory before a security action.",
  "currentSession": "Current session",
  "outsideWindow": "The current session is outside the loaded page. Load the remaining sessions before making a change.",
  "otherSessions": "Other active sessions",
  "noOtherSessions": "No other active sessions in this snapshot.",
  "endedHistory": "Ended history",
  "noEndedSessions": "No expired or revoked sessions in this snapshot.",
  "loadingMore": "Loading next snapshot page…",
  "loadMore": "Load remaining sessions",
  "invalidTitle": "Invalid session directory",
  "invalidDescription": "The session data does not match the current session. Reload the directory.",
  "actions": "Credential and bulk session actions",
  "credentialAction": "Local password security",
  "rotate": "Rotate local credential",
  "available": "Available",
  "unavailable": "Unavailable",
  "rotationDescription": "Replaces the local password and revokes all active sessions created from that credential in one transaction. OIDC sessions remain active.",
  "currentPassword": "Current password",
  "newPassword": "Replacement password",
  "passwordHelp": "12–1024 characters; the server checks the password policy.",
  "confirmPassword": "Confirm replacement",
  "impactConsent": "I understand that sessions created from this credential will be revoked.",
  "loadBeforeRotation": "Load the complete directory to review which sessions the password change will revoke.",
  "reviewRotation": "Review credential rotation",
  "keepCurrent": "Keep current session",
  "revokeOthers": "Revoke other active sessions",
  "bulkDescription": "Sign out all other active sessions together. This browser session remains active.",
  "loadBeforeBulk": "Load the complete directory before reviewing the target sessions.",
  "reviewOthers": "Review revoke others",
  "privacyBoundary": "Device fingerprints, IP addresses, raw User-Agent data, upstream tokens, passwords and session directory data are not saved in browser storage.",
  "thisSession": "This browser session",
  "lastVerified": "Last verified",
  "reviewRevoke": "Review revoke",
  "confirmRotation": "Rotate credential and revoke its sessions?",
  "confirmExact": "Revoke this exact session?",
  "confirmBulk": "Revoke every other active session?",
  "confirmation": "Explicit security confirmation",
  "currentWillClose": "The current session is included. After success, its authentication cookie is cleared and you must sign in again.",
  "currentWillRemain": "The current session remains active. OIDC sessions outside the listed set also remain active.",
  "targetSet": "Reviewed target sessions · {{countText}}",
  "noTargets": "No active target is present in this snapshot.",
  "atomicity": "If the password change or any session revocation fails, the old credential and all sessions remain unchanged.",
  "cancel": "Cancel and clear input",
  "rotateAndRevoke": "Rotate and revoke",
  "revokeExact": "Revoke exact session",
  "commitBulk": "Revoke other sessions",
  "reload": "Reload session directory",
  "methods": {
    "local_password": "Local password",
    "oidc": "Radish OIDC"
  },
  "states": {
    "active": "Active",
    "expired": "Expired",
    "revoked": "Revoked"
  },
  "inputErrors": {
    "unavailable": "This account has no active local credential to rotate.",
    "recent": "Sign out and authenticate again before rotating the local credential.",
    "incomplete": "Load the complete directory before reviewing the affected sessions.",
    "range": "Enter the current password and a replacement of 12–1024 characters.",
    "mismatch": "The replacement password confirmation does not match.",
    "impact": "Confirm the session impact before continuing."
  },
  "success": {
    "currentRevoked": "Current session revoked. Sign in again to continue.",
    "exactRevoked": "The selected session was revoked.",
    "othersRevoked_one": "{{countText}} other active session revoked.",
    "othersRevoked_other": "{{countText}} other active sessions revoked.",
    "credentialClosed": "Credential rotated under {{policyVersion}}; the current local session is closed.",
    "credentialRevoked_one": "Credential rotated; {{countText}} session created from that credential revoked.",
    "credentialRevoked_other": "Credential rotated; {{countText}} sessions created from that credential revoked."
  },
  "failures": {
    "authentication_required": {
      "title": "Authentication required",
      "message": "The current session no longer authorizes this action. Sign in again."
    },
    "denied": {
      "title": "Session scope denied",
      "message": "The account, session, Origin or CSRF check no longer matches. No change was applied."
    },
    "recent_authentication": {
      "title": "Recent authentication required",
      "message": "Sign out and authenticate again before retrying this security action."
    },
    "conflict": {
      "title": "Session or credential changed",
      "message": "The session or credential changed before commit. Reload before making another decision."
    },
    "credential_unavailable": {
      "title": "Local credential unavailable",
      "message": "This account has no active local credential. No login method was created."
    },
    "credential_invalid": {
      "title": "Password verification rejected",
      "message": "The current password was invalid or the replacement reused it. All prior state remains unchanged."
    },
    "credential_policy": {
      "title": "Password policy rejected",
      "message": "The replacement did not satisfy the server password policy. Re-enter both password fields."
    },
    "unavailable": {
      "title": "Security service unavailable",
      "message": "The service could not complete this action. Reload to check the current state before retrying."
    },
    "invalid_response": {
      "title": "Invalid security response",
      "message": "The response did not match the expected format or privacy boundary and was rejected."
    },
    "failed": {
      "title": "Security request failed",
      "message": "The result could not be verified. Reload to check the current state."
    },
    "selectionStale": {
      "title": "Selected session changed",
      "message": "The target is no longer active in this snapshot. Reload before making another decision."
    },
    "reviewStale": {
      "title": "Credential review expired",
      "message": "The credential input was cleared before commit. Re-enter both password fields."
    }
  }
} as const;
