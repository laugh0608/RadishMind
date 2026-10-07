export const adminShell = {
  "controlPlane": "S7 · Admin Control Plane",
  "quotaHeader": "S9 · Admin Quota Admission",
  "pricingHeader": "S7 · Admin Pricing",
  "invitations": "Workspace invitations",
  "quotaTitle": "Application request quota",
  "pricingTitle": "Provider model pricing",
  "administration": "Administration and routing",
  "invitationHelp": "Issue one-time, time-limited access intent for this exact workspace. Membership and role access begin only after a signed-in claimant completes the atomic claim.",
  "quotaHelp": "Maintain the exact development/test application policy and its quota-owner UTC usage.",
  "pricingHelp": "Maintain one exact immutable USD pricing revision for future development/test request snapshots.",
  "adminHelp": "Review scoped identity evidence, then use the explicit development/test configuration owner.",
  "tenant": "Tenant",
  "workspace": "Workspace",
  "unavailable": "unavailable",
  "authSource": "Auth source",
  "environment": "Environment",
  "resourceNavigation": "Admin Control Plane resources",
  "resourcePath": "Resource path",
  "singleOwner": "One owner at a time",
  "loadingInvitations": "Loading workspace invitations…",
  "loadingIdentity": "Loading local identity administration…",
  "supportingEvidence": "Supporting readiness and deployment evidence",
  "loadingEvidence": "Loading supporting evidence…",
  "boundary": "User and Role consume only the exact workspace member directory and server-owned built-in role catalog. Invitations express one-time access intent. Email delivery, global account search, admin invitations, custom roles, production IAM and bootstrap HTTP stay closed.",
  "tasks": {
    "tenant": {
      "label": "Tenant",
      "scope": "tenant:read"
    },
    "user": {
      "label": "User",
      "scope": "Local identity"
    },
    "role": {
      "label": "Role",
      "scope": "Local grants"
    },
    "invitations": {
      "label": "Invitations",
      "scope": "One-time claim"
    },
    "audit": {
      "label": "Audit",
      "scope": "audit:read"
    },
    "provider": {
      "label": "Provider",
      "scope": "Inventory reference"
    },
    "profile": {
      "label": "Profile",
      "scope": "Assignment"
    },
    "route": {
      "label": "Route",
      "scope": "Generation CAS"
    },
    "quota": {
      "label": "Quota",
      "scope": "UTC daily CAS"
    },
    "pricing": {
      "label": "Pricing",
      "scope": "USD / 1M CAS"
    }
  },
  "status": {
    "authenticated": "Authenticated",
    "offlineEvidence": "Offline evidence",
    "offlineWindow": "Offline window",
    "devControl": "Dev/test control",
    "offline": "Offline",
    "members": "Member directory",
    "roles": "Built-in catalog",
    "claim": "One-time claim",
    "quota": "UTC daily CAS",
    "pricing": "USD / 1M CAS",
    "auditCursor": "{{countText}} · more pages available",
    "auditPage": "{{countText}} · end of page window",
    "failure": "Read failed · {{code}}"
  },
  "auth": {
    "localSession": "Local Web session",
    "offlineFixtures": "Offline fixtures",
    "oidcTest": "OIDC integration test",
    "signedTest": "Signed test token",
    "devHeaders": "Development headers"
  },
  "control": "dev/test control",
  "offlineOwner": "offline",
  "loadingProvider": "Loading controlled configuration owner…",
  "quotaOwner": "Quota · development / test",
  "quotaOwnerTitle": "Application request quota admission",
  "quotaOwnerHelp": "Read the current UTC owner, then review and confirm one expected-version policy update.",
  "loadingQuota": "Loading application quota owner…",
  "pricingOwner": "Pricing · development / test",
  "pricingOwnerTitle": "Immutable model pricing revisions",
  "pricingOwnerHelp": "Read one exact Provider / Profile / Model owner, then review and confirm a CAS revision for future requests.",
  "loadingPricing": "Loading model pricing owner…",
  "loadingRead": "Loading read-only records…",
  "developmentOwner": "{{surface}} · development / test",
  "owners": {
    "provider": {
      "title": "Provider inventory boundary",
      "help": "Reference existing runtime inventory without copying endpoint or credential material."
    },
    "profile": {
      "title": "Provider Profile assignments",
      "help": "Bind stable profile references and capabilities inside one development/test configuration."
    },
    "route": {
      "title": "Versioned model routes",
      "help": "Review immutable candidates before an explicit generation switch changes later Gateway requests."
    }
  }
} as const;
