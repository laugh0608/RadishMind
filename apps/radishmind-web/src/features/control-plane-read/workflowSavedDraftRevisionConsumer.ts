import type { WorkflowDraftDesignerDraft } from "./workflowDraftDesigner.ts";
import {
  savedWorkflowDraftHeadersForApplication,
  workflowDraftFromSavedWorkflowDraftDocument,
  type SavedWorkflowDraftDocument,
  type WorkflowSavedDraftConsumerConfig,
} from "./savedWorkflowDraftConsumer.ts";

export type WorkflowSavedDraftRevisionKind = "saved" | "restored" | "backfilled_current";

export type WorkflowSavedDraftRevisionSummary = {
  schemaVersion: string;
  draftId: string;
  draftVersion: number;
  revisionKind: WorkflowSavedDraftRevisionKind;
  restoredFromVersion: number;
  draftStatus: string;
  name: string;
  updatedAt: string;
  updatedByActorRef: string;
  nodeCount: number;
  edgeCount: number;
  blockedCount: number;
};

export type WorkflowSavedDraftRevisionDetail = {
  schemaVersion: string;
  revisionKind: WorkflowSavedDraftRevisionKind;
  restoredFromVersion: number;
  draftVersion: number;
  draft: WorkflowDraftDesignerDraft;
};

export type WorkflowSavedDraftRevisionHistoryState = {
  status: "disabled" | "idle" | "loading" | "ready" | "empty" | "failed";
  revisions: WorkflowSavedDraftRevisionSummary[];
  nextCursor: string;
  hasMore: boolean;
  failureCode: string | null;
  requestId: string;
  auditRef: string;
  summary: string;
};

export type WorkflowSavedDraftRevisionRestoreResult = {
  draft: WorkflowDraftDesignerDraft | null;
  currentDraftVersion: number;
  failureCode: string | null;
  requestId: string;
  auditRef: string;
  summary: string;
};

type SavedWorkflowDraftRevisionSummaryDocument = {
  schema_version: string;
  draft_id: string;
  draft_version: number;
  revision_kind: WorkflowSavedDraftRevisionKind;
  restored_from_version: number;
  draft_status: string;
  name: string;
  updated_at: string;
  updated_by_actor_ref: string;
  node_count: number;
  edge_count: number;
  blocked_count: number;
};

type SavedWorkflowDraftRevisionDocument = {
  schema_version: string;
  draft: SavedWorkflowDraftDocument | null;
  revision_kind: WorkflowSavedDraftRevisionKind;
  restored_from_version: number;
};

type SavedWorkflowDraftRevisionListEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  revisions: SavedWorkflowDraftRevisionSummaryDocument[];
  next_cursor: string;
  has_more: boolean;
  failure_code: string | null;
  audit_ref: string;
};

type SavedWorkflowDraftRevisionEnvelope = {
  request_id: string;
  workspace_id: string;
  application_id: string;
  revision: SavedWorkflowDraftRevisionDocument | null;
  draft: SavedWorkflowDraftDocument | null;
  failure_code: string | null;
  current_draft_version: number;
  audit_ref: string;
};

export function initialWorkflowSavedDraftRevisionHistoryState(
  config: WorkflowSavedDraftConsumerConfig,
): WorkflowSavedDraftRevisionHistoryState {
  if (config.mode !== "dev_saved_draft_http") {
    return {
      status: "disabled",
      revisions: [],
      nextCursor: "",
      hasMore: false,
      failureCode: null,
      requestId: "workflow-saved-draft-revisions-disabled",
      auditRef: "audit_workflow_saved_draft_revisions_disabled",
      summary: "修订历史仅在开发 / 测试态已保存草案服务启用时可用。",
    };
  }
  return {
    status: "idle",
    revisions: [],
    nextCursor: "",
    hasMore: false,
    failureCode: null,
    requestId: "workflow-saved-draft-revisions-idle",
    auditRef: "audit_workflow_saved_draft_revisions_idle",
    summary: "可读取当前草案的不可变修订历史。",
  };
}

export async function listWorkflowSavedDraftRevisions(
  draft: WorkflowDraftDesignerDraft,
  config: WorkflowSavedDraftConsumerConfig,
  cursor = "",
  limit = 20,
): Promise<WorkflowSavedDraftRevisionHistoryState> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
    limit: String(limit),
  });
  if (cursor) query.set("cursor", cursor);
  const envelope = await requestRevisionJSON<SavedWorkflowDraftRevisionListEnvelope>(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}/revisions?${query.toString()}`,
    config,
    draft.applicationRef,
    `dev-saved-draft-revisions-${draft.draftId}`,
    { method: "GET" },
    isRevisionListEnvelope,
  );
  const revisions = envelope.revisions.map(toRevisionSummary);
  if (envelope.failure_code) {
    return {
      status: "failed",
      revisions: [],
      nextCursor: "",
      hasMore: false,
      failureCode: envelope.failure_code,
      requestId: envelope.request_id,
      auditRef: envelope.audit_ref,
      summary: `修订历史读取失败：${envelope.failure_code}。`,
    };
  }
  return {
    status: revisions.length === 0 ? "empty" : "ready",
    revisions,
    nextCursor: envelope.next_cursor,
    hasMore: envelope.has_more,
    failureCode: null,
    requestId: envelope.request_id,
    auditRef: envelope.audit_ref,
    summary: revisions.length === 0
      ? "当前草案尚无可读取的修订记录。"
      : `已读取 ${revisions.length} 条不可变修订记录。`,
  };
}

export async function readWorkflowSavedDraftRevision(
  draft: WorkflowDraftDesignerDraft,
  draftVersion: number,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftRevisionDetail> {
  const query = new URLSearchParams({
    workspace_id: config.workspaceId,
    application_id: draft.applicationRef,
  });
  const envelope = await requestRevisionJSON<SavedWorkflowDraftRevisionEnvelope>(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}/revisions/${draftVersion}?${query.toString()}`,
    config,
    draft.applicationRef,
    `dev-saved-draft-revision-${draft.draftId}-${draftVersion}`,
    { method: "GET" },
    isRevisionEnvelope,
  );
  if (envelope.failure_code || !envelope.revision?.draft) {
    throw new Error(`草案修订读取失败：${envelope.failure_code ?? "draft_revision_missing"}。`);
  }
  return {
    schemaVersion: envelope.revision.schema_version,
    revisionKind: envelope.revision.revision_kind,
    restoredFromVersion: envelope.revision.restored_from_version,
    draftVersion: envelope.revision.draft.draft_version,
    draft: workflowDraftFromSavedWorkflowDraftDocument(envelope.revision.draft),
  };
}

