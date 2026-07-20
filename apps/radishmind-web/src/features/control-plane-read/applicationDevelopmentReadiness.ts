import type {
  ApplicationDevelopmentStageId,
  ApplicationDevelopmentWorkspaceContext,
} from "./applicationDevelopmentWorkspace.ts";

export const APPLICATION_DEVELOPMENT_SOURCE_GROUP_IDS = [
  "application",
  "configuration_candidate",
  "workflow_authority",
  "rag_authority",
  "controlled_test",
  "evaluation",
  "operations",
] as const;

export const APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS = [
  "application_lifecycle",
  "configuration_draft",
  "publish_candidate",
  "workflow_definition",
  "rag_binding",
  "rag_assignment",
  "controlled_run",
  "evaluation_review",
  "operations_coverage",
] as const;

export type ApplicationDevelopmentSourceGroupId =
  typeof APPLICATION_DEVELOPMENT_SOURCE_GROUP_IDS[number];
export type ApplicationDevelopmentContributionId =
  typeof APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS[number];
export type ApplicationDevelopmentEvidenceStatus =
  | "not_started"
  | "incomplete"
  | "available"
  | "blocked"
  | "partial_failure";
export type ApplicationDevelopmentEvidenceCoverage = "none" | "partial" | "complete";
export type ApplicationDevelopmentReadinessStatus =
  | "review_not_started"
  | "review_incomplete"
  | "review_blocked"
  | "dev_test_evidence_reviewable";

export type ApplicationDevelopmentEvidenceRef = {
  kind:
    | "application"
    | "draft"
    | "candidate"
    | "definition"
    | "binding"
    | "assignment"
    | "session"
    | "run"
    | "request"
    | "evaluation";
  id: string;
  version?: number;
};

export type ApplicationDevelopmentEvidenceBlocker = {
  code: string;
  summary: string;
};

export type ApplicationDevelopmentEvidenceContribution = {
  contributionId: ApplicationDevelopmentContributionId;
  sourceGroupId: ApplicationDevelopmentSourceGroupId;
  status: ApplicationDevelopmentEvidenceStatus;
  coverage: ApplicationDevelopmentEvidenceCoverage;
  evidenceRefs: ApplicationDevelopmentEvidenceRef[];
  missingEvidence: string[];
  blockers: ApplicationDevelopmentEvidenceBlocker[];
  failureCodes: string[];
  nextStage: ApplicationDevelopmentStageId;
  nextAnchor: string;
};

export type ApplicationDevelopmentEvidenceInput = Omit<
  ApplicationDevelopmentEvidenceContribution,
  "sourceGroupId" | "nextStage" | "nextAnchor"
> & {
  applicationId: string;
  workspaceGenerationKey: string;
};

export type ApplicationDevelopmentOwnerEvidence = Omit<
  ApplicationDevelopmentEvidenceInput,
  "applicationId" | "workspaceGenerationKey"
>;

export type ApplicationDevelopmentEvidenceState = {
  applicationId: string;
  workspaceGenerationKey: string;
  contributions: Record<ApplicationDevelopmentContributionId, ApplicationDevelopmentEvidenceContribution>;
};

export type ApplicationDevelopmentReadinessSource = {
  sourceGroupId: ApplicationDevelopmentSourceGroupId;
  label: string;
  status: ApplicationDevelopmentEvidenceStatus;
  coverage: ApplicationDevelopmentEvidenceCoverage;
  evidenceRefs: ApplicationDevelopmentEvidenceRef[];
  missingEvidence: string[];
  blockers: ApplicationDevelopmentEvidenceBlocker[];
  failureCodes: string[];
  nextStage: ApplicationDevelopmentStageId;
  nextAnchor: string;
};

export type ApplicationDevelopmentReadinessViewModel = {
  status: ApplicationDevelopmentReadinessStatus;
  summary: string;
  sources: ApplicationDevelopmentReadinessSource[];
  evidenceCount: number;
  blockerCount: number;
  missingCount: number;
  canPersistReadiness: false;
  canPublish: false;
};

