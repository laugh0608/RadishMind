import { useEffect, useMemo, useState } from "react";

import {
  archiveWorkflowRAGSnapshot,
  createWorkflowRAGSnapshot,
  listWorkflowRAGSnapshots,
  readWorkflowRAGSnapshot,
  readWorkflowRAGSnapshotConfig,
  validateWorkflowRAGSnapshotWriteInput,
  versionWorkflowRAGSnapshot,
  type WorkflowRAGContentClassification,
  type WorkflowRAGFragmentInput,
  type WorkflowRAGSnapshotLifecycle,
  type WorkflowRAGSnapshotOperationResult,
  type WorkflowRAGSnapshotRecord,
  type WorkflowRAGSnapshotResource,
  type WorkflowRAGSnapshotWriteInput,
} from "./workflowRAGSnapshotConsumer.ts";

const config = readWorkflowRAGSnapshotConfig();
const EMPTY_FRAGMENT: WorkflowRAGFragmentInput = {
  fragmentRef: "overview",
  sourceType: "manual",
  sourceRef: "application_manual",
  pageSlug: "overview",
  title: "应用说明",
  isOfficial: true,
  content: "",
};

type SnapshotEditor = {
  snapshotKey: string;
  displayName: string;
  contentClassification: WorkflowRAGContentClassification;
  fragmentsJSON: string;
};

type SnapshotCollection = {
  active: WorkflowRAGSnapshotResource[];
  archived: WorkflowRAGSnapshotResource[];
  activeCursor: string;
  archivedCursor: string;
  failureCode: string;
  summary: string;
};