export async function restoreWorkflowSavedDraftRevision(
  draft: WorkflowDraftDesignerDraft,
  sourceDraftVersion: number,
  expectedCurrentDraftVersion: number,
  config: WorkflowSavedDraftConsumerConfig,
): Promise<WorkflowSavedDraftRevisionRestoreResult> {
  const envelope = await requestRevisionJSON<SavedWorkflowDraftRevisionEnvelope>(
    `/v1/user-workspace/workflow-drafts/${encodeURIComponent(draft.draftId)}/revisions/${sourceDraftVersion}/restore`,
    config,
    draft.applicationRef,
    `dev-saved-draft-revision-restore-${draft.draftId}-${sourceDraftVersion}`,
    {
      method: "POST",
      body: JSON.stringify({
        expected_current_draft_version: expectedCurrentDraftVersion,
      }),
    },
    isRevisionEnvelope,
  );
  if (envelope.failure_code || !envelope.draft) {
    return {
      draft: null,
      currentDraftVersion: envelope.current_draft_version,
      failureCode: envelope.failure_code ?? "draft_revision_restore_failed",
      requestId: envelope.request_id,
      auditRef: envelope.audit_ref,
      summary: `恢复未执行：${envelope.failure_code ?? "draft_revision_restore_failed"}。`,
    };
  }
  return {
    draft: workflowDraftFromSavedWorkflowDraftDocument(envelope.draft),
    currentDraftVersion: envelope.current_draft_version,
    failureCode: null,
    requestId: envelope.request_id,
    auditRef: envelope.audit_ref,
    summary: `已从版本 ${sourceDraftVersion} 创建新版本 ${envelope.current_draft_version}；历史版本保持不变。`,
  };
}

async function requestRevisionJSON<T>(
  path: string,
  config: WorkflowSavedDraftConsumerConfig,
  applicationRef: string,
  requestId: string,
  init: RequestInit,
  guard: (value: unknown) => value is T,
): Promise<T> {
  if (config.mode !== "dev_saved_draft_http") {
    throw new Error("草案修订历史服务未启用。");
  }
  const headers = new Headers(savedWorkflowDraftHeadersForApplication(
    config,
    applicationRef,
    requestId,
    init.method === "POST",
  ));
  if (init.method === "POST") {
    headers.set(
      "X-RadishMind-Dev-Read-Membership-Permissions",
      "workflow_drafts:read,workflow_drafts:write",
    );
  }
  const response = await fetch(`${config.baseUrl}${path}`, {
    ...init,
    headers,
  });
  const body: unknown = await response.json();
  if (!response.ok) {
    throw new Error(`草案修订接口返回 HTTP ${response.status}。`);
  }
  if (!guard(body)) {
    throw new Error("草案修订接口返回了无法识别的响应结构。");
  }
  return body;
}

function toRevisionSummary(
  document: SavedWorkflowDraftRevisionSummaryDocument,
): WorkflowSavedDraftRevisionSummary {
  return {
    schemaVersion: document.schema_version,
    draftId: document.draft_id,
    draftVersion: document.draft_version,
    revisionKind: document.revision_kind,
    restoredFromVersion: document.restored_from_version,
    draftStatus: document.draft_status,
    name: document.name,
    updatedAt: document.updated_at,
    updatedByActorRef: document.updated_by_actor_ref,
    nodeCount: document.node_count,
    edgeCount: document.edge_count,
    blockedCount: document.blocked_count,
  };
}

function isRevisionListEnvelope(value: unknown): value is SavedWorkflowDraftRevisionListEnvelope {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<SavedWorkflowDraftRevisionListEnvelope>;
  return typeof candidate.request_id === "string" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    Array.isArray(candidate.revisions) &&
    typeof candidate.next_cursor === "string" &&
    typeof candidate.has_more === "boolean" &&
    (typeof candidate.failure_code === "string" || candidate.failure_code === null) &&
    typeof candidate.audit_ref === "string";
}

function isRevisionEnvelope(value: unknown): value is SavedWorkflowDraftRevisionEnvelope {
  if (!value || typeof value !== "object") return false;
  const candidate = value as Partial<SavedWorkflowDraftRevisionEnvelope>;
  return typeof candidate.request_id === "string" &&
    typeof candidate.workspace_id === "string" &&
    typeof candidate.application_id === "string" &&
    typeof candidate.current_draft_version === "number" &&
    (typeof candidate.failure_code === "string" || candidate.failure_code === null) &&
    typeof candidate.audit_ref === "string";
}