type ContributionDefinition = {
  sourceGroupId: ApplicationDevelopmentSourceGroupId;
  missingLabel: string;
  nextStage: ApplicationDevelopmentStageId;
  nextAnchor: string;
};

const REF_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]{1,159}$/u;
const CODE_PATTERN = /^[a-z][a-z0-9_]{2,95}$/u;

const CONTRIBUTIONS: Record<ApplicationDevelopmentContributionId, ContributionDefinition> = {
  application_lifecycle: {
    sourceGroupId: "application",
    missingLabel: "Select an authoritative Application.",
    nextStage: "configure_build",
    nextAnchor: "workspace-applications",
  },
  configuration_draft: {
    sourceGroupId: "configuration_candidate",
    missingLabel: "Save and validate an Application configuration draft.",
    nextStage: "configure_build",
    nextAnchor: "application-configuration-draft",
  },
  publish_candidate: {
    sourceGroupId: "configuration_candidate",
    missingLabel: "Review an immutable Application publish candidate.",
    nextStage: "human_promotion",
    nextAnchor: "application-publish-review",
  },
  workflow_definition: {
    sourceGroupId: "workflow_authority",
    missingLabel: "Activate an approved immutable Workflow Definition.",
    nextStage: "human_promotion",
    nextAnchor: "workflow-definition-promotion",
  },
  rag_binding: {
    sourceGroupId: "rag_authority",
    missingLabel: "Approve an exact RAG knowledge binding.",
    nextStage: "human_promotion",
    nextAnchor: "workflow-rag-promotion-review",
  },
  rag_assignment: {
    sourceGroupId: "rag_authority",
    missingLabel: "Select the current Application RAG assignment.",
    nextStage: "controlled_test",
    nextAnchor: "application-rag-invocation",
  },
  controlled_run: {
    sourceGroupId: "controlled_test",
    missingLabel: "Record a reviewable v4 or v5 controlled run.",
    nextStage: "controlled_test",
    nextAnchor: "application-interaction-session",
  },
  evaluation_review: {
    sourceGroupId: "evaluation",
    missingLabel: "Review compatible evaluation evidence.",
    nextStage: "evidence_review",
    nextAnchor: "workflow-rag-evaluation-panel",
  },
  operations_coverage: {
    sourceGroupId: "operations",
    missingLabel: "Load Gateway and Workflow operations coverage.",
    nextStage: "evidence_review",
    nextAnchor: "application-operations",
  },
};

const SOURCE_LABELS: Record<ApplicationDevelopmentSourceGroupId, string> = {
  application: "Application",
  configuration_candidate: "Configuration / Candidate",
  workflow_authority: "Workflow authority",
  rag_authority: "RAG authority",
  controlled_test: "Controlled test",
  evaluation: "Evaluation",
  operations: "Operations",
};

export function initialApplicationDevelopmentEvidenceState(
  context: ApplicationDevelopmentWorkspaceContext,
): ApplicationDevelopmentEvidenceState {
  const contributions = Object.fromEntries(
    APPLICATION_DEVELOPMENT_CONTRIBUTION_IDS.map((contributionId) => [
      contributionId,
      emptyContribution(contributionId),
    ]),
  ) as Record<ApplicationDevelopmentContributionId, ApplicationDevelopmentEvidenceContribution>;

  if (context.applicationId) {
    contributions.application_lifecycle = context.status === "active"
      ? contribution("application_lifecycle", "available", "complete", [
        { kind: "application", id: context.applicationId, version: context.recordVersion },
      ])
      : contribution(
        "application_lifecycle",
        "blocked",
        "complete",
        [{ kind: "application", id: context.applicationId, version: context.recordVersion }],
        [],
        [{ code: "application_archived", summary: "Archived Applications retain evidence but cannot enter controlled testing." }],
      );
  }

  return {
    applicationId: context.applicationId,
    workspaceGenerationKey: context.generationKey,
    contributions,
  };
}