export default function WorkflowRAGSnapshotPanel({
  applicationId,
  applicationName,
  applicationActive,
}: {
  applicationId: string;
  applicationName: string;
  applicationActive: boolean;
}) {
  const [collection, setCollection] = useState<SnapshotCollection>(emptyCollection);
  const [filter, setFilter] = useState<WorkflowRAGSnapshotLifecycle>("active");
  const [selectedResource, setSelectedResource] = useState<WorkflowRAGSnapshotResource | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<WorkflowRAGSnapshotRecord | null>(null);
  const [editor, setEditor] = useState<SnapshotEditor>(emptyEditor);
  const [pending, setPending] = useState<"" | "listing" | "reading" | "creating" | "versioning" | "archiving">("");
  const [operation, setOperation] = useState<WorkflowRAGSnapshotOperationResult | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showArchiveConfirm, setShowArchiveConfirm] = useState(false);

  const visibleResources = useMemo(
    () => filter === "active" ? collection.active : collection.archived,
    [collection.active, collection.archived, filter],
  );
  const canRead = config.scopes.has("workflow_rag_snapshots:read");
  const canWrite = applicationActive && config.scopes.has("workflow_rag_snapshots:write");
  const canArchive = applicationActive && config.scopes.has("workflow_rag_snapshots:archive");

  useEffect(() => {
    let cancelled = false;
    setCollection(emptyCollection());
    setSelectedResource(null);
    setSelectedRecord(null);
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    if (config.mode === "offline" || !canRead || !applicationId.trim()) return;
    setPending("listing");
    Promise.all([
      listWorkflowRAGSnapshots(config, applicationId, "active"),
      listWorkflowRAGSnapshots(config, applicationId, "archived"),
    ]).then(([active, archived]) => {
      if (cancelled) return;
      setPending("");
      setCollection(collectionFromResults(active, archived));
    });
    return () => { cancelled = true; };
  }, [applicationId, canRead]);

  const selectResource = async (resource: WorkflowRAGSnapshotResource) => {
    setSelectedResource(resource);
    setSelectedRecord(null);
    setOperation(null);
    setShowCreate(false);
    setShowArchiveConfirm(false);
    setPending("reading");
    const result = await readWorkflowRAGSnapshot(config, applicationId, resource.snapshotId, resource.latestVersion);
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(editorFromRecord(result.record));
  };

  const refreshCollections = async () => {
    const [active, archived] = await Promise.all([
      listWorkflowRAGSnapshots(config, applicationId, "active"),
      listWorkflowRAGSnapshots(config, applicationId, "archived"),
    ]);
    const nextCollection = collectionFromResults(active, archived);
    setCollection(nextCollection);
    return nextCollection;
  };

  const submitCreate = async () => {
    const parsed = parseEditor(editor);
    if (typeof parsed === "string") {
      setOperation(localFailure(parsed));
      return;
    }
    setPending("creating");
    const result = await createWorkflowRAGSnapshot(config, applicationId, parsed);
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(editorFromRecord(result.record));
    setShowCreate(false);
    setFilter("active");
    const refreshed = await refreshCollections();
    setSelectedResource(refreshed.active.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const submitVersion = async () => {
    if (!selectedRecord) return;
    const parsed = parseEditor(editor);
    if (typeof parsed === "string") {
      setOperation(localFailure(parsed));
      return;
    }
    setPending("versioning");
    const result = await versionWorkflowRAGSnapshot(
      config,
      applicationId,
      selectedRecord.snapshotId,
      selectedRecord.snapshotVersion,
      parsed,
    );
    setPending("");
    setOperation(result);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(editorFromRecord(result.record));
    const refreshed = await refreshCollections();
    setSelectedResource(refreshed.active.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const submitArchive = async () => {
    if (!selectedRecord) return;
    setPending("archiving");
    const result = await archiveWorkflowRAGSnapshot(
      config,
      applicationId,
      selectedRecord.snapshotId,
      selectedRecord.snapshotVersion,
    );
    setPending("");
    setOperation(result);
    setShowArchiveConfirm(false);
    if (!result.record) return;
    setSelectedRecord(result.record);
    setEditor(editorFromRecord(result.record));
    setFilter("archived");
    const refreshed = await refreshCollections();
    setSelectedResource(refreshed.archived.find((resource) => resource.snapshotId === result.record?.snapshotId) ?? null);
  };

  const loadMore = async () => {
    const cursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
    if (!cursor) return;
    setPending("listing");
    const result = await listWorkflowRAGSnapshots(config, applicationId, filter, cursor);
    setPending("");
    setCollection((current) => mergePage(current, filter, result.records, result.nextCursor, result.failureCode, result.summary));
  };

  if (config.mode === "offline") {
    return <BoundaryPanel status="offline" summary="RAG 知识快照保持 offline；本面板发送 0 个请求，也不会模拟写入成功。" />;
  }
  if (!canRead || !applicationId.trim()) {
    return <BoundaryPanel status="scope denied" summary="缺少 application scope 或 workflow_rag_snapshots:read；本面板发送 0 个请求。" />;
  }

  const currentCursor = filter === "active" ? collection.activeCursor : collection.archivedCursor;
  const writeDisabled = pending !== "" || !canWrite;
  const archiveDisabled = pending !== "" || !canArchive || selectedRecord?.lifecycleState !== "active";

  return (
    <section className="workflow-rag-snapshot-panel" id="workflow-rag-snapshot-panel" aria-labelledby="workflow-rag-snapshot-title">
      <div className="section-heading compact-heading">
        <div>
          <p className="eyebrow">Workflow RAG · Application knowledge</p>
          <h4 id="workflow-rag-snapshot-title">知识快照与精确版本</h4>
        </div>
        <span className={`status-badge ${collection.failureCode ? "bad" : "good"}`}>{pending || (collection.failureCode ? "failed" : "ready")}</span>
      </div>

      <div className="workflow-rag-scope-grid">
        <article><span>Application</span><strong>{applicationName || applicationId}</strong><code>{applicationId}</code></article>
        <article><span>Repository scope</span><strong>{config.workspaceId}</strong><code>{config.tenantRef}</code></article>
        <article><span>Profile</span><strong>lexical-ngram-dev.v1</strong><small>精确版本可绑定到独立 retrieval execution。</small></article>
        <article><span>Write boundary</span><strong>{canWrite ? "create / version enabled" : "read-only"}</strong><small>归档使用独立 archive scope。</small></article>
      </div>

      {!applicationActive ? <p className="workflow-rag-boundary-note">当前应用已归档；知识快照正文与历史仍可精确读取，但创建、版本化和归档入口均关闭。</p> : null}
      {collection.failureCode ? <p className="workflow-rag-failure" role="alert"><code>{collection.failureCode}</code> · {collection.summary}</p> : null}

      <div className="workflow-rag-toolbar">
        <div className="workflow-rag-filter" aria-label="知识快照生命周期筛选">
          {(["active", "archived"] as const).map((state) => (
            <button key={state} type="button" className={filter === state ? "selected" : ""} onClick={() => setFilter(state)}>{state}</button>
          ))}
        </div>
        <button type="button" disabled={writeDisabled} onClick={() => { setShowCreate(true); setSelectedResource(null); setSelectedRecord(null); setEditor(emptyEditor()); setOperation(null); }}>
          新建知识快照
        </button>
      </div>

      <div className="workflow-rag-layout">
        <div className="workflow-rag-list" aria-label={`${filter} knowledge snapshots`}>
          {visibleResources.map((resource) => (
            <button key={resource.snapshotId} type="button" className={selectedResource?.snapshotId === resource.snapshotId ? "selected" : ""} onClick={() => void selectResource(resource)}>
              <span><strong>{resource.displayName}</strong><code>{resource.latestRAGRef}</code></span>
              <span><small>{resource.fragmentCount} fragments</small><small>{resource.totalContentBytes} bytes</small></span>
            </button>
          ))}
          {!visibleResources.length && pending !== "listing" ? <p>当前作用域没有 {filter} 知识快照。</p> : null}
          {currentCursor ? <button type="button" disabled={pending !== ""} onClick={() => void loadMore()}>加载下一页</button> : null}
        </div>

        <div className="workflow-rag-editor">
          {showCreate || selectedRecord ? (
            <SnapshotEditorFields editor={editor} immutableKey={Boolean(selectedRecord)} disabled={pending !== "" || (!showCreate && !canWrite)} onChange={setEditor} />
          ) : (
            <article className="workflow-rag-empty"><strong>选择精确快照版本</strong><p>列表只含 metadata；选择后才以 read scope 拉取当前明确版本的正文。</p></article>
          )}

          {showCreate ? (
            <div className="workflow-rag-actions"><button type="button" disabled={writeDisabled} onClick={() => void submitCreate()}>{pending === "creating" ? "创建中…" : "创建 v1"}</button><button type="button" disabled={pending !== ""} onClick={() => setShowCreate(false)}>取消</button></div>
          ) : selectedRecord ? (
            <>
              <SnapshotRecordEvidence record={selectedRecord} />
              {selectedRecord.lifecycleState === "active" ? (
                <div className="workflow-rag-actions">
                  <button type="button" disabled={writeDisabled} onClick={() => void submitVersion()}>{pending === "versioning" ? "写入中…" : `完整替换并创建 v${selectedRecord.snapshotVersion + 1}`}</button>
                  <button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => setShowArchiveConfirm(true)}>归档快照</button>
                </div>
              ) : null}
              {showArchiveConfirm ? (
                <div className="workflow-rag-archive-confirm" role="alert"><p>归档后禁止创建新版本；现有版本仍保持精确可读。</p><button type="button" className="danger-action" disabled={archiveDisabled} onClick={() => void submitArchive()}>确认归档</button><button type="button" disabled={pending !== ""} onClick={() => setShowArchiveConfirm(false)}>取消</button></div>
              ) : null}
            </>
          ) : null}

          {operation ? <OperationEvidence operation={operation} /> : null}
        </div>
      </div>
    </section>
  );
}

function SnapshotEditorFields({ editor, immutableKey, disabled, onChange }: { editor: SnapshotEditor; immutableKey: boolean; disabled: boolean; onChange: (value: SnapshotEditor) => void }) {
  return (
    <div className="workflow-rag-fields">
      <label><span>Snapshot key</span><input value={editor.snapshotKey} disabled={disabled || immutableKey} maxLength={48} onChange={(event) => onChange({ ...editor, snapshotKey: event.currentTarget.value })} /></label>
      <label><span>Display name</span><input value={editor.displayName} disabled={disabled} maxLength={120} onChange={(event) => onChange({ ...editor, displayName: event.currentTarget.value })} /></label>
      <label><span>Classification</span><select value={editor.contentClassification} disabled={disabled} onChange={(event) => onChange({ ...editor, contentClassification: event.currentTarget.value as WorkflowRAGContentClassification })}><option value="workspace_internal">workspace_internal</option><option value="public">public</option></select></label>
      <label className="workflow-rag-fragments-field"><span>Replacement fragments（camelCase JSON array）</span><textarea value={editor.fragmentsJSON} disabled={disabled} spellCheck={false} onChange={(event) => onChange({ ...editor, fragmentsJSON: event.currentTarget.value })} /><small>版本化必须提交完整 replacement；单 fragment 正文最多 8 KiB，总量最多 1 MiB。</small></label>
    </div>
  );
}

function SnapshotRecordEvidence({ record }: { record: WorkflowRAGSnapshotRecord }) {
  return (
    <article className="workflow-rag-record">
      <div><strong>{record.ragRef}</strong><span className="status-badge neutral">{record.lifecycleState}</span></div>
      <dl><div><dt>Digest</dt><dd>{record.snapshotDigest}</dd></div><div><dt>Profile</dt><dd>{record.profileRef}</dd></div><div><dt>Fragments</dt><dd>{record.fragmentCount}</dd></div><div><dt>Content bytes</dt><dd>{record.totalContentBytes}</dd></div><div><dt>Request</dt><dd>{record.requestId}</dd></div><div><dt>Audit</dt><dd>{record.auditRef}</dd></div></dl>
    </article>
  );
}

function OperationEvidence({ operation }: { operation: WorkflowRAGSnapshotOperationResult }) {
  return <article className={`workflow-rag-operation ${operation.status === "failed" || operation.status === "version_conflict" ? "failed" : ""}`} aria-live="polite"><strong>{operation.status}</strong><p>{operation.summary}</p>{operation.failureCode ? <code>{operation.failureCode}</code> : null}{operation.status === "version_conflict" ? <small>Current: v{operation.currentLatestVersion} · {operation.currentLifecycleState}</small> : null}</article>;
}

function BoundaryPanel({ status, summary }: { status: string; summary: string }) {
  return <section className="workflow-rag-snapshot-panel offline" aria-label="Workflow RAG knowledge snapshot"><div className="section-heading compact-heading"><div><p className="eyebrow">Workflow RAG · Application knowledge</p><h4>知识快照未启用</h4></div><span className="status-badge neutral">{status}</span></div><p>{summary}</p></section>;
}

function parseEditor(editor: SnapshotEditor): WorkflowRAGSnapshotWriteInput | string {
  let fragments: unknown;
  try {
    fragments = JSON.parse(editor.fragmentsJSON);
  } catch {
    return "workflow_rag_fragment_invalid";
  }
  if (!isFragmentInputs(fragments)) return "workflow_rag_fragment_invalid";
  const input: WorkflowRAGSnapshotWriteInput = { snapshotKey: editor.snapshotKey, displayName: editor.displayName, contentClassification: editor.contentClassification, fragments };
  return validateWorkflowRAGSnapshotWriteInput(input) || input;
}

function editorFromRecord(record: WorkflowRAGSnapshotRecord): SnapshotEditor {
  return { snapshotKey: record.snapshotKey, displayName: record.displayName, contentClassification: record.contentClassification, fragmentsJSON: JSON.stringify(record.fragments.map((fragment) => ({ fragmentRef: fragment.fragmentRef, sourceType: fragment.sourceType, sourceRef: fragment.sourceRef, pageSlug: fragment.pageSlug, title: fragment.title, isOfficial: fragment.isOfficial, content: fragment.content })), null, 2) };
}

function emptyEditor(): SnapshotEditor {
  return { snapshotKey: "", displayName: "", contentClassification: "workspace_internal", fragmentsJSON: JSON.stringify([EMPTY_FRAGMENT], null, 2) };
}

function emptyCollection(): SnapshotCollection {
  return { active: [], archived: [], activeCursor: "", archivedCursor: "", failureCode: "", summary: "" };
}

function collectionFromResults(active: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>, archived: Awaited<ReturnType<typeof listWorkflowRAGSnapshots>>): SnapshotCollection {
  return { active: active.records, archived: archived.records, activeCursor: active.nextCursor, archivedCursor: archived.nextCursor, failureCode: active.failureCode || archived.failureCode, summary: active.failureCode ? active.summary : archived.failureCode ? archived.summary : `${active.records.length} active / ${archived.records.length} archived` };
}

function mergePage(current: SnapshotCollection, lifecycle: WorkflowRAGSnapshotLifecycle, records: WorkflowRAGSnapshotResource[], cursor: string, failureCode: string, summary: string): SnapshotCollection {
  const merged = mergeResources(lifecycle === "active" ? current.active : current.archived, records);
  return lifecycle === "active" ? { ...current, active: merged, activeCursor: cursor, failureCode, summary } : { ...current, archived: merged, archivedCursor: cursor, failureCode, summary };
}

function mergeResources(current: WorkflowRAGSnapshotResource[], incoming: WorkflowRAGSnapshotResource[]): WorkflowRAGSnapshotResource[] {
  return [...new Map([...current, ...incoming].map((resource) => [resource.snapshotId, resource])).values()].sort((left, right) => left.snapshotKey.localeCompare(right.snapshotKey));
}

function isFragmentInputs(value: unknown): value is WorkflowRAGFragmentInput[] {
  if (!Array.isArray(value)) return false;
  const expectedKeys = new Set(["fragmentRef", "sourceType", "sourceRef", "pageSlug", "title", "isOfficial", "content"]);
  return value.every((fragment) => {
    if (typeof fragment !== "object" || fragment === null || Array.isArray(fragment)) return false;
    const record = fragment as Record<string, unknown>;
    return Object.keys(record).length === expectedKeys.size && Object.keys(record).every((key) => expectedKeys.has(key)) &&
      typeof record.fragmentRef === "string" && typeof record.sourceType === "string" && typeof record.sourceRef === "string" &&
      typeof record.pageSlug === "string" && typeof record.title === "string" && typeof record.isOfficial === "boolean" && typeof record.content === "string";
  });
}

function localFailure(failureCode: string): WorkflowRAGSnapshotOperationResult {
  return { status: "failed", record: null, failureCode, currentLatestVersion: 0, currentLifecycleState: "", summary: "知识快照输入在请求发送前被拒绝。" };
}