export function applyApplicationDevelopmentEvidence(
  state: ApplicationDevelopmentEvidenceState,
  context: ApplicationDevelopmentWorkspaceContext,
  input: ApplicationDevelopmentEvidenceInput,
): ApplicationDevelopmentEvidenceState {
  if (state.workspaceGenerationKey !== context.generationKey || input.workspaceGenerationKey !== context.generationKey) {
    throw new Error("Application development evidence generation is stale.");
  }
  if (!context.applicationId || input.applicationId.trim() !== context.applicationId) {
    throw new Error("Application development evidence scope does not match the current workspace.");
  }
  const normalized = normalizeContribution(input);
  return {
    ...state,
    contributions: { ...state.contributions, [input.contributionId]: normalized },
  };
}

export function buildApplicationDevelopmentReadinessViewModel(
  state: ApplicationDevelopmentEvidenceState,
): ApplicationDevelopmentReadinessViewModel {
  const sources = APPLICATION_DEVELOPMENT_SOURCE_GROUP_IDS.map((sourceGroupId) =>
    rollupSource(sourceGroupId, state.contributions),
  );
  const evidenceCount = sources.reduce((total, source) => total + source.evidenceRefs.length, 0);
  const blockerCount = sources.reduce((total, source) => total + source.blockers.length + source.failureCodes.length, 0);
  const missingCount = sources.reduce((total, source) => total + source.missingEvidence.length, 0);
  const status: ApplicationDevelopmentReadinessStatus = !state.applicationId
    ? "review_not_started"
    : sources.some((source) => source.status === "blocked" || source.status === "partial_failure")
      ? "review_blocked"
      : sources.every((source) => source.status === "available" && source.coverage === "complete")
        ? "dev_test_evidence_reviewable"
        : "review_incomplete";
  return {
    status,
    summary: readinessSummary(status, blockerCount, missingCount),
    sources,
    evidenceCount,
    blockerCount,
    missingCount,
    canPersistReadiness: false,
    canPublish: false,
  };
}

function emptyContribution(contributionId: ApplicationDevelopmentContributionId): ApplicationDevelopmentEvidenceContribution {
  const definition = CONTRIBUTIONS[contributionId];
  return {
    contributionId,
    sourceGroupId: definition.sourceGroupId,
    status: "not_started",
    coverage: "none",
    evidenceRefs: [],
    missingEvidence: [definition.missingLabel],
    blockers: [],
    failureCodes: [],
    nextStage: definition.nextStage,
    nextAnchor: definition.nextAnchor,
  };
}

function contribution(
  contributionId: ApplicationDevelopmentContributionId,
  status: ApplicationDevelopmentEvidenceStatus,
  coverage: ApplicationDevelopmentEvidenceCoverage,
  evidenceRefs: ApplicationDevelopmentEvidenceRef[] = [],
  missingEvidence: string[] = [],
  blockers: ApplicationDevelopmentEvidenceBlocker[] = [],
  failureCodes: string[] = [],
): ApplicationDevelopmentEvidenceContribution {
  return normalizeContribution({
    contributionId,
    status,
    coverage,
    evidenceRefs,
    missingEvidence,
    blockers,
    failureCodes,
  });
}

function normalizeContribution(
  input: ApplicationDevelopmentOwnerEvidence,
): ApplicationDevelopmentEvidenceContribution {
  const definition = CONTRIBUTIONS[input.contributionId];
  if (!definition) throw new Error("Application development evidence contribution is invalid.");
  const evidenceRefs = unique(input.evidenceRefs.map((ref) => {
    const id = ref.id.trim();
    if (!REF_PATTERN.test(id) || ref.version !== undefined && (!Number.isInteger(ref.version) || ref.version < 1)) {
      throw new Error("Application development evidence reference is invalid.");
    }
    return { ...ref, id };
  }), (ref) => `${ref.kind}:${ref.id}:${ref.version ?? 0}`).slice(0, 12);
  const blockers = unique(input.blockers.map((blocker) => {
    const code = blocker.code.trim();
    const summary = safeSummary(blocker.summary);
    if (!CODE_PATTERN.test(code)) throw new Error("Application development evidence blocker code is invalid.");
    return { code, summary };
  }), (blocker) => blocker.code).slice(0, 12);
  const failureCodes = unique(input.failureCodes.map((code) => code.trim()).filter(Boolean), (code) => code);
  if (failureCodes.some((code) => !CODE_PATTERN.test(code))) {
    throw new Error("Application development evidence failure code is invalid.");
  }
  const missingEvidence = unique(input.missingEvidence.map(safeSummary), (item) => item).slice(0, 12);
  return {
    contributionId: input.contributionId,
    sourceGroupId: definition.sourceGroupId,
    status: input.status,
    coverage: input.coverage,
    evidenceRefs,
    missingEvidence,
    blockers,
    failureCodes: failureCodes.slice(0, 12),
    nextStage: definition.nextStage,
    nextAnchor: definition.nextAnchor,
  };
}

function rollupSource(
  sourceGroupId: ApplicationDevelopmentSourceGroupId,
  contributions: ApplicationDevelopmentEvidenceState["contributions"],
): ApplicationDevelopmentReadinessSource {
  const members = Object.values(contributions).filter((item) => item.sourceGroupId === sourceGroupId);
  const status: ApplicationDevelopmentEvidenceStatus = members.some((item) => item.status === "blocked")
    ? "blocked"
    : members.some((item) => item.status === "partial_failure")
      ? "partial_failure"
      : members.every((item) => item.status === "available")
        ? "available"
        : members.some((item) => item.status === "incomplete" || item.status === "available")
          ? "incomplete"
          : "not_started";
  const coverage: ApplicationDevelopmentEvidenceCoverage = members.every((item) => item.coverage === "complete")
    ? "complete"
    : members.some((item) => item.coverage !== "none")
      ? "partial"
      : "none";
  const next = members.find((item) => item.status !== "available") ?? members[0];
  return {
    sourceGroupId,
    label: SOURCE_LABELS[sourceGroupId],
    status,
    coverage,
    evidenceRefs: unique(members.flatMap((item) => item.evidenceRefs), (ref) => `${ref.kind}:${ref.id}:${ref.version ?? 0}`),
    missingEvidence: unique(members.flatMap((item) => item.status === "available" ? [] : item.missingEvidence), (item) => item),
    blockers: unique(members.flatMap((item) => item.blockers), (blocker) => blocker.code),
    failureCodes: unique(members.flatMap((item) => item.failureCodes), (code) => code),
    nextStage: next.nextStage,
    nextAnchor: next.nextAnchor,
  };
}

function readinessSummary(
  status: ApplicationDevelopmentReadinessStatus,
  blockerCount: number,
  missingCount: number,
): string {
  if (status === "review_not_started") return "Select an authoritative Application before reviewing evidence.";
  if (status === "review_blocked") return `${blockerCount} owner blocker or failure signal(s) require review.`;
  if (status === "review_incomplete") return `${missingCount} required development evidence item(s) are still missing.`;
  return "All required development and test evidence groups are available for human review.";
}

function safeSummary(value: string): string {
  const normalized = value.trim().replace(/\s+/gu, " ");
  if (!normalized || normalized.length > 240) {
    throw new Error("Application development evidence summary is invalid.");
  }
  return normalized;
}

function unique<T>(items: T[], key: (item: T) => string): T[] {
  const seen = new Set<string>();
  return items.filter((item) => {
    const value = key(item);
    if (seen.has(value)) return false;
    seen.add(value);
    return true;
  });
}
